package graph

import (
	"gonum.org/v1/gonum/graph"
)

var (
	MultipathGraphNodeMapping = map[string]*MultipathGraphNode{}
)

type MultipathGraphNode struct {
	graph.Node
	NodeName    string
	ExcessValue int // 存储节点的入度信息
}

func CreateMultipathGraphNode(nodeName string) *MultipathGraphNode {
	return &MultipathGraphNode{
		Node:        nil,
		NodeName:    nodeName,
		ExcessValue: 0,
	}
}
