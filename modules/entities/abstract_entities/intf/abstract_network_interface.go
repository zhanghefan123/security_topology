package intf

type NetworkInterface struct {
	Ifidx          int    // 接口索引
	IfName         string // 接口名
	SourceIpv4Addr string // 接口 ipv4 地址
	SourceIpv6Addr string // 接口 ipv6 地址
	TargetIpv4Addr string // 对侧的 ipv4 地址
	TargetIpv6Addr string // 对侧的 ipv6 地址
	LinkIdentifier int    // 链路标识
}

// NewNetworkInterface 进行新的网络接口的创建
func NewNetworkInterface(Ifidx int, IfName, SourceIpv4Addr, SourceIpv6Addr, TargetIpv4Addr, TargetIpv6Addr string, LinkIdentifier int) *NetworkInterface {
	return &NetworkInterface{
		Ifidx:          Ifidx,
		IfName:         IfName,
		SourceIpv4Addr: SourceIpv4Addr,
		SourceIpv6Addr: SourceIpv6Addr,
		TargetIpv4Addr: TargetIpv4Addr,
		TargetIpv6Addr: TargetIpv6Addr,
		LinkIdentifier: LinkIdentifier,
	}
}
