package entities

import (
	"encoding/json"
	"fmt"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"strings"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/corrupt_decider"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/params"
	"zhanghefan123/security_topology/modules/entities/types"
)

type SimGraph struct {
	DirectedGraph                 *simple.DirectedGraph
	GraphParams                   *params.GraphParams
	SourceNode                    *SimAbstractNode
	DestinationNode               *SimAbstractNode
	SimAbstractNodes              []*SimAbstractNode
	SimDirectedNormalLinks        []*SimDirectedNormalLink
	SimDirectedPvLinks            []*SimDirectedPvLink
	SimAbstractNodesMapping       map[string]*SimAbstractNode
	SimDirectedNormalLinksMapping map[string]map[string]*SimDirectedNormalLink
	SimDirectedPvLinksMapping     map[string]map[string]*SimDirectedPvLink

	AvailablePaths       []*SimPath
	CoveragePaths        []*SimPath          // 覆盖这张图所有边的路径
	AvailablePathMapping map[string]*SimPath // 从描述到实际对象的 mapping
	CoveragePathMapping  map[string]*SimPath // 从描述到实际对象的 mapping

	TotalPathWeights float64 // 所有路径的总的权重
}

func NewSimGraph() *SimGraph {
	return &SimGraph{
		DirectedGraph:                 &simple.DirectedGraph{},
		GraphParams:                   &params.GraphParams{},
		SimAbstractNodes:              make([]*SimAbstractNode, 0),
		SimDirectedNormalLinks:        make([]*SimDirectedNormalLink, 0),
		SimDirectedPvLinks:            make([]*SimDirectedPvLink, 0),
		SimAbstractNodesMapping:       make(map[string]*SimAbstractNode),
		SimDirectedNormalLinksMapping: make(map[string]map[string]*SimDirectedNormalLink),
		SimDirectedPvLinksMapping:     make(map[string]map[string]*SimDirectedPvLink),

		AvailablePaths:       make([]*SimPath, 0),
		CoveragePaths:        make([]*SimPath, 0),
		AvailablePathMapping: make(map[string]*SimPath),
		CoveragePathMapping:  make(map[string]*SimPath),

		TotalPathWeights: 0,
	}
}

// IsCoveragePath 判定某条路径是否是覆盖集之中的路径
func (simGraph *SimGraph) IsCoveragePath(path *SimPath) bool {
	if _, ok := simGraph.CoveragePathMapping[path.PathDescription]; ok {
		return true
	} else {
		return false
	}
}

// ClearAllEdgesProbabilities 清除图中所有边的概率信息
func (simGraph *SimGraph) ClearAllEdgesProbabilities() {
	for _, directedPvLink := range simGraph.SimDirectedPvLinks {
		directedPvLink.ExploreProbability = 0
	}
}

// LoadGraphParamsFromConfigurationFile 从配置文件中加载图的参数信息
func (simGraph *SimGraph) LoadGraphParamsFromConfigurationFile(simulationGraphPath string) error {
	data, err := os.ReadFile(simulationGraphPath)
	if err != nil {
		return fmt.Errorf("read topology file failed, %w", err)
	}
	err = json.Unmarshal(data, simGraph.GraphParams)
	if err != nil {
		return fmt.Errorf("unmarshal binary data failed, %w", err)
	}
	return nil
}

// LoadNodesFromNodeParams 从 GraphParams 中的节点参数信息中加载图中的节点
func (simGraph *SimGraph) LoadNodesFromNodeParams() error {
	for _, nodeParam := range simGraph.GraphParams.Nodes {
		nodeType, err := params.ResolveSimNodeType(nodeParam.Type)
		if err != nil {
			return fmt.Errorf("resolve node type failed, %w", err)
		}
		nodeName, err := params.ResolveSimNodeName(&nodeParam)
		if err != nil {
			return fmt.Errorf("resolve node name failed, %w", err)
		}

		if nodeType == types.SimNetworkNodeType_NormalRouter {
			// create actual node
			var uniformCorruptDecider *corrupt_decider.CorruptDecider
			uniformCorruptDecider, err = corrupt_decider.CreateUniformCorruptDecider(0.1, 0.5)
			newSimNormalRouter := NewSimNormalRouter(nodeName, nodeParam.Index, uniformCorruptDecider)
			// create sim abstract node
			graphNode := simGraph.DirectedGraph.NewNode()
			simAbstractNode := NewSimAbstract(nodeType, newSimNormalRouter, graphNode)
			simGraph.SimAbstractNodes = append(simGraph.SimAbstractNodes, simAbstractNode)
			// add node to the grapha
			simGraph.DirectedGraph.AddNode(simAbstractNode)
		} else if nodeType == types.SimNetworkNodeType_PathValidationRouter {
			// create actual node
			newSimPathValidationRouter := NewSimPathValidationRouter(nodeName, nodeParam.Index)
			// create sim abstract node
			graphNode := simGraph.DirectedGraph.NewNode()
			simAbstractNode := NewSimAbstract(nodeType, newSimPathValidationRouter, graphNode)
			simGraph.SimAbstractNodes = append(simGraph.SimAbstractNodes, simAbstractNode)
			// add node to the graph
			simGraph.DirectedGraph.AddNode(simAbstractNode)
		} else {
			return fmt.Errorf("unsupported node type: %s", nodeType.String())
		}
	}
	return nil
}

// LoadSourceAndDest 从配置文件中加载源和目的节点的信息
func (simGraph *SimGraph) LoadSourceAndDest() error {
	if sourceNode, ok := simGraph.SimAbstractNodesMapping[simGraph.GraphParams.SourceDestParams.Source]; !ok {
		return fmt.Errorf("load source failed")
	} else {
		simGraph.SourceNode = sourceNode
	}
	if destinationNode, ok := simGraph.SimAbstractNodesMapping[simGraph.GraphParams.SourceDestParams.Destination]; !ok {
		return fmt.Errorf("load destination failed")
	} else {
		simGraph.DestinationNode = destinationNode
	}
	return nil
}

// LoadLinksFromLinkParams 从 GraphParams 中的链路参数信息中加载图中的链路
func (simGraph *SimGraph) LoadLinksFromLinkParams() error {
	for _, linkParam := range simGraph.GraphParams.Links {
		// get source name
		sourceNodeName, err := params.ResolveSimNodeName(&linkParam.SourceNode)
		if err != nil {
			return fmt.Errorf("resolve source node name failed, %w", err)
		}
		targetNodeName, err := params.ResolveSimNodeName(&linkParam.TargetNode)
		if err != nil {
			return fmt.Errorf("resolve target node name failed, %w", err)
		}
		// get abstract sim node
		sourceSimAbstractNode, ok := simGraph.SimAbstractNodesMapping[sourceNodeName]
		if !ok {
			return fmt.Errorf("cannot find source sim abstract node, name: %s", sourceNodeName)
		}
		targetSimAbstractNode, ok := simGraph.SimAbstractNodesMapping[targetNodeName]
		if !ok {
			return fmt.Errorf("cannot find target sim abstract node, name: %s", targetNodeName)
		}
		// construct edges
		isPvLink := (sourceSimAbstractNode.Type == types.SimNetworkNodeType_PathValidationRouter) || (targetSimAbstractNode.Type == types.SimNetworkNodeType_PathValidationRouter)
		if isPvLink {
			pvLink := NewSimDirectedPvLink(sourceSimAbstractNode, targetSimAbstractNode)
			directGraphEdge := simGraph.DirectedGraph.NewEdge(sourceSimAbstractNode, targetSimAbstractNode)
			simGraph.DirectedGraph.SetEdge(directGraphEdge)
			simGraph.SimDirectedPvLinks = append(simGraph.SimDirectedPvLinks, pvLink)
			// update mapping
			if _, ok = simGraph.SimDirectedPvLinksMapping[sourceNodeName]; !ok {
				simGraph.SimDirectedPvLinksMapping[sourceNodeName] = make(map[string]*SimDirectedPvLink)
			} else {
				simGraph.SimDirectedPvLinksMapping[sourceNodeName][targetNodeName] = pvLink
			}
		} else {
			normalLink := NewSimDirectedNormalLink(sourceSimAbstractNode, targetSimAbstractNode)
			directGraphEdge := simGraph.DirectedGraph.NewEdge(sourceSimAbstractNode, targetSimAbstractNode)
			simGraph.DirectedGraph.SetEdge(directGraphEdge)
			simGraph.SimDirectedNormalLinks = append(simGraph.SimDirectedNormalLinks, normalLink)
			// update mapping
			if _, ok = simGraph.SimDirectedNormalLinksMapping[sourceNodeName]; !ok {
				simGraph.SimDirectedNormalLinksMapping[sourceNodeName] = make(map[string]*SimDirectedNormalLink)
			} else {
				simGraph.SimDirectedNormalLinksMapping[sourceNodeName][targetNodeName] = normalLink
			}
		}
	}
	return nil
}

func FindNextPvRouterAndIndex(simPath *SimPath, startIndex int) (*SimAbstractNode, int) {
	for index := startIndex + 1; index < len(simPath.NodeList); index++ {
		if simPath.NodeList[index].Type == types.SimNetworkNodeType_PathValidationRouter {
			return simPath.NodeList[index], index
		}
	}
	return nil, -1
}

func (simGraph *SimGraph) LoadCoveragePathsFromParams() error {
	for _, coveragePathString := range simGraph.GraphParams.CoveragePaths {
		// fill node list in coverage path
		coveragePath := NewSimPath()
		routerListString := strings.Split(coveragePathString, ",")
		for _, routerName := range routerListString {
			if router, ok := simGraph.SimAbstractNodesMapping[routerName]; !ok {
				return fmt.Errorf("cannot find sim abstract node for router name: %s", routerName)
			} else {
				coveragePath.NodeList = append(coveragePath.NodeList, router)
			}
		}
		// fill directedPvLinks and directedPvLinksMapping in coverage path
		var startIndex = 0
		var sourceNode, targetNode *SimAbstractNode
		for {
			sourceNode = coveragePath.NodeList[startIndex]
			targetNode, startIndex = FindNextPvRouterAndIndex(coveragePath, startIndex)
			if targetNode == nil {
				break
			} else {
				sourceNodeName, err := sourceNode.GetSimNodeName()
				if err != nil {
					return fmt.Errorf("get directed pv links failed due to: %v", err)
				}
				targetNodeName, err := targetNode.GetSimNodeName()
				if err != nil {
					return fmt.Errorf("get directed pv links failed due to: %v", err)
				}
				directedPvLink := simGraph.SimDirectedPvLinksMapping[sourceNodeName][targetNodeName]
				coveragePath.DirectedPvLinks = append(coveragePath.DirectedPvLinks, directedPvLink)
				if _, ok := coveragePath.DirectedPvLinksMapping[sourceNodeName]; !ok {
					coveragePath.DirectedPvLinksMapping[sourceNodeName] = make(map[string]*SimDirectedPvLink)
				}
				coveragePath.DirectedPvLinksMapping[sourceNodeName][targetNodeName] = directedPvLink
			}
		}
	}
	return nil
}

func (simGraph *SimGraph) CalculateKShortestPaths() error {
	paths := path.YenKShortestPaths(simGraph.DirectedGraph, simGraph.GraphParams.KShortestPathParmas.NumberOfPaths, simGraph.GraphParams.KShortestPathParmas.LimitOfCost, simGraph.SourceNode, simGraph.DestinationNode)
	for _, graphNodeList := range paths {
		singleSimPath := NewSimPath()
		for _, abstractSimNode := range graphNodeList {
			singleModifiedNode, ok := abstractSimNode.(*SimAbstractNode)
			if ok {
				singleSimPath.NodeList = append(singleSimPath.NodeList, singleModifiedNode)
			} else {
				return fmt.Errorf("type assertion failed when calculating k shortest path")
			}
		}
		simPathDesc, err := singleSimPath.GetPathDescription()
		if err != nil {
			return err
		} else {
			singleSimPath.PathDescription = simPathDesc
		}
		simGraph.AvailablePaths = append(simGraph.AvailablePaths, singleSimPath)
	}
	return nil
}
