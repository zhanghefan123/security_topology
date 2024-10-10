package main

import (
	"bufio"
	"fmt"
	"os"
	"zhanghefan123/security/normal_satellite/apps/ipv6/tcp"
	"zhanghefan123/security/normal_satellite/apps/ipv6/udp"
)

const (
	ClientChoice      = "client"
	ServerChoice      = "server"
	QuitChoice        = "quit"
	QuitShorterChoice = "q"
)

const (
	TcpProtocol = "tcp"
	UdpProtocol = "udp"
)

func HandleDifferentChoice(choice string) error {
	var err error
	switch choice {
	case QuitChoice:
	case QuitShorterChoice:
		fmt.Println("Program Exit")
		os.Exit(0)
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
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
	return nil
}

func HandleClientChoice() error {
	var err error
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(os.Stdin)
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
			return fmt.Errorf("invalid protocol choice: %s", line)
		}
	}
	return nil
}

func HandleServerChoice() error {
	var err error
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line := scanner.Text()
		switch line {
		case TcpProtocol:
			err = tcp.StartServer()
			if err != nil {
				return fmt.Errorf("error starting tcp server: %w", err)
			} else {
				return nil
			}
		case UdpProtocol:
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
			fmt.Printf("handle different choice error: %v", err)
		}
	}
}
