package network

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"net"
)

// GenerateIPv4Subnets 生成 Ipv4 子网
func GenerateIPv4Subnets(baseNetworkAddress string) (subNets []iplib.Net4, err error) {
	var baseNet *net.IPNet
	_, baseNet, err = net.ParseCIDR(baseNetworkAddress)
	if err != nil {

		return nil, fmt.Errorf("parse cidr failed")
	}
	baseIp := baseNet.IP
	maskLen, _ := baseNet.Mask.Size()
	netIpv4 := iplib.NewNet4(baseIp, maskLen)
	subNets, err = netIpv4.Subnet(30)
	if err != nil {
		return nil, fmt.Errorf("get subnets failed")
	}
	return subNets, nil
}

// GenerateIpv6Subnets 生成 Ipv6 子网
func GenerateIpv6Subnets(baseNetworkAddress string) (subNets []iplib.Net6, err error) {
	var baseNet *net.IPNet
	_, baseNet, err = net.ParseCIDR(baseNetworkAddress)
	if err != nil {
		return nil, fmt.Errorf("parse cidr failed")
	}
	baseIp := baseNet.IP
	maskLen, _ := baseNet.Mask.Size()
	netIpv6 := iplib.NewNet6(baseIp, maskLen, 0)
	subNets, err = netIpv6.Subnet(127, 0)
	if err != nil {
		return nil, fmt.Errorf("get subnets failed %w", err)
	}
	return subNets, nil
}

// GenerateTwoAddrsFromIpv4Subnet 从子网中生成两个地址
func GenerateTwoAddrsFromIpv4Subnet(subNet iplib.Net4) (string, string) {
	firstAddr, secondAddr := subNet.FirstAddress().String(), subNet.LastAddress().String()
	firstAddr, secondAddr = fmt.Sprintf("%s/30", firstAddr), fmt.Sprintf("%s/30", secondAddr)
	return firstAddr, secondAddr
}

// GenerateTwoAddrsFromIpv6Subnet 从子网之中生成两个地址
func GenerateTwoAddrsFromIpv6Subnet(subNet iplib.Net6) (string, string) {
	firstAddr, secondAddr := subNet.FirstAddress().String(), subNet.LastAddress().String()
	firstAddr, secondAddr = fmt.Sprintf("%s/127", firstAddr), fmt.Sprintf("%s/127", secondAddr)
	return firstAddr, secondAddr
}
