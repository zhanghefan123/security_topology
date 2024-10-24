package node

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
)

// AbstractNode 抽象节点
type AbstractNode struct {
	graph.Node
	Type       types.NetworkNodeType // 节点类型
	ActualNode interface{}           // 实际的节点
}

// NewAbstractNode 创建新的抽象节点
func NewAbstractNode(nodeType types.NetworkNodeType, actualNode interface{}) *AbstractNode {
	// 进行图节点的创建
	graphNode := configs.ConstellationGraph.NewNode()
	// 进行图节点的添加
	configs.ConstellationGraph.AddNode(graphNode)
	return &AbstractNode{
		Node:       graphNode,
		Type:       nodeType,
		ActualNode: actualNode,
	}
}

// GetNormalNodeFromAbstractNode 从抽象节点之中进行普通节点的获取
func (abstractNode *AbstractNode) GetNormalNodeFromAbstractNode() (*normal_node.NormalNode, error) {
	switch abstractNode.Type {
	case types.NetworkNodeType_NormalSatellite:
		if normalSat, ok := abstractNode.ActualNode.(*satellites.NormalSatellite); ok {
			return normalSat.NormalNode, nil
		}
	case types.NetworkNodeType_ConsensusSatellite:
		if consensusSat, ok := abstractNode.ActualNode.(*satellites.ConsensusSatellite); ok {
			return consensusSat.NormalNode, nil
		}
	case types.NetworkNodeType_EtcdService:
		if etcdService, ok := abstractNode.ActualNode.(*etcd.EtcdNode); ok {
			return etcdService.NormalNode, nil
		}
	case types.NetworkNodeType_PositionService:
		if positionService, ok := abstractNode.ActualNode.(*position.PositionService); ok {
			return positionService.NormalNode, nil
		}
	case types.NetworkNodeType_Router:
		if routerTmp, ok := abstractNode.ActualNode.(*nodes.Router); ok {
			return routerTmp.NormalNode, nil
		}
	case types.NetworkNodeType_NormalNode:
		if normalNode, ok := abstractNode.ActualNode.(*normal_node.NormalNode); ok {
			return normalNode, nil
		}
	case types.NetworkNodeType_ConsensusNode:
		if consensusNode, ok := abstractNode.ActualNode.(*nodes.ConsensusNode); ok {
			return consensusNode.NormalNode, nil
		}
	case types.NetworkNodeType_ChainMakerNode:
		if chainMakerNode, ok := abstractNode.ActualNode.(*nodes.ChainmakerNode); ok {
			return chainMakerNode.NormalNode, nil
		}
	case types.NetworkNodeType_MaliciousNode:
		if maliciousNode, ok := abstractNode.ActualNode.(*nodes.MaliciousNode); ok {
			return maliciousNode.NormalNode, nil
		}
	}
	return nil, fmt.Errorf("cannot get normal node from abstract node")
}
