package subnet

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"net"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleUtilsLogger = logger.GetLogger(logger.ModuleUtils)

// GenerateSubnets 生成子网
func GenerateSubnets(baseNetworkAddress string) []iplib.Net4 {
	_, baseNet, err := net.ParseCIDR(baseNetworkAddress)
	if err != nil {
		moduleUtilsLogger.Errorf("parse cidr failed")
		return nil
	}
	baseIp := baseNet.IP
	maskLen, _ := baseNet.Mask.Size()
	netIpv4 := iplib.NewNet4(baseIp, maskLen)
	subNets, err := netIpv4.Subnet(30)
	if err != nil {
		moduleUtilsLogger.Errorf("get subnets failed")
		return nil
	}
	return subNets
}

func GenerateTwoAddrsFrom30MaskSubnet(subNet iplib.Net4) (string, string) {
	firstAddr, secondAddr := subNet.FirstAddress().String(), subNet.LastAddress().String()
	firstAddr, secondAddr = fmt.Sprintf("%s/30", firstAddr), fmt.Sprintf("%s/30", secondAddr)
	return firstAddr, secondAddr
}
