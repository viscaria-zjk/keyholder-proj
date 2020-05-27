package utils

import (
    "encoding/binary"
    "net"
    "reflect"
    "unsafe"
)

// 查询TCP/IP中的端口是否有在使用
func IsPortInUse(addr string, port int, network string) bool {
    ip, err := net.ResolveIPAddr("ip", addr)
    if err != nil {
        return true
    }
    if network == "tcp" {
        // 尝试用tcp连接这个端口
        tcpAddr := net.TCPAddr {
            IP:   ip.IP,
            Port: port,
        }
        conn, err := net.DialTCP(network, nil, &tcpAddr)
        if err == nil {
            _ = conn.Close()
            return true
        }
    } else if network == "udp" {
        // 尝试占用这个端口
        udpAddr := net.UDPAddr {
            IP: ip.IP,
            Port: port,
        }
        conn, err := net.ListenUDP("udp", &udpAddr)
        if err != nil {
            return true
        }
        _ = conn.Close()
    }
    return false
}

// 64位整数转[]byte
func Uint64ToBytes(i uint64) []byte {
    buf := make([]byte, 8)
    binary.LittleEndian.PutUint64(buf, i)
    return buf
}

// 将byte无拷贝转为string
func ToString(b []byte) (s string) {
    pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
    pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
    pString.Data = pBytes.Data
    pString.Len = pBytes.Len
    return
}

// []byte转64位（8字节）无符号整数
func BytesToUInt64(buf []byte) uint64 {
    return binary.LittleEndian.Uint64(buf)
}