package link

import (
	"context"
	"fmt"
	"github.com/vishvananda/netlink"
	"go.etcd.io/etcd/client/v3"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	"zhanghefan123/security_topology/services/position/protobuf/link"
)

type AbstractLink struct {
	Type                types.NetworkLinkType `json:"-"`              // 链路的类型
	Id                  int                   `json:"id"`             // 链路 id
	SourceNodeType      types.NetworkNodeType `json:"-"`              // 源节点类型
	TargetNodeType      types.NetworkNodeType `json:"-"`              // 目的节点类型
	SourceNodeId        int                   `json:"source_node_id"` // 源节点 id
	TargetNodeId        int                   `json:"target_node_id"` // 目的节点 id
	SourceContainerName string
	TargetContainerName string
	SourceInterface     *intf.NetworkInterface `json:"-"` // 源接口
	TargetInterface     *intf.NetworkInterface `json:"-"` // 目的接口
	SourceNode          *node.AbstractNode     `json:"-"` // 源节点
	TargetNode          *node.AbstractNode     `json:"-"` // 目的节点
	BandWidth           int                    `json:"-"` // 带宽
}

func NewAbstractLink(typ types.NetworkLinkType, id int,
	sourceNodeType, targetNodeType types.NetworkNodeType,
	sourceNodeId, targetNodeId int,
	sourceContainerName, targetContainerName string,
	sourceIntf, targetIntf *intf.NetworkInterface,
	sourceNode, targetNode *node.AbstractNode,
	bandWidth int,
	graphTmp *simple.DirectedGraph) *AbstractLink {

	// 进行星座拓扑的边的添加 (注意这是有向图, 需要进行双向的链路的添加)
	// 在这里进行了双向的链路的添加
	orderEdge := graphTmp.NewEdge(sourceNode, targetNode)
	graphTmp.SetEdge(orderEdge)
	reverseOrderEdge := graphTmp.NewEdge(targetNode, sourceNode)
	graphTmp.SetEdge(reverseOrderEdge)

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
	return nil
}
