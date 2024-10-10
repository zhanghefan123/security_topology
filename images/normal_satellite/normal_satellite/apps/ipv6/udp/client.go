package udp

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func StartClient() (err error) {
	var serverPort string
	var serverAddr string
	var addr *net.UDPAddr
	var udpConn *net.UDPConn

	serverPort = os.Getenv("IPV6_SERVER_PORT")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input server addr:")
	if scanner.Scan() {
		serverAddr = scanner.Text()
	}

	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("[%s]:%s", serverAddr, serverPort))
	if err != nil {
		return fmt.Errorf("can't resolve udp address: %w", err)
	}

	// 创建 udp 连接
	udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("can't start udp client: %w", err)
	}
	// 在结束时删除连接
	defer func(udpConn *net.UDPConn) {
		err = udpConn.Close()
		if err != nil {
			err = fmt.Errorf("can't close udp client: %w", err)
		}
	}(udpConn)

	// 发送消息到服务器
	for {
		fmt.Println("Please input the message you want to send: ")
		if scanner.Scan() {
			line := scanner.Text()

			// 输入关键字时退出
			if line == "q" || line == "quit" {
				break
			}

			// 如果没有则进行发送
			_, err = udpConn.WriteToUDP([]byte(line), addr)
			if err != nil {
				return fmt.Errorf("can't send message to udp client: %w", err)
			}
		}
	}

	return nil
}
