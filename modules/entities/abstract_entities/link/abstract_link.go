package link

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/vishvananda/netlink"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	"zhanghefan123/security_topology/services/position/protobuf/link"
)

type AbstractLink struct {
	Type            types.NetworkLinkType  `json:"-"`              // 链路的类型
	Id              int                    `json:"id"`             // 链路 id
	SourceNodeType  types.NetworkNodeType  `json:"-"`              // 源节点类型
	TargetNodeType  types.NetworkNodeType  `json:"-"`              // 目的节点类型
	SourceNodeId    int                    `json:"source_node_id"` // 源节点 id
	TargetNodeId    int                    `json:"target_node_id"` // 目的节点 id
	SourceInterface *intf.NetworkInterface `json:"-"`              // 源接口
	TargetInterface *intf.NetworkInterface `json:"-"`              // 目的接口
	SourceNode      interface{}            `json:"-"`              // 源节点
	TargetNode      interface{}            `json:"-"`              // 目的节点
}

func NewAbstractLink(typ types.NetworkLinkType, id int,
	sourceNodeType, targetNodeType types.NetworkNodeType,
	sourceNodeId, targetNodeId int,
	sourceIntf, targetIntf *intf.NetworkInterface,
	sourceNode, targetNode interface{}) *AbstractLink {
	return &AbstractLink{
		Type:            typ,
		Id:              id,
		SourceNodeType:  sourceNodeType,
		TargetNodeType:  targetNodeType,
		SourceNodeId:    sourceNodeId,
		TargetNodeId:    targetNodeId,
		SourceInterface: sourceIntf,
		TargetInterface: targetIntf,
		SourceNode:      sourceNode,
		TargetNode:      targetNode,
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
