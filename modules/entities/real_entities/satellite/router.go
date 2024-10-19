package satellite

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type Router struct {
	*normal_node.NormalNode
}

// NewRouter 创建新的路由器
func NewRouter(nodeId int) *Router {
	routerType := types.NetworkNodeType_Router
	return &Router{
		NormalNode: normal_node.NewNormalNode(types.NetworkNodeStatus_Logic, nodeId, 1,
			fmt.Sprintf("%s-%d", routerType.String(), nodeId)),
	}
}
