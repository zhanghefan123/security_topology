package linux_tc_api

import (
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
)

// 后缀是 _windows 的时候, 当 windows 的时候会编译

const (
	LargeBandwidth = 100 * 1e6 // 100 mbps 的带宽
)

// SetInterfacesDelay 设置某个容器接口的延迟
func SetInterfacesDelay(containerPid int, interfaces []string, delays []float64) (err error) {
	return nil
}

// SetInterfaceBandwidth 设置带宽
func SetInterfaceBandwidth(containerInterface *intf.NetworkInterface, containerPid int, bandWidth int) error {
	return nil
}
