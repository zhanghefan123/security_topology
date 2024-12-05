package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type Entrance struct {
	*normal_node.NormalNode
}

func NewEntrance(nodeId int, X, Y float64) *Entrance {
	entranceType := types.NetworkNodeType_Entrance
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_Entrance,
		nodeId,
		fmt.Sprintf("%s-%d", entranceType.String(), nodeId))
	normalNode.X = X
	normalNode.Y = Y
	return &Entrance{normalNode}
}
