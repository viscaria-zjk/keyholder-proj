package proto

import (
    "keyholder/storage"
    "keyholder/utils"
)

// 返回0， 表示申请成功；1：该序列号使用人数已满；2：该序列号无效；
// 3：该序列号下，已经有一个相同唯一标识码的用户在线；4：未知错误；
func ApplyKey(db *storage.DBAccess, signal []byte, csNumber uint32) (sesID uint64, info int) {
    if db == nil { return 0, 4 }
    if len(signal) < 26 { return 0, 4 }
    // APPL 序列号，8字节 ⽤户名⻓长度，4字节 ⽤户名 密码⻓长度，4字节 密码
    usernameLen := utils.BytesToInt32(signal[14:18])
    utils.ReportInfo("got username:", utils.ToString(signal[19:19 + usernameLen]))
    if len(signal) < 26 + int(usernameLen) {
        utils.ReportError("instruct is too short: in username parsing, need at least", 26 + int(24 + usernameLen), ", got", len(signal))
        return 0, 4
    }
    passwordLen := utils.BytesToInt32(signal[20 + usernameLen:24 + usernameLen])
    utils.ReportInfo("got password:", utils.ToString(signal[25 + usernameLen:25 + usernameLen + passwordLen]))
    if len(signal) < 25 + int(usernameLen + passwordLen) {
        utils.ReportError("instruct is too short: in password parsing, need at least", 25 + int(usernameLen + passwordLen), ", got", len(signal))
        return 0, 4
    }
    ssi := storage.SessionInfo {
        SessionID:    0,
        KeyNum:       utils.ToString(signal[5:13]),
        Username:     utils.ToString(signal[19:19 + usernameLen]),
        Password:     utils.ToString(signal[25 + usernameLen:25 + usernameLen + passwordLen]),
        ClientServer: csNumber,
    }
    // 运行数据库
    info = 3
    for info == 3 {
        // 生成新的唯一会话标识码
        ssi.SessionID = utils.GenSesID()
        info = db.ApplyKeyExtend(&ssi)
    }
    // 返回结果
    sesID = ssi.SessionID
    return
}

// 返回0， 表示归还成功；2：该序列号无效；
// 4：未知错误；6：该序列号对应的凭据没有被该用户借出过
func ReturnKey(db *storage.DBAccess, signal []byte, csNumber uint32) int {
    if db == nil { return 4 }
    if len(signal) < 22 { return 4 }
    // RETU 序列号，8字节 SessionID，8字节
    ssi := storage.SessionInfo {
        SessionID:    utils.BytesToInt64(signal[14:22]),
        KeyNum:       utils.ToString(signal[5:13]),
        Username:     "",
        Password:     "",
        ClientServer: csNumber,
    }
    // 运行数据库及返回结果
    return db.ReturnKeyExtend(&ssi)
}

// 在服务器崩溃重启后，在数据库中，更新原本属于某个客服编号的记录至新客服编号的记录
func UpdateCSNumber(db *storage.DBAccess, beforeCsNumber int, afterCsNumber int) {
    db.UpdateCSNumberExtend(beforeCsNumber, afterCsNumber)
}

// 客服结束运行的处理：强制归还它申请的所有凭据
// 传入它的客服编号（即端口号）
func CSOver(db *storage.DBAccess, csNumber uint32) {
    info := db.ReturnCSAllKeyExtend(csNumber)
    if info > 0 {
        utils.ReportInfo("client server", csNumber, "returned", info, "keys before it terminates")
    }
}