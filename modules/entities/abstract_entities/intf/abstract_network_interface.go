package intf

type NetworkInterface struct {
	Ifidx  int    // 接口索引
	IfName string // 接口名
	Addr   string // 接口 ip 地址
}

func NewNetworkInterface(Ifidx int, IfName string, Addr string) *NetworkInterface {
	return &NetworkInterface{
		Ifidx:  Ifidx,
		IfName: IfName,
		Addr:   Addr,
	}
}
