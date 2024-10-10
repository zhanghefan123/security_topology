package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func StartClient() (err error) {
	var serverAddr string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input server addr: ")
	if scanner.Scan() {
		serverAddr = scanner.Text()
	}

	var connection net.Conn
	connection, err = net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to server %s: %w", serverAddr, err)
	}
	defer func(connection net.Conn) {
		err = connection.Close()
		if err != nil {
			err = fmt.Errorf("error closing connection to server %s: %w", serverAddr, err)
		}
	}(connection)

	// 循环进行用户输入的读取
	for {
		streamScanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Please input the message you want to send: ")
		if streamScanner.Scan() {
			line := streamScanner.Text()

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
