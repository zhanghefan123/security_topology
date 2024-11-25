package intf

type NetworkInterface struct {
	Ifidx          int    // 接口索引
	IfName         string // 接口名
	Ipv4Addr       string // 接口 ipv4 地址
	Ipv6Addr       string // 接口 ipv6 地址
	LinkIdentifier int    // 链路标识
}

// NewNetworkInterface 进行新的网络接口的创建
func NewNetworkInterface(Ifidx int, IfName, Ipv4Addr, Ipv6Addr string, LinkIdentifier int) *NetworkInterface {
	return &NetworkInterface{
		Ifidx:          Ifidx,
		IfName:         IfName,
		Ipv4Addr:       Ipv4Addr,
		Ipv6Addr:       Ipv6Addr,
		LinkIdentifier: LinkIdentifier,
	}
}
