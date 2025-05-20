package link

import (
	"context"
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"go.etcd.io/etcd/client/v3"
	"gonum.org/v1/gonum/graph/simple"
	"runtime"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	"zhanghefan123/security_topology/services/update/protobuf/link"
)

type AbstractLink struct {
	Type                types.NetworkLinkType  `json:"-"`              // 链路的类型
	Id                  int                    `json:"id"`             // 链路 id
	SourceNodeType      types.NetworkNodeType  `json:"-"`              // 源节点类型
	TargetNodeType      types.NetworkNodeType  `json:"-"`              // 目的节点类型
	SourceNodeId        int                    `json:"source_node_id"` // 源节点 id
	TargetNodeId        int                    `json:"target_node_id"` // 目的节点 id
	SourceContainerName string                 // 源容器的名称
	TargetContainerName string                 // 目的容器的名称
	SourceInterface     *intf.NetworkInterface `json:"-"` // 源接口
	TargetInterface     *intf.NetworkInterface `json:"-"` // 目的接口
	SourceNode          *node.AbstractNode     `json:"-"` // 源节点
	TargetNode          *node.AbstractNode     `json:"-"` // 目的节点
	BandWidth           int                    `json:"-"` // 带宽
	Status              bool                   `json:"-"` // 状态
	NetworkSegmentIPv4  iplib.Net4             `json:"-"` // 网络段
	NetworkSegmentIPv6  iplib.Net6             `json:"-"` // 网络段
}

func NewAbstractLink(typ types.NetworkLinkType, id int,
	sourceNodeType, targetNodeType types.NetworkNodeType,
	sourceNodeId, targetNodeId int,
	sourceContainerName, targetContainerName string,
	sourceIntf, targetIntf *intf.NetworkInterface,
	sourceNode, targetNode *node.AbstractNode,
	bandWidth int,
	graphTmp *simple.DirectedGraph,
	NetworkSegmentIPv4 iplib.Net4, NetworkSegmentIPv6 iplib.Net6,
) *AbstractLink {

	// 进行星座拓扑的边的添加 (注意这是有向图, 需要进行双向的链路的添加)
	// 在这里进行了双向的链路的添加 -> 为的是方便 LiR
	// ---------------------------------------------------------
	if types.NetworkLinkType_GroundSatelliteLink != typ {
		orderEdge := graphTmp.NewEdge(sourceNode, targetNode)
		graphTmp.SetEdge(orderEdge)
		reverseOrderEdge := graphTmp.NewEdge(targetNode, sourceNode)
		graphTmp.SetEdge(reverseOrderEdge)
	}
	// ---------------------------------------------------------

	return &AbstractLink{
		Type:                typ,
		Id:                  id,
		SourceNodeType:      sourceNodeType,
		TargetNodeType:      targetNodeType,
		SourceNodeId:        sourceNodeId,
		TargetNodeId:        targetNodeId,
		SourceContainerName: sourceContainerName,
		TargetContainerName: targetContainerName,
		SourceInterface:     sourceIntf,
		TargetInterface:     targetIntf,
		SourceNode:          sourceNode,
		TargetNode:          targetNode,
		BandWidth:           bandWidth,
		NetworkSegmentIPv4:  NetworkSegmentIPv4,
		NetworkSegmentIPv6:  NetworkSegmentIPv6,
	}
}

// GenerateVethPair 生成 veth pair
func (absLink *AbstractLink) GenerateVethPair() error {
	vethPair := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: absLink.SourceInterface.IfName,
		},
		PeerName: absLink.TargetInterface.IfName,
	}
	err := netlink.LinkAdd(vethPair)
	if err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	} else {
		return nil
	}
}

// RemoveVethPair 进行 veth pair 的删除
func (absLink *AbstractLink) RemoveVethPair() error {
	// 1. 获取环境的 namespace, 最终需要回到原始的环境
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

	// 获取 normalNodes
	sourceNormalNode, err := absLink.SourceNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("RemoveVethPair: failed to get normal node: %w", err)
	}
	// 获取源目 pid
	sourcePid := sourceNormalNode.Pid

	sourceNetNs, err := netns.GetFromPid(sourcePid)
	defer func(netNs *netns.NsHandle) {
		nsCloseErr := netNs.Close()
		if err == nil {
			err = nsCloseErr
		}
	}(&sourceNetNs)
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}

	// 6. 切换到源命名空间进行 veth 的删除 -> 其实相当于删除了对侧的veth
	runtime.LockOSThread()
	if err = netns.Set(sourceNetNs); err != nil {
		return fmt.Errorf("netns.Set(sourceNetNs) failed: %v", err)
	}
	err = netlink.LinkDel(*absLink.SourceInterface.Veth)
	if err != nil {
		return fmt.Errorf("failed to delete veth pair: %w", err)
	}
	runtime.UnlockOSThread()

	return nil
}

// SetVethNamespaceAndAddr 进行 veth 命名空间以及地址的设置
func (absLink *AbstractLink) SetVethNamespaceAndAddr() error {
	// 1. 获取环境的 namespace, 最终需要回到原始的环境
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

	// 2. 获取源和目的的普通节点
	sourceNormalNode, err := absLink.SourceNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get source normal node failed: %v", err)
	}
	targetNormalNode, err := absLink.TargetNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get target normal node failed: %v", err)
	}

	// 3. 获取源目 pid
	sourcePid := sourceNormalNode.Pid
	targetPid := targetNormalNode.Pid

	// 4. 获取源目 netns
	// ---------------------------------------------------------------------
	sourceNetNs, err := netns.GetFromPid(sourcePid)
	defer func(netNs *netns.NsHandle) {
		nsCloseErr := netNs.Close()
		if err == nil {
			err = nsCloseErr
		}
	}(&sourceNetNs)
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}

	targetNetNs, err := netns.GetFromPid(targetPid)
	defer func(netNs *netns.NsHandle) {
		nsCloseErr := netNs.Close()
		if err == nil {
			err = nsCloseErr
		}
	}(&targetNetNs)
	if err != nil {
		return fmt.Errorf("netns.Get() failed: %w", err)
	}
	// ---------------------------------------------------------------------

	// 5. 将源接口设置到命名空间之中
	// ---------------------------------------------------------------------
	// 5.1 find veth by name
	sourceVeth, err := netlink.LinkByName(absLink.SourceInterface.IfName)
	if err != nil {
		return fmt.Errorf("get link by name failed: %v", err)
	}
	// 5.2 set veth into namespace
	if err = netlink.LinkSetNsFd(sourceVeth, int(sourceNetNs)); err != nil {
		return fmt.Errorf("netlink.LinkSetNsFd(sourceVeth, int(sourceNetNs)) failed: %v", err)
	}
	// 5.3 set into interface
	absLink.SourceInterface.Veth = &sourceVeth
	// ---------------------------------------------------------------------

	// 6. 将目的接口设置到命名空间
	// ---------------------------------------------------------------------
	// 6.1 find veth by name
	targetVeth, err := netlink.LinkByName(absLink.TargetInterface.IfName)
	if err != nil {
		return fmt.Errorf("get link by name failed: %v", err)
	}
	// 6.2 set veth into namespace
	if err = netlink.LinkSetNsFd(targetVeth, int(targetNetNs)); err != nil {
		return fmt.Errorf("netlink.LinkSetNsFd(targetVeth, int(targetNetNs)) failed: %v", err)
	}
	// 6.3 set into interface
	absLink.TargetInterface.Veth = &targetVeth
	// ---------------------------------------------------------------------

	runtime.LockOSThread()

	// 7. 切换到源节点的网络命名空间 (启动接口, 设置 ipv4 地址, 设置 ipv6 地址)
	// ---------------------------------------------------------------------
	if err = netns.Set(sourceNetNs); err != nil {
		return fmt.Errorf("netns.Set(sourceNetNs) failed: %v", err)
	}
	err = netlink.LinkSetUp(sourceVeth)
	if err != nil {
		return fmt.Errorf("netlink.LinkSetUp(sourceVeth) failed: %v", err)
	}
	ifName := sourceVeth.Attrs().Name

	// 8. 设置 ipv4 地址
	ipv4Addr := sourceNormalNode.IfNameToInterfaceMap[ifName].SourceIpv4Addr
	ipv4, _ := netlink.ParseAddr(ipv4Addr)
	if err = netlink.AddrAdd(sourceVeth, ipv4); err != nil {
		return fmt.Errorf("netlink.AddrAdd(ipv4) failed: %v", err)
	}

	// 9. 设置 ipv6 地址
	ipv6Addr := sourceNormalNode.IfNameToInterfaceMap[ifName].SourceIpv6Addr
	ipv6, _ := netlink.ParseAddr(ipv6Addr)
	if err = netlink.AddrAdd(sourceVeth, ipv6); err != nil {
		fmt.Printf("netlink.AddrAdd(%s) failed: %v", ipv6, err)
		return fmt.Errorf("netlink.AddrAdd(%s) failed: %w", ipv6, err)
	}
	// ---------------------------------------------------------------------

	runtime.UnlockOSThread()

	runtime.LockOSThread()

	// 10. 切换到目的节点的网络命名空间 (启动接口, 设置 ipv4 地址, 设置 ipv6 地址)
	// ---------------------------------------------------------------------
	if err = netns.Set(targetNetNs); err != nil {
		return fmt.Errorf("netns.Set(targetNetNs) failed: %v", err)
	}
	err = netlink.LinkSetUp(targetVeth)
	if err != nil {
		return fmt.Errorf("netlink.LinkSetUp(targetVeth) failed: %v", err)
	}
	ifName = targetVeth.Attrs().Name

	ipv4Addr = targetNormalNode.IfNameToInterfaceMap[ifName].SourceIpv4Addr
	ipv4, _ = netlink.ParseAddr(ipv4Addr)
	if err = netlink.AddrAdd(targetVeth, ipv4); err != nil {
		return fmt.Errorf("netlink.AddrAdd(ipv4) failed: %v", err)
	}

	ipv6Addr = targetNormalNode.IfNameToInterfaceMap[ifName].SourceIpv6Addr
	ipv6, _ = netlink.ParseAddr(ipv6Addr)
	if err = netlink.AddrAdd(targetVeth, ipv6); err != nil {
		fmt.Printf("netlink.AddrAdd(%s) failed: %v", ipv6, err)
		return fmt.Errorf("netlink.AddrAdd(%s) failed: %w", ipv6, err)
	}
	// ---------------------------------------------------------------------

	runtime.UnlockOSThread()

	return nil
}

// StoreToEtcd 将链路信息存储到 Etcd 之中
func (absLink *AbstractLink) StoreToEtcd(etcdClient *clientv3.Client) error {
	linkPb := &link.Link{
		Type:            link.LinkType_LINK_TYPE_INTER_SATELLITE_LINK,
		Id:              int32(absLink.Id),
		SourceNodeId:    int32(absLink.SourceNodeId),
		TargetNodeId:    int32(absLink.TargetNodeId),
		SourceIfaceName: absLink.SourceInterface.IfName,
		TargetIfaceName: absLink.TargetInterface.IfName,
		Bandwidth:       0,
		Delay:           0,
	}
	linkInBytes := protobuf.MustMarshal(linkPb)
	linkPrefix := configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.ISLsPrefix
	linkKey := fmt.Sprintf("%s/%d", linkPrefix, absLink.Id)
	_, err := etcdClient.Put(context.Background(), linkKey, string(linkInBytes))
	if err != nil {
		return fmt.Errorf("failed to store link into etcd: %w", err)
	}
	return nil
}

// SetLinkParams 设置链路属性
func (absLink *AbstractLink) SetLinkParams() error {
	if absLink.BandWidth != linux_tc_api.LargeBandwidth {
		var err error
		var sourceNode, targetNode *normal_node.NormalNode
		sourceNode, err = absLink.SourceNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("failed to get source normal node: %w", err)
		}
		targetNode, err = absLink.TargetNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("failed to get target normal node: %w", err)
		}
		err = linux_tc_api.SetInterfaceBandwidth(absLink.SourceInterface, sourceNode.Pid, absLink.BandWidth)
		if err != nil {
			return fmt.Errorf("failed to set link params: %w", err)
		}
		err = linux_tc_api.SetInterfaceBandwidth(absLink.TargetInterface, targetNode.Pid, absLink.BandWidth)
		if err != nil {
			return fmt.Errorf("failed to set link params: %w", err)
		}
	}
	return nil
}
