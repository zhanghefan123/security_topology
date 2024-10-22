//go:build linux

package linux_tc_api

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
)

// NetQdiscTemplate 是一个 qdisc 模版
var NetQdiscTemplate = netlink.NewNetem(
	// 第一个参数是 QdiscAttrs，用于指定 Qdisc 的基本属性。
	netlink.QdiscAttrs{
		// Qdisc 的唯一标识符，用于在内核中区分不同的 Qdisc。
		// netlink.MakeHandle(1, 0) 生成了一个 Handle 值，表示这个 Qdisc 的句柄。
		Handle: netlink.MakeHandle(1, 0),
		// 这是 Qdisc 的父对象，netlink.HANDLE_ROOT 表示该 Qdisc 挂载在根（即网络设备本身）上。
		Parent: netlink.HANDLE_ROOT,
	},
	// NetemQdiscAttrs 是一个结构体，用于指定 netem 的网络仿真参数，例如延迟、抖动、丢包率等。
	netlink.NetemQdiscAttrs{},
)

// SetInterfacesDelay 设置某个容器接口的延迟
func SetInterfacesDelay(containerPid int, interfaces []string, delays []float64) (err error) {
	var hostNetNs netns.NsHandle

	// 1. 获取环境的 namespace
	hostNetNs, err = netns.Get()
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}

	// 2. 最后记住进行 ns 的释放
	defer func(ns netns.NsHandle) {
		err = netns.Set(ns)
	}(hostNetNs)

	// 3. 获取容器的 ns
	netNs, err := netns.GetFromPid(containerPid)
	defer func(netNs *netns.NsHandle) {
		err = netNs.Close()
		if err != nil {
			err = fmt.Errorf("netns.Get() failed: %w", err)
		}
	}(&netNs)

	// 4. 进行运行时的加锁，以及最后的释放锁
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 5. 切换到容器的网络命名空间
	if err = netns.Set(netNs); err != nil {
		return fmt.Errorf("netns.Set() failed: %w", err)
	}
	// 6. 进行所有要设置延迟的接口的便利
	for index := 0; index < len(interfaces); index++ {
		var vethInterface netlink.Link
		vethInterface, err = netlink.LinkByName(interfaces[index])
		if err != nil {
			return fmt.Errorf("netlink.LinkByName(%s) failed: %w", interfaces[index], err)
		}
		netemInfo := netlink.NewNetem(
			NetQdiscTemplate.QdiscAttrs,
			netlink.NetemQdiscAttrs{
				Latency: uint32(delays[index] * 1000),
			},
		)
		netemInfo.LinkIndex = vethInterface.Attrs().Index
		err = netlink.QdiscReplace(netemInfo)
		if err != nil {
			return fmt.Errorf("netlink.QdiscReplace() failed: %w", err)
		}
	}
	return nil
}
