package network

import (
	"net"
)

func GetAvailablePort() (int, error) {
	// 在 0 端口上监听，操作系统会分配一个可用端口
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer func(listener net.Listener) {
		closeError := listener.Close()
		if err == nil {
			err = closeError
		}
	}(listener) // 记得关闭监听器

	// 获取分配的地址，解析出端口号
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
