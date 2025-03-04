package node

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
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
func NewAbstractNode(nodeType types.NetworkNodeType, actualNode interface{}, graphTmp *simple.DirectedGraph) *AbstractNode {
	// 进行图节点的创建
	var graphNode graph.Node
	// 当节点为 etcd_service 或者 position_service 的时候, graphTmp 为 nil
	if graphTmp == nil {
		graphNode = nil
	} else {
		graphNode = graphTmp.NewNode()
		graphTmp.AddNode(graphNode)
	}
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
	case types.NetworkNodeType_GroundStation:
		if groundStation, ok := abstractNode.ActualNode.(*ground_station.GroundStation); ok {
			return groundStation.NormalNode, nil
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
	case types.NetworkNodeType_LirNode:
		if lirNode, ok := abstractNode.ActualNode.(*nodes.LiRNode); ok {
			return lirNode.NormalNode, nil
		}
	case types.NetworkNodeType_Entrance:
		if entrance, ok := abstractNode.ActualNode.(*nodes.Entrance); ok {
			return entrance.NormalNode, nil
		}
	case types.NetworkNodeType_FabricPeerNode:
		if fabricPeerNode, ok := abstractNode.ActualNode.(*nodes.FabricPeerNode); ok {
			return fabricPeerNode.NormalNode, nil
		}
	case types.NetworkNodeType_FabricOrderNode:
		if fabricOrderNode, ok := abstractNode.ActualNode.(*nodes.FabricOrderNode); ok {
			return fabricOrderNode.NormalNode, nil
		}
	case types.NetworkNodeType_LiRSatellite:
		if lirSatelliteNode, ok := abstractNode.ActualNode.(*satellites.LiRSatellite); ok {
			return lirSatelliteNode.NormalNode, nil
		}
	}

	return nil, fmt.Errorf("cannot get normal node from abstract node")
}
