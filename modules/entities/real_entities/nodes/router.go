package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type Router struct {
	*normal_node.NormalNode
}

// NewRouter 创建新的路由器
func NewRouter(nodeId int, X, Y float64) *Router {
	routerType := types.NetworkNodeType_Router
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_Router,
		nodeId,
		fmt.Sprintf("%s-%d", routerType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &Router{
		NormalNode: normalNode,
	}
}
