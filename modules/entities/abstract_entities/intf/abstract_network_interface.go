package intf

import "github.com/vishvananda/netlink"

type NetworkInterface struct {
	Ifidx          int           // 接口索引
	IfName         string        // 接口名
	SourceIpv4Addr string        `json:"source-ipv4"` // 接口 ipv4 地址
	SourceIpv6Addr string        `json:"source-ipv6"` // 接口 ipv6 地址
	TargetIpv4Addr string        // 对侧的 ipv4 地址
	TargetIpv6Addr string        // 对侧的 ipv6 地址
	LinkIdentifier int           // 链路标识
	Veth           *netlink.Link // 接口所对应的 veth
}

// NewNetworkInterface 进行新的网络接口的创建
func NewNetworkInterface(Ifidx int, IfName, SourceIpv4Addr,
	SourceIpv6Addr, TargetIpv4Addr, TargetIpv6Addr string,
	LinkIdentifier int, Veth *netlink.Link) *NetworkInterface {
	return &NetworkInterface{
		Ifidx:          Ifidx,
		IfName:         IfName,
		SourceIpv4Addr: SourceIpv4Addr,
		SourceIpv6Addr: SourceIpv6Addr,
		TargetIpv4Addr: TargetIpv4Addr,
		TargetIpv6Addr: TargetIpv6Addr,
		LinkIdentifier: LinkIdentifier,
		Veth:           Veth,
	}
}
