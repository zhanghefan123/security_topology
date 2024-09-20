package link

import (
	"github.com/vishvananda/netlink"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleAbstractEntities = logger.GetLogger(logger.ModuleAbstractEntities)

type AbstractLink struct {
	Type            types.NetworkLinkType  // 链路的类型
	Id              int                    // 链路 id
	SourceNodeType  types.NetworkNodeType  // 源节点类型
	TargetNodeType  types.NetworkNodeType  // 目的节点类型
	SourceInterface *intf.NetworkInterface // 源接口
	TargetInterface *intf.NetworkInterface // 目的接口
	SourceNode      interface{}            // 源节点
	TargetNode      interface{}            // 目的节点
}

func NewAbstractLink(typ types.NetworkLinkType, id int,
	sourceNodeType, targetNodeType types.NetworkNodeType,
	sourceIntf, targetIntf *intf.NetworkInterface,
	sourceNode, targetNode interface{}) *AbstractLink {
	return &AbstractLink{
		Type:            typ,
		Id:              id,
		SourceNodeType:  sourceNodeType,
		TargetNodeType:  targetNodeType,
		SourceInterface: sourceIntf,
		TargetInterface: targetIntf,
		SourceNode:      sourceNode,
		TargetNode:      targetNode,
	}
}

func (absLink *AbstractLink) GenerateVethPair() {
	vethPair := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: absLink.SourceInterface.IfName,
		},
		PeerName: absLink.TargetInterface.IfName,
	}
	err := netlink.LinkAdd(vethPair)
	if err != nil {
		moduleAbstractEntities.Errorf("error generating veth pair")
		return
	}
}
