package test

import (
	"fmt"
	"net"
	"testing"
)

func TestCidr(t *testing.T) {
	ip, network, _ := net.ParseCIDR("192.168.1.0/24")
	fmt.Println(ip.String())
	fmt.Println(network.String())
}
