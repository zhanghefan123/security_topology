package tcp

import (
	"fmt"
	"net"
	"os"
)

// StartServer 进行服务端的启动
func StartServer() (err error) {
	var listenPort string
	var listenAddr string

	listenPort = os.Getenv("IPV6_SERVER_PORT")
	listenAddr = fmt.Sprintf("[::]:%s", listenPort)

	// 进行地址的指定的地址的监听
	var listener net.Listener
	listener, err = net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("Error listening on %s: %w\n", listenAddr, err)
	}
	defer func(listener net.Listener) {
		listenerCloseErr := listener.Close()
		if err == nil {
			err = listenerCloseErr
		}
	}(listener)
	fmt.Println("Listening on ", listenAddr)

	// 循环的进行处理
	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			return fmt.Errorf("Error accepting connection: %w\n", err)
		}
		fmt.Println("New connection from", conn.RemoteAddr())
		go func() {
			err = HandleConnection(conn)
			if err != nil {
				err = fmt.Errorf("Error handling connection: %w\n", err)
			}
		}()
	}
}

func HandleConnection(connection net.Conn) (err error) {
	// 在最后进行连接的关闭
	defer func(connection net.Conn) {
		connCloseErr := connection.Close()
		if err == nil {
			err = connCloseErr
		}

		if connection.Close() != nil {
			err = fmt.Errorf("Error closing connection: %w\n", err)
		}
	}(connection)

	buffer := make([]byte, 1024)

	// 进行循环的读取
	var n int
	for {
		n, err = connection.Read(buffer)
		// 如果读取的时候发生错误
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("EOF")
				return nil
			}
			return fmt.Errorf("Error reading from connection: %w\n", err)
		}
		// 如果读取正常则进行结果的打印
		fmt.Printf("Received message from client: %s\n", string(buffer[:n]))
	}
}
