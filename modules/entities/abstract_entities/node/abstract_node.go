package node

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/position"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

// AbstractNode 抽象节点
type AbstractNode struct {
	Type       types.NetworkNodeType // 节点类型
	ActualNode interface{}           // 实际的节点
}

// NewAbstractNode 创建新的抽象节点
func NewAbstractNode(nodeType types.NetworkNodeType, actualNode interface{}) *AbstractNode {
	return &AbstractNode{
		Type:       nodeType,
		ActualNode: actualNode,
	}
}

// GetNormalNodeFromAbstractNode 从抽象节点之中进行普通节点的获取
func (node *AbstractNode) GetNormalNodeFromAbstractNode() (*normal_node.NormalNode, error) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		if normalSat, ok := node.ActualNode.(*satellite.NormalSatellite); ok {
			return normalSat.NormalNode, nil
		}
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		if consensusSat, ok := node.ActualNode.(*satellite.ConsensusSatellite); ok {
			return consensusSat.NormalNode, nil
		}
	} else if node.Type == types.NetworkNodeType_EtcdService {
		if etcdService, ok := node.ActualNode.(*etcd.EtcdNode); ok {
			return etcdService.NormalNode, nil
		}
	} else if node.Type == types.NetworkNodeType_PositionService {
		if positionService, ok := node.ActualNode.(*position.PositionService); ok {
			return positionService.NormalNode, nil
		}
	}
	return nil, fmt.Errorf("cannot get normal node from abstractn node")
}
