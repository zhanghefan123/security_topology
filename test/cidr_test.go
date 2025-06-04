package test

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
)

func TestCidr(t *testing.T) {
	ip, network, _ := net.ParseCIDR("192.168.1.0/24")
	fmt.Println(ip.String())
	fmt.Println(network.String())
}

func TestExample(t *testing.T) {
	for _, index := range rand.Perm(5) {
		fmt.Println(index)
	}
}

func TestTime(t *testing.T) {
	value, _ := strconv.ParseBool("true")
	fmt.Println(value)
}
