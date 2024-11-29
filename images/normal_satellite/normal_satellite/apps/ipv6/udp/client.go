package udp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// 可以选择的发送模式
const (
	InteractChoice    = "interact"
	BatchChoice       = "batch"
	QuitChoice        = "quit"
	QuitShorterChoice = "q"
)

func StartClient() (err error) {
	var serverPort string    // 服务器监听端口
	var serverAddr string    // 服务器监听地址
	var addr *net.UDPAddr    // 地址
	var udpConn *net.UDPConn // udp 连接

	// 获取监听端口
	serverPort = os.Getenv("IPV6_SERVER_PORT")

	// 获取服务器地址
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input server addr:")
	if scanner.Scan() {
		serverAddr = scanner.Text()
	}

	// 解析地址
	addr, err = net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:%s", serverAddr, serverPort))
	if err != nil {
		return fmt.Errorf("can't resolve udp address: %w", err)
	}

	// 创建 udp 连接
	udpConn, err = net.DialUDP("udp6", nil, addr)
	if err != nil {
		return fmt.Errorf("can't start udp client: %w", err)
	}
	// 在结束时删除连接
	defer func(udpConn *net.UDPConn) {
		udpCloseErr := udpConn.Close()
		if err == nil {
			err = udpCloseErr
		}
	}(udpConn)

	// 发送消息到服务器
	fmt.Println("Please input the send mode you want: (1.interact) (2.batch) (3.q or quit)")
	if scanner.Scan() {
		line := scanner.Text()
		switch line {
		case InteractChoice:
			err = HandleInteract(udpConn)
			if err != nil {
				return fmt.Errorf("can't handle interact: %w", err)
			}
		case BatchChoice:
			err = HandleBatch(udpConn)
			if err != nil {
				return fmt.Errorf("can't handle batch: %w", err)
			}
		case QuitChoice, QuitShorterChoice:
			return nil
		default:
			return fmt.Errorf("invalid send mode: %w", errors.New("error send mode"))
		}
	}
	return nil
}

// HandleInteract 处理交互式输入
func HandleInteract(udpConn *net.UDPConn) error {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input the message you want to send (q / quit to exit): ")
	for {
		if scanner.Scan() {
			line := scanner.Text()

			switch line {
			case QuitChoice, QuitShorterChoice:
				return nil
			default:
				// 如果没有则进行发送
				_, err = udpConn.Write([]byte(line))
				if err != nil {
					return fmt.Errorf("can't send message to udp client: %w", err)
				}
			}
		}
	}
}

// HandleBatch 处理批量输入
func HandleBatch(udpConn *net.UDPConn) error {
	var err error
	var numberOfMessages int
	var sizeOfMessage int
	var sendInterval float64
	scanner := bufio.NewScanner(os.Stdin)

	// 提取总共需要发送多少消息
	fmt.Println("Please input the number of message you want to send: ")
	if scanner.Scan() {
		line := scanner.Text()
		numberOfMessages, err = strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("can't parse message number: %w", err)
		}
	}

	// 提取每个消息的长度
	fmt.Println("Please input the size of the message (byte) you want to send: ")
	if scanner.Scan() {
		line := scanner.Text()
		sizeOfMessage, err = strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("can't parse message size: %w", err)
		}
		if sizeOfMessage > 1024 {
			return fmt.Errorf("invalid message size: %d larger than 1024", sizeOfMessage)
		}
	}

	// 提取发送的时间间隔
	fmt.Println("Please input the send interval: ")
	if scanner.Scan() {
		line := scanner.Text()
		sendInterval, err = strconv.ParseFloat(line, 64)
		if err != nil {
			return fmt.Errorf("can't parse send interval: %w", err)
		}
	}

	// 进行消息的构建
	for index := 0; index < numberOfMessages; index++ {
		message := strings.Repeat("a", sizeOfMessage)
		_, err = udpConn.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("can't send message to udp client: %w", err)
		}

		// 这里可以打印一下还要打印多少
		if index%100 == 0 {
			fmt.Println(fmt.Sprintf("total messages: %d current messages: %d", numberOfMessages, index+1))
		}

		time.Sleep(time.Duration(sendInterval*1000) * time.Millisecond)
	}

	return nil
}
