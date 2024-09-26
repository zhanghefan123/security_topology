package subnet

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"net"
)

// GenerateSubnets 生成子网
func GenerateSubnets(baseNetworkAddress string) (subNets []iplib.Net4, err error) {
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

// GenerateTwoAddrsFrom30MaskSubnet 从 30 位的子网掩码的网络中生成两个地址
func GenerateTwoAddrsFrom30MaskSubnet(subNet iplib.Net4) (string, string) {
	firstAddr, secondAddr := subNet.FirstAddress().String(), subNet.LastAddress().String()
	firstAddr, secondAddr = fmt.Sprintf("%s/30", firstAddr), fmt.Sprintf("%s/30", secondAddr)
	return firstAddr, secondAddr
}
