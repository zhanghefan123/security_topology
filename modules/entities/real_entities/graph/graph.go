package graph

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
	"zhanghefan123/security_topology/services/http/params"

	"gonum.org/v1/gonum/graph/simple"
)

type Params struct {
	Source           string             `json:"source"`            // 源节点
	Destination      string             `json:"destination"`       // 目的节点
	SourceIndex      int                `json:"source_index"`      // 源索引
	DestinationIndex int                `json:"destination_index"` // 目的索引
	NumberOfPaths    int                // 需要计算的路径数
	LimitOfCost      float64            `json:"limit_of_cost"` // 开销的阈值
	Nodes            []params.NodeParam `json:"nodes"`         // 所有的节点
	Links            []params.LinkParam `json:"links"`         // 所有的链路
}

type Graph struct {
	PathsFilePath       string
	TopologyFilePath    string
	GraphParams         *Params
	DirectedGraph       *simple.DirectedGraph
	NameToNodeMapping   map[string]*entities.Node
	KShortestPaths      []*entities.Path
	RelationshipMapping map[string][]string
}

func CreateGraph(topologyFilePath, pathsFilePath string) *Graph {
	return &Graph{
		PathsFilePath:       pathsFilePath,
		TopologyFilePath:    topologyFilePath,
		GraphParams:         &Params{},
		DirectedGraph:       simple.NewDirectedGraph(),
		NameToNodeMapping:   make(map[string]*entities.Node),
		KShortestPaths:      make([]*entities.Path, 0),
		RelationshipMapping: make(map[string][]string),
	}
}
