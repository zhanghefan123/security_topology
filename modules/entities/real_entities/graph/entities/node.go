package entities

import (
	"gonum.org/v1/gonum/graph"
)

type Node struct {
	graph.Node
	NodeName    string
	ExcessValue int     // 存储节点的入度信息
	Distance    float64 // 从源节点到这个节点的举例
	Indegree    int     // 节点的入度
	Outdegree   int     // 节点的出度
	Index       int     // 节点的编号
}

func CreateGraphNode(nodeName string, nodeIndex int) *Node {
	return &Node{
		Node:        nil,
		NodeName:    nodeName,
		ExcessValue: 0,
		Index:       nodeIndex,
	}
}
