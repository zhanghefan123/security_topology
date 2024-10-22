package network

import (
	"fmt"
	"testing"
)

func TestSubnetGenerator(t *testing.T) {
	GenerateIPv4Subnets("192.168.0.0/16")
}

func TestGenerateIpv6Subnets(t *testing.T) {
	subnets, err := GenerateIpv6Subnets("2001:db8:1234:5678::/112")
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10; index++ {
		subnet := subnets[index]
		fmt.Println(subnet)
		//firstAddr, secondAddr := subnet.FirstAddress().String(), subnet.LastAddress().String()
		//fmt.Println(firstAddr, secondAddr)
	}
}
