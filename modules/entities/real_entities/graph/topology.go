package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"zhanghefan123/security_topology/services/http/params"

	"gonum.org/v1/gonum/graph/path"
)

const (
	LoadBasicParams         = "LoadBasicParams"
	LoadNodes               = "LoadNodes"
	LoadLinks               = "LoadLinks"
	CalculateKShortestPaths = "CalculateKShortestPaths"
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

func (g *Graph) Init() error {
	initSteps := []map[string]InitModule{
		{LoadBasicParams: InitModule{true, g.LoadBasicParams}},
		{LoadNodes: InitModule{true, g.LoadNodes}},
		{LoadLinks: InitModule{true, g.LoadLinks}},
	}
	err := g.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("constellation init failed: %w", err)
	}
	return nil
}

func (g *Graph) initStepsNum(initSteps []map[string]InitModule) int {
	result := 0
	for _, initStep := range initSteps {
		for _, initModule := range initStep {
			if initModule.init {
				result += 1
			}
		}
	}
	return result
}

// InitializeSteps 按步骤进行初始化
func (g *Graph) initializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := g.initStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.init {
				if err = initModule.initFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				fmt.Printf("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

// LoadBasicParams 进行基础参数的加载
func (g *Graph) LoadBasicParams() error {
	g.GraphParams = &Params{}
	data, err := os.ReadFile(g.TopologyFilePath)
	if err != nil {
		return fmt.Errorf("read multiple path file failed, %s", err.Error())
	}
	err = json.Unmarshal(data, g.GraphParams)
	if err != nil {
		return fmt.Errorf("unmarshal multi path file failed, %s", err.Error())
	}
	return nil
}

// LoadNodes 进行所有的节点的加载
func (g *Graph) LoadNodes() error {
	for _, nodeParam := range g.GraphParams.Nodes {
		var graphNode *Node
		graphNodeName, err := params.ResolveNodeNameWithNodeParam(&nodeParam)
		if err != nil {
			return fmt.Errorf("resolve node name failed, %s", err.Error())
		}
		graphNode = CreateGraphNode(graphNodeName)
		newNode := g.DirectedGraph.NewNode()
		graphNode.Node = newNode
		g.DirectedGraph.AddNode(graphNode)
		g.NameToNodeMapping[graphNodeName] = graphNode
	}

	return nil
}

// LoadLinks 进行所有链路的加载
func (g *Graph) LoadLinks() error {
	for _, linkParam := range g.GraphParams.Links {
		// 拿到源和目的节点名称
		sourceNodeName, err := params.ResolveNodeNameWithNodeParam(&(linkParam.SourceNode))
		if err != nil {
			return fmt.Errorf("resolve node name failed, %s", err.Error())
		}
		targetNodeName, err := params.ResolveNodeNameWithNodeParam(&(linkParam.TargetNode))
		if err != nil {
			return fmt.Errorf("resolve node name failed, %s", err.Error())
		}
		// 获取实际节点
		sourceNode := g.NameToNodeMapping[sourceNodeName]
		targetNode := g.NameToNodeMapping[targetNodeName]

		// 构建新的边
		directEdge := g.DirectedGraph.NewEdge(sourceNode, targetNode)
		g.DirectedGraph.SetEdge(directEdge)
		reverseEdge := g.DirectedGraph.NewEdge(targetNode, sourceNode)
		g.DirectedGraph.SetEdge(reverseEdge)
	}
	return nil
}

// CalculateKShortestPaths 计算 K 条最短路径
func (g *Graph) CalculateKShortestPaths() error {
	finalResult := make([]*Path, 0)
	sourceNode := g.NameToNodeMapping[g.GraphParams.Source]
	targetNode := g.NameToNodeMapping[g.GraphParams.Destination]
	paths := path.YenKShortestPaths(g.DirectedGraph, g.GraphParams.NumberOfPaths, g.GraphParams.LimitOfCost, sourceNode, targetNode)
	for _, singlePath := range paths {
		singlePathModified := &Path{}
		for _, graphNode := range singlePath {
			singleModifiedNode := graphNode.(*Node)
			singlePathModified.NodeList = append(singlePathModified.NodeList, singleModifiedNode)
		}
		finalResult = append(finalResult, singlePathModified)
	}
	g.KShortestPaths = finalResult
	return nil
}
