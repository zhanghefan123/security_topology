package tcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// StartClient 进行tcp客户端的启动
func StartClient() (err error) {
	var serverPort string
	var serverAddr string
	serverPort = os.Getenv("IPV6_SERVER_PORT")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input server addr: ")
	if scanner.Scan() {
		serverAddr = scanner.Text()
	}

	var connection net.Conn
	connection, err = net.Dial("tcp", fmt.Sprintf("[%s]:%s", serverAddr, serverPort))
	if err != nil {
		return fmt.Errorf("error connecting to server %s: %w", serverAddr, err)
	}
	defer func(connection net.Conn) {
		connCloseErr := connection.Close()
		if err == nil {
			err = connCloseErr
		}
	}(connection)

	// 循环进行用户输入的读取
	for {
		fmt.Println("Please input the message you want to send: ")
		if scanner.Scan() {
			line := scanner.Text()

			// 输入关键字时退出
			if line == "q" || line == "quit" {
				break
			}

			// 如果没有则进行发送
			_, err = connection.Write([]byte(line))
			if err != nil {
				return fmt.Errorf("error sending message to server %s: %w", serverAddr, err)
			}
		}
	}

	return nil
}
