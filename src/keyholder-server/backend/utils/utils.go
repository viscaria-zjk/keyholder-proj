package utils

import (
    "crypto/rand"
    "encoding/binary"
    "fmt"
    "github.com/fatih/color"
    "math"
    "math/big"
    "net"
    "reflect"
    "time"
    "unsafe"
)

// 报告错误
func ReportError(errStr... interface{}) {
    fmt.Print("[", time.Now().Format("07/03/2006 15:04:05 MST"), "]")
    _, _ = color.New(color.FgRed).Print("ERROR: ")
    fmt.Println(errStr...)
}

// 报告警告k
func ReportWarning(warnStr... interface{}) {
    fmt.Print("[", time.Now().Format("07/03/2006 15:04:05 MST"), "]")
    _, _ = color.New(color.FgYellow).Print("WARNING: ")
    fmt.Println(warnStr...)
}

// 报告信息
func ReportInfo(infoStr... interface{}) {
    fmt.Print("[", time.Now().Format("07/03/2006 15:04:05 MST"), "]")
    _, _ = color.New(color.FgGreen).Print("INFO: ")
    fmt.Println(infoStr...)
}

// 获得一个时间的Unix时间戳
func GetUniversalTime(time time.Time) int64 {
    return time.UTC().Unix()
}

// 获得一个Unix时间戳的时间字符串
func RenderTime(timeStamp int64) (rend string) {
    timeUnix := time.Unix(timeStamp, 0)
    rend = timeUnix.Format("2006-01-02 15:04:05")
    return
}

// 获得一个数据库时间字符串（形如"2006-01-02T15:04:05Z"）的时间戳
func ParseTime(timeString string) int64 {
    dateTime, err := time.Parse("2006-01-02T15:04:05Z", timeString)
    if err != nil {
        return 0
    } else {
        return dateTime.Unix()
    }
}

// 获得一个日期字符串的（形如"2006-01-02"）Unix时间戳
func GerUniversalDate(date string) int64 {
    dateTime, err := time.Parse("2006-01-02", date)
    if err != nil {
        return time.Now().UTC().Unix()
    }
    return dateTime.UTC().Unix()
}

// 查询TCP/IP中的端口是否有在使用
func IsPortInUse(addr string, port int, network string) bool {
    ip, err := net.ResolveIPAddr("ip", addr)
    if err != nil {
        return true
    }
    tcpAddr := net.TCPAddr {
        IP:   ip.IP,
        Port: port,
    }
    // 尝试用tcp连接这个端口
    conn, err := net.DialTCP(network, nil, &tcpAddr)
    if err == nil {
        _ = conn.Close()
        return true
    }
    return false
}

// 随机指定一个UDP端口的监听器
func RandomUDPConn(randomAddr *net.UDPAddr) *net.UDPConn {
    ls, err := net.ListenUDP("udp", randomAddr)
    if err != nil {
        return nil
    }
    return ls
}

// 32位整数转[]byte
func Uint32ToBytes(i uint32) []byte {
    buf := make([]byte, 4)
    binary.LittleEndian.PutUint32(buf, i)
    return buf
}

// 64位整数转[]byte
func Uint64ToBytes(i uint64) []byte {
    buf := make([]byte, 8)
    binary.LittleEndian.PutUint64(buf, i)
    return buf
}

// []byte转32位（4字节）无符号整数
func BytesToInt32(buf []byte) uint32 {
    return binary.LittleEndian.Uint32(buf)
}

// []byte转64位（8字节）无符号整数
func BytesToInt64(buf []byte) uint64 {
    return binary.LittleEndian.Uint64(buf)
}

// 将byte无拷贝转为string
func ToString(b []byte) (s string) {
    pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
    pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
    pString.Data = pBytes.Data
    pString.Len = pBytes.Len
    return
}

// 将string无拷贝转为byte
func ToSlice(s string) (b []byte) {
    pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
    pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
    pBytes.Data = pString.Data
    pBytes.Len = pString.Len
    pBytes.Cap = pString.Len
    return
}

// 用真随机生成一个8字节的会话ID（为uint32）
func GenSesID() (sesID uint64) {
    n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
    return n.Uint64()
}

// 用真随机生成一个8字节的序列号（全都为可视字元）
const str = "0123456789@qwertyuiopasdfghjklzxcvbnm#QWERTYUIOPASDFGHJKLZXCVBNM$"
func GenSerialNum(long int) (sesID string) {
    bytes := ToSlice(str)
    result := make([]byte, long)
    for i := 0; i < long; i++ {
        n, _ := rand.Int(rand.Reader, big.NewInt(65))
        result[i] = bytes[n.Int64()]
    }
    return ToString(result)
}