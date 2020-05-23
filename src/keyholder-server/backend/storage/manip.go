// 查询数据库

package storage

import (
    "fmt"
    "keyholder/utils"
    "time"
)

type KeyInfo struct {
    KeyNum string
    KeyCapacity int
    KeyResidue int
    Username string
    Password string
    ExpDate int64   // 结束时间戳
    LastLogin int64 // 上次登陆时间戳
}

type SessionInfo struct {
    SessionID uint64
    KeyNum string
    Username string
    Password string
    ClientServer uint32
}

func NewKeyInfo(KeyNum string, KeyCapacity int, KeyResidue int, Username string, Password string, ExpDate int64) (ki *KeyInfo, err error) {
    ki = new(KeyInfo)
    ki.KeyNum = KeyNum
    if KeyCapacity > 0 {
        ki.KeyCapacity = KeyCapacity
    } else {
        return nil, fmt.Errorf("key capacity should be positive")
    }
    if KeyResidue > 0 && KeyResidue <= KeyCapacity {
        ki.KeyResidue = KeyResidue
    } else {
        return nil, fmt.Errorf("key residue should be positive and less than capacity")
    }
    if len(Username) > 18 {
        return nil, fmt.Errorf("username length should be less than 18")
    } else {
        ki.Username = Username
    }
    if len(Password) > 18 || len(Password) < 6 {
        return nil, fmt.Errorf("password length should be between 6 and 18")
    } else {
        ki.Password = Password
    }
    ki.ExpDate = ExpDate
    // 设置上次登陆时间为现在
    ki.LastLogin = utils.GetUniversalTime(time.Now())

    return ki, nil
}

// 查询某个序列号的详情
func (dba *DBAccess) GetKeyInfo(keyNum string) (ki *KeyInfo, err error) {
    res, err := dba.Query(fmt.Sprintf(
        `SELECT KEY_NUM, KEY_CAPACITY, 
KEY_RESIDUE, USERNAME, PASSWORD, EXPIRE_DATE, 
LAST_LOGIN 
FROM KEYS WHERE KEY_NUM='%s';`, keyNum))
    // 查询过程中出错
    if err != nil {
        return nil, err
    }
    // 没有获取到该key的相关记录
    if len(res.Values) == 0 {
        return nil, nil
    }
    infoRow := res.Values[0]
    ki = new(KeyInfo)
    ki.KeyNum = infoRow[0].(string)
    ki.KeyCapacity = int(infoRow[1].(float64))
    ki.KeyResidue = int(infoRow[2].(float64))
    ki.Username = infoRow[3].(string)
    ki.Password = infoRow[4].(string)
    ki.ExpDate = utils.ParseTime(infoRow[5].(string))
    ki.LastLogin = utils.ParseTime(infoRow[6].(string))
    return ki, nil
}

// 新增一个序列号
func (dba *DBAccess) AddKeyInfo(info *KeyInfo) (err error) {
    if info == nil { return fmt.Errorf("bad keyinfo") }
    aff, err := dba.Execute(fmt.Sprintf(
        `INSERT INTO KEYS(KEY_NUM, KEY_CAPACITY, 
KEY_RESIDUE, USERNAME, PASSWORD, 
EXPIRE_DATE, LAST_LOGIN)
VALUES ('%s', %d, %d, '%s', '%s', %d, %d);`,
info.KeyNum, info.KeyCapacity, info.KeyResidue, info.Username,
info.Password, info.ExpDate, info.LastLogin))
    if err != nil {
        return err
    }
    if aff == 0 {  // 没有新增成功
        return fmt.Errorf("got an unknown error which causes no row affected")
    }
    return nil
}

// 删除一个序列号的记录
func (dba *DBAccess) DelKeyInfo(keyNum string) (err error) {
    aff, err := dba.Execute(fmt.Sprintf(
        `DELETE FROM KEYS WHERE KEY_NUM='%s'`, keyNum))
    if err != nil {
        return err
    }
    if aff == 0 {  // 没有新增成功
        return fmt.Errorf("got an unknown error which causes no row affected")
    }
    return nil
}

// 申请序列号使用（返回信息参见协议）
func (dba *DBAccess) ApplyKeyExtend(ss *SessionInfo) int {
    if ss == nil {
        utils.ReportError("unknown error when applying keys: got nil session info")
        return 4
    }
    // 查询钥匙信息
    keyInfo, err := dba.GetKeyInfo(ss.KeyNum)
    if err != nil {
        utils.ReportError("unknown error when applying keys: getting key info:", err)
        return 4
    }
    if keyInfo == nil { // 002 该序列号无效
        return 2
    }
    if keyInfo.KeyResidue == 0 { // “001”：该序列号使用人数已满
        return 1
    }

    // 查询是否已经存在了会话信息
    r, err := dba.Query(fmt.Sprintf(
        `SELECT * FROM SESSIONS WHERE SESSION_ID='%d'`, ss.SessionID))
    if err != nil {
        utils.ReportError("unknown error when applying keys: querying session:", err)
        return 4
    }
    if len(r.Values) != 0 { // “003”：该序列号下，已经有一个相同唯一标识码的用户在线
        return 3
    }

    // 可以增加使用记录
    aff, err := dba.Execute(fmt.Sprintf(`INSERT INTO SESSIONS(SESSION_ID, KEY_NUM, 
LOGIN_TIMESTAMP, LAST_UPDATED, CLIENT_SERVER_NUM) VALUES ('%d', '%s', %d, %d, %d);`,
        ss.SessionID, ss.KeyNum, time.Now().Unix(), time.Now().Unix(), ss.ClientServer))
    if err != nil || aff == 0 { // 未新增成功
        if err != nil {
            utils.ReportError("unknown error when applying keys: inserting record:", err)
        } else {
            utils.ReportError("unknown error when applying keys: inserting record: nothing affected")
        }
        return 4
    }
    return 0
}

// 归还序列号使用（返回信息参见协议）
func (dba *DBAccess) ReturnKeyExtend(ss *SessionInfo) int {
    if ss == nil {
        utils.ReportError("unknown error when returning key: got nil session info")
        return 4
    }
    // 查询钥匙信息
    keyInfo, err := dba.GetKeyInfo(ss.KeyNum)
    if err != nil {
        utils.ReportError("unknown error when returning key: getting key info:", err)
        return 4
    }
    if keyInfo == nil { // 002 该序列号无效
        return 2
    }
    if keyInfo.KeyResidue == keyInfo.KeyCapacity { // “006”：该序列号对应的凭据没有被该用户借出过
        return 6
    }
    // 查询是否已经存在了会话信息
    r, err := dba.Query(fmt.Sprintf(
        `SELECT KEY_NUM FROM SESSIONS WHERE SESSION_ID='%d'`, ss.SessionID))
    if err != nil {
        utils.ReportError("unknown error when returning key: selecting sessions:", err)
        return 4
    }
    if len(r.Values) == 0 { // “006”：该序列号对应的凭据没有被该用户借出过
        return 6
    }
    // 查询会话信息中的钥匙和客户提供的钥匙是否相同
    infoRow := r.Values[0]
    if infoRow[0].(string) != ss.KeyNum { // “008”：（RETU）该用户提供的归还密钥和记录中她借走的的密钥不匹配
        return 8
    }

    // 可以归还凭据
    aff, err := dba.Execute(fmt.Sprintf(`DELETE FROM SESSIONS 
WHERE SESSION_ID='%d';`, ss.SessionID))
    if err != nil || aff == 0 { // 未归还成功
        if err != nil {
            utils.ReportError("unknown error when returning keys: deleting sessions:", err)
        } else {
            utils.ReportError("unknown error when returning keys: deleting sessions: nothing affected")
        }
        return 4
    }
    return 0
}

// 归还该client server申请过的所有序列号使用，返回归还的钥匙数量
func (dba *DBAccess) ReturnCSAllKeyExtend(csNumber uint32) int {
    // 归还所有凭据
    aff, err := dba.Execute(fmt.Sprintf(`DELETE FROM SESSIONS 
WHERE CLIENT_SERVER_NUM=%d;`, csNumber))
    if err != nil { // 未归还成功
        utils.ReportError("unknown error when returning keys: deleting sessions:", err)
        return -1
    }
    return aff
}

// 更新客服编号至新编号
func (dba *DBAccess) UpdateCSNumberExtend(before int, after int) {
    // 归还凭据
    aff, err := dba.Execute(fmt.Sprintf(`UPDATE SESSIONS 
SET CLIENT_SERVER_NUM=%d WHERE CLIENT_SERVER_NUM=%d;`, after, before))
    if err != nil || aff == 0 { // 未归还成功
        if err != nil {
            utils.ReportError("unknown error when changing client server numbers: changing:", err)
        } else {
            utils.ReportWarning("changing client server numbers: nothing affected")
        }
    }
}