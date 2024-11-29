package normal_node

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/types"
)

// NormalNode 基础的网络节点
type NormalNode struct {
	Type                    types.NetworkNodeType   // 节点的类型
	Status                  types.NetworkNodeStatus // 节点状态
	Id                      int                     // 节点编号
	Pid                     int                     // 对应的进程编号
	Ifidx                   int                     // 接口索引
	Interfaces              []*intf.NetworkInterface
	IfNameToInterfaceMap    map[string]*intf.NetworkInterface // 从接口名称到对应的 ip 地址的映射
	ConnectedIpv4SubnetList []string                          // 连接到的 IPV4 的子网
	ConnectedIpv6SubnetList []string                          // 连接到的 IPV6 的子网
	ContainerName           string                            // 对应的容器的名称
	ContainerId             string                            // 容器的 ID
	X                       float64                           // 在前端之中的横坐标
	Y                       float64                           // 在前端之中的纵坐标
}

// NewNormalNode 创建普通系欸但
func NewNormalNode(typ types.NetworkNodeType, id int, containerName string) *NormalNode {
	return &NormalNode{
		Type:                    typ,
		Status:                  types.NetworkNodeStatus_Logic,
		Id:                      id,
		Ifidx:                   1,
		Interfaces:              make([]*intf.NetworkInterface, 0),
		IfNameToInterfaceMap:    make(map[string]*intf.NetworkInterface),
		ConnectedIpv4SubnetList: make([]string, 0),
		ContainerName:           containerName,
	}
}

// SetVethNamespace 设置 veth 命名空间
func (normalNode *NormalNode) SetVethNamespace() (err error) {
	// 1. 获取环境的 namespace
	hostNetNs, err := netns.Get()
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}
	defer func(ns netns.NsHandle) {
		nsSetErr := netns.Set(ns)
		if err == nil {
			err = nsSetErr
		}
	}(hostNetNs)

	// 2. 获取 pid
	pid := normalNode.Pid

	// 3. 获取容器的 netns
	netNs, err := netns.GetFromPid(pid)
	defer func(netNs *netns.NsHandle) {
		nsCloseErr := netNs.Close()
		if err == nil {
			err = nsCloseErr
		}
	}(&netNs)
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}

	// 4. 遍历卫星的接口并设置到命名空间之中
	var veths []netlink.Link
	var veth netlink.Link
	for _, networkInterface := range normalNode.IfNameToInterfaceMap {
		// find veth by name
		veth, err = netlink.LinkByName(networkInterface.IfName)
		if err != nil {
			return fmt.Errorf("netlink.LinkByName(%s) failed: %w", networkInterface.IfName, err)
		}
		// set veth into namespace
		if err = netlink.LinkSetNsFd(veth, int(netNs)); err != nil {
			return fmt.Errorf("netlink.LinkSetNsFd(%d) failed: %w", veth, err)
		}
		veths = append(veths, veth)
	}

	// 2. 进入容器命名空间启动 veth 设置 addr

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 切换到容器的网络命名空间
	if err = netns.Set(netNs); err != nil {
		return fmt.Errorf("netns.Set() failed: %w", err)
	}

	for _, veth = range veths {
		// 启动容器接口
		err = netlink.LinkSetUp(veth)
		if err != nil {
			return fmt.Errorf("netlink.LinkSetUp(%d) failed: %w", veth, err)
		}

		// 设置 ipv4 地址
		ifName := veth.Attrs().Name
		ipv4Addr := normalNode.IfNameToInterfaceMap[ifName].Ipv4Addr
		ipv4, _ := netlink.ParseAddr(ipv4Addr)
		if err = netlink.AddrAdd(veth, ipv4); err != nil {
			return fmt.Errorf("netlink.AddrAdd(%s) failed: %w", ipv4, err)
		}

		// 设置 ipv6 地址
		ipv6Addr := normalNode.IfNameToInterfaceMap[ifName].Ipv6Addr
		ipv6, _ := netlink.ParseAddr(ipv6Addr)
		if err = netlink.AddrAdd(veth, ipv6); err != nil {
			fmt.Printf("netlink.AddrAdd(%s) failed: %v", ipv6, err)
			return fmt.Errorf("netlink.AddrAdd(%s) failed: %w", ipv6, err)
		}

	}
	return nil
}

func (normalNode *NormalNode) GetId() int {
	return normalNode.Id
}

func (normalNode *NormalNode) AppendIpv4Subnet(ipv4Subnet string) {
	normalNode.ConnectedIpv4SubnetList = append(normalNode.ConnectedIpv4SubnetList, ipv4Subnet)
}

func (normalNode *NormalNode) AppendIpv6Subnet(ipv6Subnet string) {
	normalNode.ConnectedIpv6SubnetList = append(normalNode.ConnectedIpv6SubnetList, ipv6Subnet)
}
