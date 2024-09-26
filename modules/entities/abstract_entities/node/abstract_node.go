package node

import (
	"zhanghefan123/security_topology/modules/entities/types"
)

type AbstractNode struct {
	Type       types.NetworkNodeType // 节点类型
	ActualNode interface{}           // 实际的节点
}

// NewAbstractNode 创建新的抽象节点
func NewAbstractNode(nodeType types.NetworkNodeType, actualNode interface{}) *AbstractNode {
	return &AbstractNode{
		Type:       nodeType,
		ActualNode: actualNode,
	}
}
