package server

import (
    "bytes"
    "fmt"
    "keyholder/ciph"
    "keyholder/proto"
    "keyholder/storage"
    "keyholder/utils"
    "net"
    "time"
)

// 一个服务器实例
type Server struct {
    isRunning        bool              // 标志是否正在运行
    ListenAddr       string            // 监听的地址
    ListenPort       int               // 监听的端口
    MaxBuffSize      int               // 最大缓冲区大小
    MaxConcurrent    int               // 最大并发
    Concurrent       int               // 目前并发数
    ConnTimeout      int               // 建立连接超时时间
    HeartbeatTimeout int               // 心跳超时时间
    DB               *storage.DBAccess // 数据库访问实例
    MainConn         *net.UDPConn      // UDP主监听套接字
}

// 一个会话实例
type Session struct {
    isRunning   bool                // 标志是否正运行
    conn        *net.UDPConn        // 本次会话的随机UDP监听器
    port        uint32              // 本次会话的随机UDP监听器的端口
    client      *net.UDPAddr        // 客户端地址及其端口
    sv          *Server             // 本会话对应的服务器实例
}

// 处理一个会话
func (ss *Session)Run() {
    defer ss.conn.Close()
    defer func() { ss.sv.Concurrent-- }()
    defer utils.ReportWarning("closing port", ss.port)
    // 新建本会话缓存
    buffWrite := new(bytes.Buffer)
    buffRead := make([]byte, ss.sv.MaxBuffSize)
    //utils.ReportInfo("urging", ss.client.String(), "to connect at", ss.port)
    // 1. 给客服发回本次会话的地址
    buffWrite.WriteString("CONN ")
    buffWrite.Write(utils.Uint32ToBytes(ss.port))
    _, err := ss.sv.MainConn.WriteToUDP(buffWrite.Bytes(), ss.client)
    if err != nil {
        utils.ReportError("Write to UDP failed because:", err)
        return
    }
    // 2. 收取客服的"OKAY"信号，超时就关闭所有连接
    l := 0
    timer := time.AfterFunc(time.Millisecond * time.Duration(ss.sv.ConnTimeout), func() {
        _ = ss.conn.Close()
    })
    var recvAddr *net.UDPAddr
    for {
        l, recvAddr, err = ss.conn.ReadFromUDP(buffRead)
        if err != nil {
            return
        } // 超时退出
        if string(buffRead[:4]) != "OKAY" {
            utils.ReportWarning("received not OKAY:", string(buffRead[:4]))
            continue
        }  // 非法信号，重读
        timer.Stop()
        if !recvAddr.IP.Equal(ss.client.IP) {
            utils.ReportWarning("received from weird IP:", recvAddr.IP)
        }
        break
    }
    // NAT会篡改端口！
    ss.client.Port = recvAddr.Port
    utils.ReportInfo("host from", ss.client.String(), "is at", ss.port)
    defer utils.ReportWarning("host from", ss.client.String(), "leaved", ss.port)
    // 3. 生成RSA钥匙
    pubK, priK, err := ciph.GenRSAKeys(1024)
    if err != nil {
        utils.ReportError("RSA Keys generator threw an error:", err)
        return
    }
    // 4. 发送RSA公钥给客服
    buffWrite.Reset()
    buffWrite.WriteString("PUBK ")
    buffWrite.Write(utils.Uint32ToBytes(uint32(len(pubK))))
    buffWrite.WriteString(" ")
    buffWrite.Write(pubK)
    _, err = ss.conn.WriteToUDP(buffWrite.Bytes(), ss.client)
    utils.ReportInfo("RSA public was sent to", ss.client.String())
    // 5. 只接收"AESK"回应。
    var aesKey []byte
    for {
        timer.Reset(time.Millisecond * time.Duration(ss.sv.ConnTimeout))
        l, recvAddr, err = ss.conn.ReadFromUDP(buffRead)
        if err != nil {
            return
        } // 超时退出
        aesKey = ciph.RSADecrypt(buffRead[:l], priK)
        if aesKey == nil { // 解密错误，重读
            utils.ReportWarning("RSA decode error")
            continue
        }
        if utils.ToString(aesKey[0:4]) != "AESK" {
            utils.ReportWarning("want \"AESK\", received", utils.ToString(aesKey[0:4]))
        }
        // 拿到密码
        aesKey = aesKey[5:21]
        timer.Stop()
        break
    }
    // 6. 回复客服"OKAY"
    encryptedOKAY := ciph.AESEncrypt([]byte("OKAY"), aesKey)
    _, err = ss.conn.WriteToUDP(encryptedOKAY, ss.client)
    // 7. 交换密钥结束，开始接收客服请求。(ss.sv.HeartbeatTimeout)秒内必须要有一条回应，否则关闭连接
    utils.ReportInfo("established conn with", ss.client.String(), "at", ss.port)
    for {
        // 重设长连接断开计时器
        timer.Reset(time.Second * time.Duration(ss.sv.HeartbeatTimeout))
        l, recvAddr, err = ss.conn.ReadFromUDP(buffRead)
        if err != nil {
            return
        } // 超时退出
        decoded := ciph.AESDecrypt(buffRead[:l], aesKey)
        if decoded == nil { // 解密错误，退出
            continue
        }
        if len(decoded) < 4 {
            return
        }
        timer.Stop()    // 停止秒表
        // 检测命令类型
        switch string(decoded[:4]) {
        case "APPL" :  // 申请
            sesID, ret := proto.ApplyKey(ss.sv.DB, decoded, ss.port)
            utils.ReportInfo("cs wants to apply a key, with status code:", ret)
            utils.ReportInfo("the sesID was", sesID)
            buffWrite.Reset()
            if ret == 0 {   // 申请成功，返回会话编号
                buffWrite.WriteString(fmt.Sprintf("OKAY "))
                buffWrite.Write(utils.Uint64ToBytes(sesID))
            } else {
                buffWrite.WriteString(fmt.Sprintf("NOOK %03d", ret))
            }
            _, err = ss.conn.WriteToUDP(ciph.AESEncrypt(buffWrite.Bytes(), aesKey), ss.client)
            break
        case "RETU" :  // 归还
            ret := proto.ReturnKey(ss.sv.DB, decoded, ss.port)
            utils.ReportInfo("cs wants to return a key, with status code:", ret)
            buffWrite.Reset()
            if ret == 0 {   // 归还成功
                buffWrite.WriteString(fmt.Sprintf("OKAY"))
            } else {
                buffWrite.WriteString(fmt.Sprintf("NOOK %03d", ret))
            }
            _, err = ss.conn.WriteToUDP(ciph.AESEncrypt(buffWrite.Bytes(), aesKey), ss.client)
            break
        case "NORM" :  // 正常心跳
        	utils.ReportInfo("cs", ss.port, "heart beat received")
            _, err = ss.conn.WriteToUDP(encryptedOKAY, ss.client)
            break
        case "GBYE" :  // 客户端关闭
            proto.CSOver(ss.sv.DB, ss.port)
            _, err = ss.conn.WriteToUDP(encryptedOKAY, ss.client)
            return     // 切断连接
        } // 检测命令类型
    } // for
}

// 新建一服务器实例
func New(addr string, port int) (*Server, error) {
    // 查询端口是否被占用
    if utils.IsPortInUse(addr, port, "udp") {
        return nil, fmt.Errorf("port %d is in use", port)
    }
    // 若没有被使用就返回一新实例
    return &Server {
        ListenAddr:       addr,
        ListenPort:       port,
        MaxBuffSize:      2048,
        MaxConcurrent:    1000,
        HeartbeatTimeout: 300, // 心跳最大间隔：300秒
        Concurrent:       0,
        ConnTimeout:      10000, // 连接建立超时时间：10秒
    }, nil
}

// 强制结束一服务器实例
func (sv *Server) Close() {
    // storage.DBEnd()
    sv.isRunning = false
    sv.DB = nil
}

// 运行一服务器实例
func (sv *Server) Run() error {
    if sv.isRunning {
        return fmt.Errorf("server is running")
    }
    // 打开RDBMS
    sv.DB = storage.DBStartup()
    if sv.DB == nil {   // dbms启动失败
        return fmt.Errorf("DBMS was not start")
    }
    // 解析UDP地址
    fullAddr := fmt.Sprintf("%s:%d", sv.ListenAddr, sv.ListenPort)
    ipFullAddr, err := net.ResolveUDPAddr("udp", fullAddr)
    if err != nil {
        return fmt.Errorf("address resolve not passed: %s", err)
    }

    // 新建UDP监听套接字（服务器主线程监听套接字）
    svMainConn, err := net.ListenUDP("udp", ipFullAddr)
    if err != nil {
        return fmt.Errorf("address resolve not passed: %s", err)
    }
    defer svMainConn.Close()
    sv.MainConn = svMainConn

    // 开始监听发来的NEWS请求
    utils.ReportInfo("Listening at", svMainConn.LocalAddr().String())
    buff := make([]byte, sv.MaxBuffSize); sv.isRunning = true
    for {
        // 收取报文
        l, clAddr, err := svMainConn.ReadFromUDP(buff)
        if err != nil { // 抛出异常：收讯不能成功
            return fmt.Errorf("read from udp failed: %s", err)
        }
        // 用新goroutine解析udp报文
        if sv.Concurrent > sv.MaxConcurrent || l < 4 || string(buff[:4]) != "NEWS" {
            // 报文太短、不合法报文或是并发超出，扔掉
            continue
        }
        go func(l int, clAddr *net.UDPAddr) {
            // 信号有效。查看有无提供端口号，若无则申请一个udp端口以供连接
            var csConn *net.UDPConn
            if l == 9 {
                // 提供了端口号，是因为之前服务器崩溃了，而客服想恢复之前的状态
                oldCSNumber := utils.BytesToInt32(buff[5:9])
                newAddrStr := fmt.Sprintf("%s:%d", sv.ListenAddr, oldCSNumber)
                newAddr, _ := net.ResolveUDPAddr("udp", newAddrStr)
                csConn, err = net.ListenUDP("udp", newAddr)
                if err != nil { // 这个端口号竟然被占用了。。。
                    csConn = nil
                    for csConn == nil {
                        csConn = utils.RandomUDPConn(newAddr)
                    }
                    // 在数据库中更新这个旧端口号的记录（旧客服号）至新端口号
                    proto.UpdateCSNumber(sv.DB, int(oldCSNumber), csConn.LocalAddr().(*net.UDPAddr).Port)
                }
            } else {
                // 未提供端口号，则提供一个随机高位端口号
                newAddrStr := fmt.Sprintf("%s:0", sv.ListenAddr)
                newAddr, _ := net.ResolveUDPAddr("udp", newAddrStr)
                csConn = utils.RandomUDPConn(newAddr)
                for csConn == nil {
                    csConn = utils.RandomUDPConn(newAddr)
                }
            }
            // 新建协程处理该会话
            newSes := &Session {
                isRunning:  true,
                conn:       csConn,
                sv:         sv,
                client:     clAddr,
                port:       uint32(csConn.LocalAddr().(*net.UDPAddr).Port),
            }
            // 并发数+1
            sv.Concurrent ++
            go newSes.Run()
        }(l, clAddr)
    }
}