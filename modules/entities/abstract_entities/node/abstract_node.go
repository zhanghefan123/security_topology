package node

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/position"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

// AbstractNode 抽象节点
type AbstractNode struct {
	graph.Node                       // 用来计算路由的
	Type       types.NetworkNodeType // 节点类型
	ActualNode interface{}           // 实际的节点
}

// NewAbstractNode 创建新的抽象节点
func NewAbstractNode(nodeType types.NetworkNodeType, actualNode interface{}) *AbstractNode {
	graphNode := configs.ConstellationGraph.NewNode()
	configs.ConstellationGraph.AddNode(graphNode)
	return &AbstractNode{
		Node:       graphNode,
		Type:       nodeType,
		ActualNode: actualNode,
	}
}

// GetNormalNodeFromAbstractNode 从抽象节点之中进行普通节点的获取
func (abstractNode *AbstractNode) GetNormalNodeFromAbstractNode() (*normal_node.NormalNode, error) {
	if abstractNode.Type == types.NetworkNodeType_NormalSatellite {
		if normalSat, ok := abstractNode.ActualNode.(*satellite.NormalSatellite); ok {
			return normalSat.NormalNode, nil
		}
	} else if abstractNode.Type == types.NetworkNodeType_ConsensusSatellite {
		if consensusSat, ok := abstractNode.ActualNode.(*satellite.ConsensusSatellite); ok {
			return consensusSat.NormalNode, nil
		}
	} else if abstractNode.Type == types.NetworkNodeType_EtcdService {
		if etcdService, ok := abstractNode.ActualNode.(*etcd.EtcdNode); ok {
			return etcdService.NormalNode, nil
		}
	} else if abstractNode.Type == types.NetworkNodeType_PositionService {
		if positionService, ok := abstractNode.ActualNode.(*position.PositionService); ok {
			return positionService.NormalNode, nil
		}
	}
	return nil, fmt.Errorf("cannot get normal node from abstractn node")
}
