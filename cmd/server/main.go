package main

import (
	"flag"
	"fmt"
	"io"
	"net"
)

var listenAddr1 string
var listenAddr2 string
var connChan = make(chan net.Conn, 100) // 全局连接channel，缓冲大小100

func init() {
	flag.StringVar(&listenAddr1, "l1", ":25565", "监听地址1,玩家连接的地址")
	flag.StringVar(&listenAddr2, "l2", ":9999", "监听地址2.frpc连接的地址")
	flag.Parse()
}

func main() {
	go listen1() // 监听玩家连接
	go listen2() // 监听frpc连接
	// 创建一个永不关闭的通道，用于阻塞主函数，防止程序退出
	bl := make(chan struct{})
	<-bl
}

func handleForward(conn net.Conn) {
	//获取ip转成4字节数组
	ip := conn.RemoteAddr().(*net.TCPAddr).IP
	ip4 := ip.To4()
	if ip4 == nil {
		fmt.Printf("不是ipv4地址: %v\n", ip)
		return
	}
	conn2 := <-connChan
	//前四个字节是真实ip
	conn2.Write(ip4)
	//开始转发
	go io.Copy(conn2, conn)
	go io.Copy(conn, conn2)
}

// 监听玩家连接
func listen1() {
	listener, err := net.Listen("tcp", listenAddr1)
	if err != nil {
		panic(fmt.Sprintf("监听地址1失败: %v", err))
	}
	defer listener.Close()

	fmt.Printf("监听地址1: %s\n", listenAddr1)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接错误: %v\n", err)
			continue
		}
		//收到玩家连接就开始转发
		go handleForward(conn)
	}
}

// 监听frpc连接
func listen2() {
	listener, err := net.Listen("tcp", listenAddr2)
	if err != nil {
		panic(fmt.Sprintf("监听地址2失败: %v", err))
	}
	defer listener.Close()

	fmt.Printf("监听地址2: %s\n", listenAddr2)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接错误: %v\n", err)
			continue
		}
		connChan <- conn // 将新连接放入channel
	}
}
