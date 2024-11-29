package udp

import (
	"fmt"
	"net"
	"os"
)

func StartServer() (err error) {
	var listenPort string
	var serverAddr string
	var addr *net.UDPAddr
	var udpConn *net.UDPConn

	listenPort = os.Getenv("IPV6_SERVER_PORT")
	serverAddr = fmt.Sprintf("[::]:%s", listenPort)

	// 进行 udp 地址的解析
	addr, err = net.ResolveUDPAddr("udp6", serverAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// 监听 udp 地址
	udpConn, err = net.ListenUDP("udp6", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP client: %w", err)
	}
	defer func(udpConn *net.UDPConn) {

		closeErr := udpConn.Close()

		if err == nil {
			err = closeErr
		}

	}(udpConn)

	fmt.Println("Listening on UDP port " + listenPort)

	var n int
	var clientAddr *net.UDPAddr
	buffer := make([]byte, 1024)
	for {
		// 读取客户端的数据
		n, clientAddr, err = udpConn.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from UDP client: %w", err)
		}

		// 打印接收到的消息
		message := string(buffer[:n])
		fmt.Printf("receive message %s from client %s\n", message, clientAddr)
	}
}
