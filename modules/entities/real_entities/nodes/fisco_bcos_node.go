package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type FiscoBcosNode struct {
	*normal_node.NormalNode
}

// NewFiscoBcosNode 创建新的 FiscoBcos
func NewFiscoBcosNode(nodeId int, X, Y float64) *FiscoBcosNode {
	fiscoBcosNodeType := types.NetworkNodeType_FiscoBcosNode
	normalNode := normal_node.NewNormalNode(
		fiscoBcosNodeType,
		nodeId,
		fmt.Sprintf("%s-%d", fiscoBcosNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &FiscoBcosNode{normalNode}
}
