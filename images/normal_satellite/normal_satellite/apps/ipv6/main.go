package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"zhanghefan123/security/normal_satellite/apps/ipv6/tcp"
	"zhanghefan123/security/normal_satellite/apps/ipv6/udp"
)

// 用户可以选择的选项
const (
	ClientChoice      = "client"
	ServerChoice      = "server"
	QuitChoice        = "quit"
	QuitShorterChoice = "q"
)

// 用户可以选择的协议
const (
	TcpProtocol = "tcp"
	UdpProtocol = "udp"
)

// HandleDifferentChoice 处理不同的选择
func HandleDifferentChoice(choice string) error {
	var err error
	switch choice {
	case ClientChoice:
		err = HandleClientChoice()
		if err != nil {
			return fmt.Errorf("start client error: %w", err)
		}
	case ServerChoice:
		err = HandleServerChoice()
		if err != nil {
			return fmt.Errorf("start server error: %w", err)
		}
	case QuitChoice, QuitShorterChoice:
		fmt.Println("Program Exit")
		os.Exit(0)
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
	return nil
}

// HandleClientChoice 当用户选择了创建客户端的时候
func HandleClientChoice() error {
	var err error
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("Please input the protocol type (udp/tcp): ")
	if scanner.Scan() {
		line := scanner.Text()
		switch line {
		case TcpProtocol:
			err = tcp.StartClient()
			if err != nil {
				return fmt.Errorf("error starting tcp client: %w", err)
			} else {
				return nil
			}
		case UdpProtocol:
			err = udp.StartClient()
			if err != nil {
				return fmt.Errorf("error starting udp client: %w", err)
			} else {
				return nil
			}
		default:
			return fmt.Errorf("invalid protocol choice: %w", errors.New("error protocol"))
		}
	}
	return nil
}

// HandleServerChoice 当用户选择创建了服务器的时候
func HandleServerChoice() error {
	var err error
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("Please input the protocol type (udp/tcp): ")
	if scanner.Scan() {
		line := scanner.Text()
		switch line {
		case TcpProtocol: // 如果用户选择了 TCP 协议
			err = tcp.StartServer()
			if err != nil {
				return fmt.Errorf("error starting tcp server: %w", err)
			} else {
				return nil
			}
		case UdpProtocol: // 如果用户选择了 UDP 协议
			err = udp.StartServer()
			if err != nil {
				return fmt.Errorf("error starting udp server: %w", err)
			} else {
				return nil
			}
		default:
			return fmt.Errorf("invalid protocol choice: %s", line)
		}
	}
	return nil
}

func main() {
	// 获取用户输入, 看用户选择的是客户端还是服务器端
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input (1.client) (2.server) (3.q or quit)")
	if scanner.Scan() {
		line := scanner.Text()
		err := HandleDifferentChoice(line)
		if err != nil {
			fmt.Println(fmt.Sprintf("handle different choice error: %v\n", err))
			_, _ = os.Create("test.txt")
		}
	}
}
