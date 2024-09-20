package normal_node

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleNormalNodeLogger = logger.GetLogger(logger.ModuleNormalNode)

// NormalNode 基础的网络节点
type NormalNode struct {
	Status               types.NetworkNodeStatus           // 节点状态
	Id                   int                               // 节点编号
	Pid                  int                               // 对应的进程编号
	Ifidx                int                               // 接口索引
	IfNameToInterfaceMap map[string]*intf.NetworkInterface // 从接口索引到对应的 ip 地址的映射
	ConnectedSubnetList  []string                          // 连接到的子网的数量
	ContainerName        string                            // 对应的容器的名称
	ContainerId          string                            // 容器的 ID
}

func (normalNode *NormalNode) SetVethNamespace() {

	// 1. 将主机命名空间之中的 veth 设置到正确的 namespace

	// 获取环境的 namespace
	hostNetNs, err := netns.Get()
	if err != nil {
		moduleNormalNodeLogger.Errorf("Get host netns error %v", err)
	}
	defer func(ns netns.NsHandle) {
		err := netns.Set(ns)
		if err != nil {
			moduleNormalNodeLogger.Errorf("reset netns error %v", err)
		}
	}(hostNetNs)

	// 获取 pid
	pid := normalNode.Pid

	// 获取容器的 netns
	netNs, err := netns.GetFromPid(pid)
	defer func(netNs *netns.NsHandle) {
		err := netNs.Close()
		if err != nil {
			moduleNormalNodeLogger.Errorf("close netns error %v", err)
		}
	}(&netNs)
	if err != nil {
		moduleNormalNodeLogger.Errorf("Get netns error %v", err)
	}

	// 遍历卫星的接口并设置到命名空间之中
	var veths []netlink.Link
	for _, networkInterface := range normalNode.IfNameToInterfaceMap {
		// find veth by name
		veth, err := netlink.LinkByName(networkInterface.IfName)
		if err != nil {
			moduleNormalNodeLogger.Errorf("LinkByName error %v", err)
		}
		// set veth into namespace
		if err := netlink.LinkSetNsFd(veth, int(netNs)); err != nil {
			moduleNormalNodeLogger.Errorf("LinkSetNsFd error %v", err)
		}
		veths = append(veths, veth)
	}

	// 2. 进入容器命名空间启动 veth 设置 addr

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 切换到容器的网络命名空间
	if err = netns.Set(netNs); err != nil {
		moduleNormalNodeLogger.Errorf("Set netns error %v", err)
	}

	for _, veth := range veths {
		// 启动容器接口
		err = netlink.LinkSetUp(veth)
		if err != nil {
			moduleNormalNodeLogger.Errorf("LinkSetUp error %v", err)
		}

		// 设置 ip 地址
		ifName := veth.Attrs().Name
		addr := normalNode.IfNameToInterfaceMap[ifName].Addr
		ip, _ := netlink.ParseAddr(addr)
		if err = netlink.AddrAdd(veth, ip); err != nil {
			moduleNormalNodeLogger.Errorf("AddrAdd error %v", err)
		}
	}

}
