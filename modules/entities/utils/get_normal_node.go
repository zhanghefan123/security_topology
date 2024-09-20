package utils

import (
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

func GetNormalNodeFromAbstractNode(node *node.AbstractNode) *normal_node.NormalNode {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		if normalSat, ok := node.ActualNode.(*satellite.NormalSatellite); ok {
			return normalSat.NormalNode
		}
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		if consensusSat, ok := node.ActualNode.(*satellite.ConsensusSatellite); ok {
			return consensusSat.NormalNode
		}
	}
	return nil
}
