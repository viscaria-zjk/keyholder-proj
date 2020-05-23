// 作者：李涵
// 内建数据库系统访问API

package storage

import (
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "keyholder/utils"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)
// 最大重定向次数
const maxRedirect = 21
// 访问数据库的参数
type dbAccessArgs struct {
    Protocol    string
    Host        string
    Port        uint16
    Prefix      string
    Insecure    bool
    CACert      string
    Credentials string
    MaxRedirect int
}
// 执行回复
type execResponse struct {
    Results []*execResult `json:"results,omitempty"`
    Error   string        `json:"error,omitempty"`
    Time    float64       `json:"time,omitempty"`
}
// 查询结果
type queryResult struct {
    Results []*rows `json:"results"`
    Error   string  `json:"error,omitempty"`
    Time    float64 `json:"time"`
}
// 执行结果
type execResult struct {
    LastInsertID int     `json:"last_insert_id,omitempty"`
    RowsAffected int     `json:"rows_affected,omitempty"`
    Time         float64 `json:"time,omitempty"`
    Error        string  `json:"error,omitempty"`
}
// 查询结果列
type rows struct {
    Columns []string        `json:"columns"`
    Types   []string        `json:"types"`
    Values  [][]interface{} `json:"values"`
    Time    float64         `json:"time"`
    Error   string          `json:"error,omitempty"`
}

// db访问对象
type DBAccess struct {
    httpClient    *http.Client
    argv          *dbAccessArgs
    isTimerNeeded bool
    execSpan      float64
}

// 全局变量：数据库系统在OS执行的进程
var dbms *exec.Cmd

// 根据访问数据库的参数，得到HTTP用户对象
func getHTTPClient(argv *dbAccessArgs) (*http.Client, error) {
    var rootCAs *x509.CertPool

    if argv.CACert != "" {
        pemCerts, err := ioutil.ReadFile(argv.CACert)
        if err != nil {
            return nil, err
        }
        rootCAs = x509.NewCertPool()
        ok := rootCAs.AppendCertsFromPEM(pemCerts)
        if !ok {
            return nil, fmt.Errorf("failed to parse root CA certificate(s)")
        }
    }

    client := http.Client{Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: argv.Insecure, RootCAs: rootCAs},
        Proxy:           http.ProxyFromEnvironment,
    }}

    // Explicitly handle redirects.
    client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }

    return &client, nil
}
// 解读Query或是Execute回复的json进入特定对象
func parseResponse(response *[]byte, ret interface{}) error {
    return json.Unmarshal(*response, ret)
}
// 验证查询结果
func (r *rows) validateQueryResponse() error {
    if r.Error != "" {
        return fmt.Errorf(r.Error)
    }
    if r.Columns == nil || r.Types == nil {
        return fmt.Errorf("unexpected result")
    }
    return nil
}
// 制作执行SQL使用的json格式
func makeJSONBody(line string) string {
    data, err := json.Marshal([]string{line})
    if err != nil {
        return ""
    }
    return string(data)
}

// 查询（SELECT）
func (dba *DBAccess) Query(query string) (rowsDerived *rows, err error) {
    queryStr := url.Values{}
    queryStr.Set("q", query)
    if dba.isTimerNeeded {
        queryStr.Set("timings", "")
    }
    // 构造http地址以访问RDBMS
    u := url.URL {
        Scheme:   dba.argv.Protocol,
        Host:     fmt.Sprintf("%s:%d", dba.argv.Host, dba.argv.Port),
        Path:     fmt.Sprintf("%sdb/query", dba.argv.Prefix),
        RawQuery: queryStr.Encode(),
    }
    urlStr := u.String()
    nRedirect := 0
    for {
        // 新建GET请求
        req, err := http.NewRequest("GET", urlStr, nil)
        if err != nil {
            return nil, err
        }
        // 若提供了用户名/密码
        if dba.argv.Credentials != "" {
            creds := strings.Split(dba.argv.Credentials, ":")
            if len(creds) != 2 {
                return nil, fmt.Errorf("invalid Basic Auth credentials format")
            }
            req.SetBasicAuth(creds[0], creds[1])
        }
        // 开始HTTP请求及接收回应
        resp, err := dba.httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        // 读取回应
        response, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }
        _ = resp.Body.Close()
        // 若访问凭据不被允许
        if resp.StatusCode == http.StatusUnauthorized {
            return nil, fmt.Errorf("unauthorized")
        }
        // 若被重定向
        if resp.StatusCode == http.StatusMovedPermanently {
            nRedirect++
            if nRedirect > dba.argv.MaxRedirect {
                return nil, fmt.Errorf("maximum leader redirect limit exceeded")
            }
            urlStr = resp.Header["Location"][0]
            continue
        }
        // 其他不成功的原因
        if resp.StatusCode != http.StatusOK {
            return nil, fmt.Errorf("server responded with: %s", resp.Status)
        }
        // 若成功，则获取回应并解析结论
        ret := &queryResult{}
        if err = parseResponse(&response, &ret); err != nil {
            return nil, err
        }
        if ret.Error != "" {
            return nil, fmt.Errorf(ret.Error)
        }
        if len(ret.Results) != 1 {
            return nil, fmt.Errorf("unexpected results length: %d", len(ret.Results))
        }
        result := ret.Results[0]
        if err := result.validateQueryResponse(); err != nil {
            return nil, err
        }

        if dba.isTimerNeeded {
            fmt.Printf("Run Time: %f seconds\n", result.Time)
        }
        return result, nil
    }
}

// 执行（CREATE，DROP，DELETE，ALTER，UPDATE，INSERT等）
func (dba *DBAccess) Execute(statement string) (rowsAffected int, err error) {
    // 构造HTTP对象
    queryStr := url.Values{}
    if dba.isTimerNeeded {
        queryStr.Set("timings", "")
    }
    u := url.URL{
        Scheme: dba.argv.Protocol,
        Host:   fmt.Sprintf("%s:%d", dba.argv.Host, dba.argv.Port),
        Path:   fmt.Sprintf("%sdb/execute", dba.argv.Prefix),
    }
    urlStr := u.String()
    requestData := strings.NewReader(makeJSONBody(statement))
    nRedirect := 0
    for {
        if _, err := requestData.Seek(0, io.SeekStart); err != nil {
            return 0, err
        }
        // 利用HTTP POST发送请求
        req, err := http.NewRequest("POST", urlStr, requestData)
        if err != nil {
            return 0, err
        }
        if dba.argv.Credentials != "" {
            creds := strings.Split(dba.argv.Credentials, ":")
            if len(creds) != 2 {
                return 0, fmt.Errorf("invalid Basic Auth credentials format")
            }
            req.SetBasicAuth(creds[0], creds[1])
        }
        // 发送请求
        resp, err := dba.httpClient.Do(req)
        if err != nil {
            return 0, err
        }
        response, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return 0, err
        }
        _ = resp.Body.Close()
        if resp.StatusCode == http.StatusUnauthorized {
            return 0, fmt.Errorf("unauthorized")
        }
        if resp.StatusCode == http.StatusMovedPermanently {
            nRedirect++
            if nRedirect > dba.argv.MaxRedirect {
                return 0, fmt.Errorf("maximum leader redirect limit exceeded")
            }
            urlStr = resp.Header["Location"][0]
            continue
        }
        if resp.StatusCode != http.StatusOK {
            return 0, fmt.Errorf("server responded with: %s", resp.Status)
        }
        // 获取response
        ret := &execResponse{}
        if err := parseResponse(&response, &ret); err != nil {
            return 0, err
        }
        if ret.Error != "" {
            return 0, fmt.Errorf(ret.Error)
        }
        if len(ret.Results) != 1 {
            return 0, fmt.Errorf("unexpected results length: %d", len(ret.Results))
        }
        // 执行时候出现错误（无表等等）
        result := ret.Results[0]
        if result.Error != "" {
            return 0, fmt.Errorf("execute failed because %s\n", result.Error)
        }
        if dba.isTimerNeeded {
            dba.execSpan = result.Time
        }
        return result.RowsAffected, nil
    }
}

// 新建一个数据库访问实例
func openDBA(args *dbAccessArgs, needTimer bool) (newDBA *DBAccess) {
    // 获得HTTP客户端实例
    client, err := getHTTPClient(args)
    if err != nil {	// 创建错误，返回nil
        return nil
    }
    // 新建DBAccess对象
    newDBA = new(DBAccess)
    newDBA.argv = args
    newDBA.httpClient = client
    newDBA.isTimerNeeded = needTimer
    return
}
// 启动数据库系统
func openDBMS(dbmsPath string, atPort int, maxRetryTimes int) (dbms *exec.Cmd, err error) {
    // 查询待联入的端口是否有在使用
    isInUse := utils.IsPortInUse("localhost", atPort, "tcp")
    if isInUse {
        return nil, fmt.Errorf("the given port (%d) is in use so dbms could not start", atPort)
    }
    // 打开RDBMS
    dbms = exec.Command(dbmsPath, filepath.Join(filepath.Dir(dbmsPath), "/nodes/node.1"))
    err = dbms.Start()
    if err != nil {

    }
    fmt.Print("Starting DBMS..")
    // 不断查询目标端口是否有在使用，直到发现DBMS启动成功为止
    var retryTimes = 0
    for retryTimes < maxRetryTimes {
        fmt.Print("."); retryTimes++
        if utils.IsPortInUse("127.0.0.1", atPort, "tcp") {
            fmt.Println("OK")
            return dbms, nil
        }
        time.Sleep(700 * time.Millisecond)
    }
    fmt.Println("ERROR")
    return nil, fmt.Errorf("RDBMS does not exist or is broken")
}
// 关闭数据库系统
func closeDBMS(dbms *exec.Cmd) {
    // TODO - 清洁操作
    // 杀死数据库进程
    _ = dbms.Process.Kill()
}

// 建立初始表
func (dba *DBAccess) initDB() error {
    _, err := dba.Execute(`
CREATE TABLE IF NOT EXISTS KEYS (
    KEY_NUM CHAR(8) PRIMARY KEY,
    KEY_CAPACITY INT NOT NULL,
    KEY_RESIDUE INT NOT NULL,
    USERNAME TEXT NOT NULL,
    PASSWORD TEXT,
    EXPIRE_DATE DATETIME,
    LAST_LOGIN DATETIME,
    
    CHECK (KEY_RESIDUE <= KEY_CAPACITY AND KEY_CAPACITY > 0 AND KEY_RESIDUE >= 0),
    CHECK (LENGTH(KEY_NUM) = 8)
);

CREATE TABLE IF NOT EXISTS SESSIONS (
    SESSION_ID TEXT PRIMARY KEY,
    KEY_NUM CHAR(8) NOT NULL,
    LOGIN_TIMESTAMP INTEGER NOT NULL,
    LAST_UPDATED INTEGER NOT NULL,
    CLIENT_SERVER_NUM INTEGER,

    FOREIGN KEY (KEY_NUM) REFERENCES KEYS(KEY_NUM)
);

CREATE TRIGGER IF NOT EXISTS TRIG_LOGIN AFTER INSERT
ON SESSIONS
BEGIN
    -- UPDATE THE KEY RESIDUE
    UPDATE KEYS SET KEY_RESIDUE = KEY_RESIDUE - 1 WHERE KEY_NUM = new.KEY_NUM;
    -- UPDATE LAST LOGIN DATETIME
    UPDATE KEYS SET LAST_LOGIN = DATETIME('now') WHERE KEY_NUM = new.KEY_NUM;
END;

CREATE TRIGGER IF NOT EXISTS TRIG_LOGOUT AFTER DELETE
ON SESSIONS
FOR EACH ROW
BEGIN
    -- UPDATE THE KEY RESIDUE
    UPDATE KEYS SET KEY_RESIDUE = KEY_RESIDUE + 1 WHERE KEY_NUM = old.KEY_NUM;
END;
`)
    return err
}

// 启动数据库并返回一个新数据库实例
func DBStartup() *DBAccess {
    // 打开RDBMS（不必了）
    var err error
    // dbms, err = openDBMS(filepath.Join(filepath.Dir(os.Args[0]), "/res/rqlited"), 4001, 12)
    //if err != nil {
    //    if dbms != nil {
    //        closeDBMS(dbms)
    //    }
    //    utils.ReportError("on server startup: DBMS failed to start:", err)
    //    return nil
    //}
    dbArgs := new(dbAccessArgs)
    dbArgs.Protocol = "http"
    dbArgs.Host = "127.0.0.1"       // 注意：这是数据库访问，需要在本机运行，必是127.0.0.1
    dbArgs.Port = 4001
    dbArgs.Prefix = "/"
    dbArgs.Insecure = false
    dbArgs.CACert = ""
    dbArgs.Credentials = ""
    dbArgs.MaxRedirect = maxRedirect
    // 新建DBAccess
    dba := openDBA(dbArgs, false)
    if dba == nil {
        closeDBMS(dbms)
        utils.ReportError("on server startup: DBMS failed to start")
        return nil
    }
    // 初始化数据表（如果需要）
    err = dba.initDB()
    if err != nil {
        closeDBMS(dbms)
        utils.ReportError("on server startup: DBMS failed to initialise:", err)
        return nil
    }
    return dba
}

// 关闭数据库
func DBEnd() {
    closeDBMS(dbms)
    dbms = nil
}

func doDB() {
    // 打开RDBMS
    dbms, err := openDBMS(filepath.Join(filepath.Dir(os.Args[0]), "/res/rqlited"), 4001, 12)
    if err != nil {
        if dbms != nil { closeDBMS(dbms) }
        os.Exit(-1)
    }
    dbArgs := new(dbAccessArgs)
    dbArgs.Protocol = "http"
    dbArgs.Host = "127.0.0.1"
    dbArgs.Port = 4001
    dbArgs.Prefix = "/"
    dbArgs.Insecure = false
    dbArgs.CACert = ""
    dbArgs.Credentials = ""
    dbArgs.MaxRedirect = maxRedirect
    // 新建DBAccess
    dba := openDBA(dbArgs, false)
    if dba == nil {
        closeDBMS(dbms)
        os.Exit(-1)
    }
    // 初始化数据表（如果需要）
    err = dba.initDB()
    if err != nil {
        closeDBMS(dbms)
        os.Exit(-1)
    }

    // 新增记录
    info, err := NewKeyInfo("8p4r5Fuj",
        2, 2, "HANLI", "LIHAN1313",
        utils.GerUniversalDate("2099-01-01"))
    if err != nil {
        closeDBMS(dbms)
        panic(err)
    }
    err = dba.AddKeyInfo(info)
    if err != nil {
        closeDBMS(dbms)
        panic(err)
    }
    // 查询
    info, err = dba.GetKeyInfo(info.KeyNum)
    if err != nil {
        closeDBMS(dbms)
        panic(err)
    }
    // 表达
    if info != nil {
        fmt.Println(info.KeyNum, info.KeyCapacity, info.KeyResidue, info.Username,
            info.Password, utils.RenderTime(info.ExpDate), utils.RenderTime(info.LastLogin))
    }
    // 删除
    err = dba.DelKeyInfo(info.KeyNum)
    if err != nil {
        closeDBMS(dbms)
        panic(err)
    }

    closeDBMS(dbms)
}
