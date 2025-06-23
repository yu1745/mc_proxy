package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

var remoteAddr string
var gameAddr string
var cacheConnCount int
var connChan = make(chan struct{}, 1000)

func init() {
	flag.StringVar(&remoteAddr, "r", ":9999", "frps的监听地址2")
	flag.StringVar(&gameAddr, "g", "localhost:25565", "mc服务器地址")
	flag.IntVar(&cacheConnCount, "c", 10, "缓存连接数")
	flag.Parse()
}

func connect() {
	for {
		if len(connChan) < cacheConnCount {
			conn, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				continue
			}
			connChan <- struct{}{}
			go handle(conn)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func handle(conn net.Conn) {
	ip4 := make([]byte, 4)
	nR, err := conn.Read(ip4)
	if err != nil {
		fmt.Printf("读取ip失败: %v\n", err)
		return
	}
	if nR != 4 {
		fmt.Printf("读取ip失败: %v\n", err)
		return
	}
	ip6 := make([]byte, 16)
	//设置前缀fd80
	ip6[0] = 0xfd
	ip6[1] = 0x80
	copy(ip6[12:], ip4)
	// 连接游戏服务器
	lAddr := &net.TCPAddr{
		IP:   ip6,
		Port: 0,
	}
	host, port, err := net.SplitHostPort(gameAddr)
	if err != nil {
		fmt.Printf("解析地址失败: %v\n", err)
		return
	}
	portN, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("解析端口失败: %v\n", err)
		return
	}
	rAddr := &net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: portN,
	}
	netConn, err := net.DialTCP("ip6", lAddr, rAddr)
	if err != nil {
		fmt.Printf("连接游戏服务器失败: %v\n", err)
		return
	}
	for {
		io.Copy(netConn, conn)
		io.Copy(conn, netConn)
	}
}

func main() {
	go connect()
}
