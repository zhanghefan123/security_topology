package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type ConsensusNode struct {
	*normal_node.NormalNode
}

func NewConsensusNode(nodeId int, X, Y float64) *ConsensusNode {
	consensusNodeType := types.NetworkNodeType_ConsensusNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_ConsensusNode,
		nodeId,
		fmt.Sprintf("%s-%d", consensusNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &ConsensusNode{
		NormalNode: normalNode,
	}
}
