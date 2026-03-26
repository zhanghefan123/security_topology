package entities

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/params"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type SimGraph struct {
	DirectedGraph               *simple.DirectedGraph
	GraphParams                 *params.GraphParams
	SourceNode                  *SimAbstractNode
	DestinationNode             *SimAbstractNode
	SimAbstractNodes            []*SimAbstractNode
	SimDirectedRealLinks        []*SimDirectedRealLink
	SimDirectedAbsLinks         []*SimDirectedAbsLink
	SimAbstractNodesMapping     map[string]*SimAbstractNode
	SimDirectedAbsLinksMapping  map[string]*SimDirectedAbsLink
	SimDirectedRealLinksMapping map[string]*SimDirectedRealLink

	AvailablePaths       []*SimPath
	CoveragePaths        []*SimPath          // 覆盖这张图所有边的路径
	AvailablePathMapping map[string]*SimPath // 从描述到实际对象的 mapping
	CoveragePathMapping  map[string]*SimPath // 从描述到实际对象的 mapping

	TotalPathWeights float64    // 所有路径的总的权重
	SelectedPaths    []*SimPath // 整个模拟过程选择的路径序列
	Regrets          []float64  // 进行悔值的计算
	CurrentLoss      float64    // 当前的损失
}

func NewSimGraph() *SimGraph {
	return &SimGraph{
		DirectedGraph:               simple.NewDirectedGraph(),
		GraphParams:                 &params.GraphParams{},
		SimAbstractNodes:            make([]*SimAbstractNode, 0),
		SimDirectedRealLinks:        make([]*SimDirectedRealLink, 0),
		SimDirectedAbsLinks:         make([]*SimDirectedAbsLink, 0),
		SimAbstractNodesMapping:     make(map[string]*SimAbstractNode),
		SimDirectedRealLinksMapping: make(map[string]*SimDirectedRealLink),
		SimDirectedAbsLinksMapping:  make(map[string]*SimDirectedAbsLink),

		AvailablePaths:       make([]*SimPath, 0),
		CoveragePaths:        make([]*SimPath, 0),
		AvailablePathMapping: make(map[string]*SimPath),
		CoveragePathMapping:  make(map[string]*SimPath),

		TotalPathWeights: 0,
		SelectedPaths:    make([]*SimPath, 0),
		Regrets:          make([]float64, 0),
		CurrentLoss:      0,
	}
}

// IsCoveragePath 判定某条路径是否是覆盖集之中的路径
func (simGraph *SimGraph) IsCoveragePath(path *SimPath) bool {
	if _, ok := simGraph.CoveragePathMapping[path.Description]; ok {
		return true
	} else {
		return false
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
			var newSimNormalRouter *SimNormalRouter
			newSimNormalRouter, err = NewSimNormalRouter(nodeName, nodeParam.Index,
				nodeParam.DropRatio.Start, nodeParam.DropRatio.End,
				nodeParam.CorruptRatio.Start, nodeParam.CorruptRatio.End,
				nodeParam.CorruptSpecialPacketRatio.Start, nodeParam.CorruptSpecialPacketRatio.End)
			if err != nil {
				return fmt.Errorf("create newSimNormalRouter failed, %w", err)
			}
			// create sim abstract node
			graphNode := simGraph.DirectedGraph.NewNode()
			simAbstractNode := NewSimAbstract(nodeType, newSimNormalRouter, graphNode)
			// add to list
			simGraph.SimAbstractNodes = append(simGraph.SimAbstractNodes, simAbstractNode)
			// add to mapping
			simGraph.SimAbstractNodesMapping[newSimNormalRouter.NodeName] = simAbstractNode
			// add node to the graph
			simGraph.DirectedGraph.AddNode(simAbstractNode)
		} else if nodeType == types.SimNetworkNodeType_PathValidationRouter {
			// create actual node
			newSimPathValidationRouter := NewSimPathValidationRouter(nodeName, nodeParam.Index)
			// create sim abstract node
			graphNode := simGraph.DirectedGraph.NewNode()
			simAbstractNode := NewSimAbstract(nodeType, newSimPathValidationRouter, graphNode)
			// add to list
			simGraph.SimAbstractNodes = append(simGraph.SimAbstractNodes, simAbstractNode)
			// add to mapping
			simGraph.SimAbstractNodesMapping[newSimPathValidationRouter.NodeName] = simAbstractNode
			// add node to the graph
			simGraph.DirectedGraph.AddNode(simAbstractNode)
		} else if nodeType == types.SimNetworkNodeType_EndHost {
			// create actual node
			newEndHost := NewEndHost(nodeName, nodeParam.Index)
			// create sim abstract node
			graphNode := simGraph.DirectedGraph.NewNode()
			simAbstractNode := NewSimAbstract(nodeType, newEndHost, graphNode)
			// add to list
			simGraph.SimAbstractNodes = append(simGraph.SimAbstractNodes, simAbstractNode)
			// add to mapping
			simGraph.SimAbstractNodesMapping[newEndHost.NodeName] = simAbstractNode
			// add to graph
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

func (simGraph *SimGraph) LoadLinkParams(linkType types.SimDirectedLinkType, linkParams []params.SimAbsLinkParam) error {
	// 处理 access links
	for _, linkParam := range linkParams {
		// get source name
		sourceNodeName, err := params.ResolveSimNodeName(&linkParam.SourceNode)
		if err != nil {
			return fmt.Errorf("resolve source node name failed, %w", err)
		}
		intermediateNodeName, err := params.ResolveSimNodeName(&linkParam.IntermediateNode)
		if err != nil {
			return fmt.Errorf("resolve intermediate node name failed, %w", err)
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
		intermediateSimAbstractNode, ok := simGraph.SimAbstractNodesMapping[intermediateNodeName]
		if !ok {
			return fmt.Errorf("cannot find intermediate sim abstract node, name: %s", intermediateNodeName)
		}
		targetSimAbstractNode, ok := simGraph.SimAbstractNodesMapping[targetNodeName]
		if !ok {
			return fmt.Errorf("cannot find target sim abstract node, name: %s", targetNodeName)
		}
		// get description
		linkDesc := fmt.Sprintf("%s->%s->%s", sourceNodeName, intermediateNodeName, targetNodeName)
		// create pv link
		link := NewSimDirectedAbsLink(linkType, linkDesc, sourceSimAbstractNode, intermediateSimAbstractNode, targetSimAbstractNode)
		// update pv link list
		simGraph.SimDirectedAbsLinks = append(simGraph.SimDirectedAbsLinks, link)
		// update pv link mapping
		if _, ok = simGraph.SimDirectedAbsLinksMapping[linkDesc]; !ok {
			simGraph.SimDirectedAbsLinksMapping[linkDesc] = link
		} else {
			return fmt.Errorf("duplicate link desc: %s", linkDesc)
		}
	}
	return nil
}

// LoadAccessLinksAndPvLinksParams 从 GraphParams 中的 access 链路参数信息中加载图中的 accessLink / 从 pvLink 参数信息中加载图中的 pvLink
func (simGraph *SimGraph) LoadAccessLinksAndPvLinksParams() error {

	err := simGraph.LoadLinkParams(types.SimDirectedLinkType_SimDirectedAccessLink, simGraph.GraphParams.AccessLinks)
	if err != nil {
		return fmt.Errorf("load access links failed due to: %w", err)
	}
	err = simGraph.LoadLinkParams(types.SimDirectedLinkType_SimDirectedPvLink, simGraph.GraphParams.PvLinks)
	if err != nil {
		return fmt.Errorf("load pv links failed due to: %w", err)
	}

	return nil
}

// LoadRealLinksFromLinkParams 从 GraphParams 中的链路参数信息中加载图中的链路
func (simGraph *SimGraph) LoadRealLinksFromLinkParams() error {
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
		// get description
		realLinkDesc := fmt.Sprintf("%s->%s", sourceNodeName, targetNodeName)
		// create real link
		realLink := NewSimDirectedRealLink(sourceSimAbstractNode, targetSimAbstractNode)
		directGraphEdge := simGraph.DirectedGraph.NewEdge(sourceSimAbstractNode, targetSimAbstractNode)
		simGraph.DirectedGraph.SetEdge(directGraphEdge)
		// append
		simGraph.SimDirectedRealLinks = append(simGraph.SimDirectedRealLinks, realLink)
		// update mapping
		if _, ok = simGraph.SimDirectedRealLinksMapping[sourceNodeName]; !ok {
			simGraph.SimDirectedRealLinksMapping[realLinkDesc] = realLink
		} else {
			return fmt.Errorf("duplicate real link description: %s", realLinkDesc)
		}
	}
	return nil
}

// CalculateKShortestPaths 计算图中从源节点到目的节点的 k 条最短路径，并更新 AvailablePaths 和 AvailablePathMapping
func (simGraph *SimGraph) CalculateKShortestPaths() error {
	paths := path.YenKShortestPaths(simGraph.DirectedGraph, simGraph.GraphParams.KShortestPathParamas.NumberOfPaths, simGraph.GraphParams.KShortestPathParamas.LimitOfCost, simGraph.SourceNode, simGraph.DestinationNode)

	for _, graphNodeList := range paths {
		singleSimPath := NewSimPath()
		// fill node list in this path
		for _, abstractSimNode := range graphNodeList {
			singleModifiedNode, ok := abstractSimNode.(*SimAbstractNode)
			if ok {
				singleSimPath.NodeList = append(singleSimPath.NodeList, singleModifiedNode)
			} else {
				return fmt.Errorf("type assertion failed when calculating k shortest path")
			}
		}
		// fill directedPvLinks and directedPvLinksMapping in this path
		err := singleSimPath.UpdateInfo(simGraph.SimDirectedAbsLinksMapping)
		if err != nil {
			return fmt.Errorf("fill directed pv links failed due to: %w", err)
		}

		// get path description for this path
		simPathDesc, err := singleSimPath.GetPathDescription()
		if err != nil {
			return err
		} else {
			singleSimPath.Description = simPathDesc
		}
		// update available path list
		simGraph.AvailablePaths = append(simGraph.AvailablePaths, singleSimPath)
		// update available path mapping
		simGraph.AvailablePathMapping[simPathDesc] = singleSimPath
		// calculate link score (link score 是用来进行排序的)
		singleSimPath.CalculateScore()
	}

	// 进行排序
	sort.Slice(simGraph.AvailablePaths, func(i, j int) bool {
		return simGraph.AvailablePaths[i].Score < simGraph.AvailablePaths[j].Score
	})

	// 给路径进行 id 的分配
	for index, simPath := range simGraph.AvailablePaths {
		simPath.PathId = index + 1
	}

	return nil
}

// LoadCoveragePathsFromParams 从 GraphParams 中的覆盖路径参数信息中加载图中的覆盖路径，并更新 CoveragePaths 和 CoveragePathMapping
func (simGraph *SimGraph) LoadCoveragePathsFromParams() error {
	// for each coverage path string, construct the coverage path and update the coverage path list and mapping
	for _, coveragePathString := range simGraph.GraphParams.CoveragePaths {
		pathDescription := ""
		routerListString := strings.Split(coveragePathString, ",")
		for index, routerName := range routerListString {
			if index != (len(routerListString) - 1) {
				pathDescription += fmt.Sprintf("%s->", routerName)
			} else {
				pathDescription += fmt.Sprintf("%s", routerName)
			}
		}
		// get the coverage path in mapping via description
		if coveragePath, ok := simGraph.AvailablePathMapping[pathDescription]; ok {
			// update the coverage paths
			simGraph.CoveragePaths = append(simGraph.CoveragePaths, coveragePath)
			// update the coverage path mapping
			simGraph.CoveragePathMapping[pathDescription] = coveragePath
		} else {
			return fmt.Errorf("k shortest paths have no path: %s", pathDescription)
		}
	}
	return nil
}

// RemovePathContainingTheLink 移除掉包含链路的路径
func (simGraph *SimGraph) RemovePathContainingTheLink(pvLink *SimDirectedAbsLink) {
	result := make([]*SimPath, 0)
	for _, simPath := range simGraph.AvailablePaths {
		if _, ok := simPath.DirectedAbsLinksMapping[pvLink.Description]; !ok {
			result = append(result, simPath)
		}
	}
	simGraph.AvailablePaths = result
}
