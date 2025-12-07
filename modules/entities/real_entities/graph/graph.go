package graph

import (
	"zhanghefan123/security_topology/services/http/params"

	"gonum.org/v1/gonum/graph/simple"
)

type Params struct {
	Source        string             `json:"source"`      // 源节点
	Destination   string             `json:"destination"` // 目的节点
	NumberOfPaths int                // 需要计算的路径数
	LimitOfCost   float64            `json:"limit_of_cost"` // 开销的阈值
	Nodes         []params.NodeParam `json:"nodes"`         // 所有的节点
	Links         []params.LinkParam `json:"links"`         // 所有的链路
}

type Graph struct {
	TopologyFilePath  string
	GraphParams       *Params
	DirectedGraph     *simple.DirectedGraph
	NameToNodeMapping map[string]*Node
	KShortestPaths    []*Path
}

func CreateGraph(topologyFilePath string) *Graph {
	return &Graph{
		TopologyFilePath:  topologyFilePath,
		GraphParams:       &Params{},
		DirectedGraph:     simple.NewDirectedGraph(),
		NameToNodeMapping: make(map[string]*Node),
		KShortestPaths:    make([]*Path, 0),
	}
}
