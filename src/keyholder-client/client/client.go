package client

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "key-client/utils"
    "net"
    "strconv"
    "time"
)

type Client struct {
    CSPort  int          // 缺省客服UDP端口
    CSConn  net.Conn     // 客服套接字
    Timeout int          // 和客服沟通的超时
    BuffLen int          // 缓冲区的大小
    Title   string       // 客服标题
    SesID   uint64       // 会话ID
    KHID    string       // KHID
}

const csPort = 9898 // 缺省客服UDP端口

// 新建一个客户端
func NewClient(softName string, timeout int, csAddr string, port int) *Client {
    cl := new(Client)
    cl.CSPort = port
    if port <= 0 {
        cl.CSPort = csPort
    }
    // 检查该端口是否占用
    isPortInUse := utils.IsPortInUse("127.0.0.1", cl.CSPort, "udp")
    if !isPortInUse {
        return nil
    }

    // 新建套接字
    fullAddr := fmt.Sprintf("%s:%d", csAddr, cl.CSPort)
    var err error
    cl.CSConn, err = net.Dial("udp", fullAddr)
    if err != nil {
        return nil
    }

    cl.Timeout = timeout
    cl.Title = softName
    return cl
}

// 发送钥匙 (错误码：0——10继承；11：发送数据包出错；12：设置超时时出错；13：超时发生；14：回应太短；15：回应非法)
func (cl *Client) Request(user string, pass string) int {
    buffWrite := new(bytes.Buffer)
    buffTemp := make([]byte, 30)

    // 1. 构造数据报(APPL)
    buffWrite.WriteString("APPL ")
    buffWrite.WriteString(cl.KHID[:8] + " ")

    binary.LittleEndian.PutUint32(buffTemp[:4], uint32(len(user)))
    buffWrite.Write(buffTemp[:4])
    buffWrite.WriteString(" " + user + " ")

    binary.LittleEndian.PutUint32(buffTemp[:4], uint32(len(pass)))
    buffWrite.Write(buffTemp[:4])
    buffWrite.WriteString(" " + pass)

    // 2. 发送数据报(APPL)
    _, err := cl.CSConn.Write(buffWrite.Bytes())
    if err != nil {
        return 11
    }

    // 3. 接收回应
    err = cl.CSConn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(cl.Timeout)))
    if err != nil {
        return 12
    }
    n, err := cl.CSConn.Read(buffTemp)
    if err != nil {
        if err.(net.Error).Timeout() == true {
            return 13
        }
        return 4
    }
    if n < 8 {
        return 14
    }

    // 4. 分析
    switch utils.ToString(buffTemp[:4]) {
    case "OKAY":
        if n < 13 {
            return 14
        }
        sesIDBuff := buffTemp[5:13]
        cl.SesID = utils.BytesToUInt64(sesIDBuff)
        return 0
    case "NOOK":
        nookCode, err := strconv.Atoi(utils.ToString(buffTemp[5:8]))
        if err != nil {
            return 15
        }
        return nookCode
    }
    return 4
}

// 归还钥匙(错误码：0——10继承；11：发送数据包出错；12：设置超时时出错；13：超时发生；14：回应太短；15：回应非法)
func (cl *Client) Return() int {
    buffWrite := new(bytes.Buffer)
    buffTemp := make([]byte, 30)

    // 1. 构造数据报(RETU)
    buffWrite.WriteString("RETU ")
    buffWrite.WriteString(cl.KHID[:8] + " ")

    binary.LittleEndian.PutUint64(buffTemp[:8], cl.SesID)
    buffWrite.Write(buffTemp[:8])

    // 2. 发送数据报(RETU)
    _, err := cl.CSConn.Write(buffWrite.Bytes())
    if err != nil {
        return 11
    }

    // 3. 接收回应
    err = cl.CSConn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(cl.Timeout)))
    if err != nil {
        return 12
    }
    n, err := cl.CSConn.Read(buffTemp)
    if err != nil {
        if err.(net.Error).Timeout() == true {
            return 13
        }
        return 4
    }
    if n < 4 {
        return 14
    }

    // 4. 分析回应
    switch utils.ToString(buffTemp[:4]) {
    case "OKAY":
        return 0
    case "NOOK":
        nookCode, err := strconv.Atoi(utils.ToString(buffTemp[5:8]))
        if err != nil {
            return 15
        }
        return nookCode
    }
    return 4
}

// 关闭一个客户端
func (cl *Client) CloseClient() error {
    err := cl.CSConn.Close()
    return err
}