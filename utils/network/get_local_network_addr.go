package network

import (
	"fmt"
	"net"
)

func GetLocalNetworkAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("cannot connect to local network address %w", err)
	}
	addr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	return addr, nil
}
