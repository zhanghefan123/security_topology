package utils

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

var (
	ErrCouldNotGetNormalNode = fmt.Errorf("could not get normal node")
)

// GetNormalNodeFromAbstractNode 从抽象节点之中获取普通节点
func GetNormalNodeFromAbstractNode(node *node.AbstractNode) (*normal_node.NormalNode, error) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		if normalSat, ok := node.ActualNode.(*satellite.NormalSatellite); ok {
			return normalSat.NormalNode, nil
		}
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		if consensusSat, ok := node.ActualNode.(*satellite.ConsensusSatellite); ok {
			return consensusSat.NormalNode, nil
		}
	} else if node.Type == types.NetworkNodeType_EtcdNode {
		if etcdNode, ok := node.ActualNode.(*etcd.EtcdNode); ok {
			return etcdNode.NormalNode, nil
		}
	}
	return nil, ErrCouldNotGetNormalNode
}
