package graph

import (
	"fmt"
	"strings"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
	"zhanghefan123/security_topology/utils/extract"
	"zhanghefan123/security_topology/utils/file"
)

const (
	LoadNodesAndPathsFromPathsFile       = "LoadNodesAndPathsFromPathsFile"
	LoadLinksFromPaths                   = "LoadLinksFromPaths"
	FindSplitPointsAndCorrespondingPaths = "FindSplitPointsAndCorrespondingPaths"
	PrintRelationshipMapping             = "PrintRelationshipMapping"
)

func (g *Graph) InitFromPathsFile() error {
	initSteps := []map[string]InitModule{
		{LoadNodesAndPathsFromPathsFile: InitModule{true, g.LoadNodesAndPathsFromPathsFile}},
		{LoadLinksFromPaths: InitModule{true, g.LoadLinksFromPaths}},
		{FindSplitPointsAndCorrespondingPaths: InitModule{true, g.FindSplitPointsAndCorrespondingPaths}},
		{PrintRelationshipMapping: InitModule{true, g.PrintRelationshipMapping}},
	}
	err := g.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("graph init failed: %w", err)
	}
	return nil
}

func (g *Graph) LoadNodesAndPathsFromPathsFile() error {
	result, err := file.ReadFile(g.PathsFilePath)
	if err != nil {
		return fmt.Errorf("cannot find path: %v", err)
	}
	// 将 result 进行分行
	paths := strings.Split(result, "\n")
	var allPaths []*entities.Path
	// 遍历每一个 path 并进行提取
	for index, path := range paths {
		singlePath := &entities.Path{}
		singlePath.PathId = index + 1
		nodeList := strings.Split(path, "->")
		for _, nodeName := range nodeList {
			var graphNode *entities.Node
			var nodeIndex int
			graphNodeName := nodeName
			nodeIndex, err = extract.NumberFromString(graphNodeName)
			// 进行节点的创建
			if _, ok := g.NameToNodeMapping[graphNodeName]; !ok {
				graphNode = entities.CreateGraphNode(graphNodeName, nodeIndex)
				newNode := g.DirectedGraph.NewNode()
				graphNode.Node = newNode
				g.DirectedGraph.AddNode(graphNode)
				g.NameToNodeMapping[graphNodeName] = graphNode
			} else {
				graphNode = g.NameToNodeMapping[graphNodeName]
			}
			singlePath.NodeList = append(singlePath.NodeList, graphNode)
		}
		allPaths = append(allPaths, singlePath)
	}

	// 遍历每条路径的新节点的设置
	for _, singlePath := range allPaths {
		// 遍历路径之中的每一个节点
		for _, node := range singlePath.NodeList {
			node.Node = g.NameToNodeMapping[node.NodeName].Node
		}
	}

	g.KShortestPaths = allPaths
	return nil
}

// LoadLinksFromPaths 进行边的加载
func (g *Graph) LoadLinksFromPaths() error {
	for _, singlePath := range g.KShortestPaths {
		for index, node := range singlePath.NodeList {
			if index != (len(singlePath.NodeList) - 1) {
				nextIndex := index + 1
				nextNode := singlePath.NodeList[nextIndex]
				// 创建新的边
				directEdge := g.DirectedGraph.NewEdge(node, nextNode)
				// 添加新的边
				g.DirectedGraph.SetEdge(directEdge)
			}
		}
	}
	return nil
}

func (g *Graph) FindSplitPointsAndCorrespondingPaths() error {
	// out degree lager than 2
	splitNodes := make(map[string]*entities.Node)
	edgeIterator := g.DirectedGraph.Edges()
	for _, node := range g.NameToNodeMapping {
		node.Outdegree = 0
	}
	for {
		if !(edgeIterator.Next()) {
			break
		}
		edge := edgeIterator.Edge()
		currentNode := g.NameToNodeMapping[edge.From().(*entities.Node).NodeName]
		currentNode.Outdegree++
		if currentNode.Outdegree >= 2 {
			splitNodes[currentNode.NodeName] = currentNode
		}
	}

	// node split
	finalResultMapping := make(map[string]map[int]string)

	// 进行所有的查找
	for splitNodeName, _ := range splitNodes {
		finalResultMapping[splitNodeName] = make(map[int]string)
		for _, path := range g.KShortestPaths {
			for index, _ := range path.NodeList {
				if index != (len(path.NodeList) - 1) {
					currentNode := path.NodeList[index]
					nextNode := path.NodeList[index+1]
					if splitNodeName == currentNode.NodeName {
						if _, ok := finalResultMapping[splitNodeName][path.PathId]; !ok {
							finalResultMapping[splitNodeName][path.PathId] = nextNode.NodeName
						}
					}
				}
			}
		}
	}

	// 进行最终的每个节点对应的 path 的打印
	for splitNode, pathToTargetNodeId := range finalResultMapping {
		finalString := ""
		targetNodeToPathIdsMapping := map[int][]int{}
		for pathId, targetNode := range pathToTargetNodeId {
			// 1->20 | 3->20
			targetNodeId, _ := extract.NumberFromString(targetNode)
			targetNodeToPathIdsMapping[targetNodeId] = append(targetNodeToPathIdsMapping[targetNodeId], pathId)
		}

		for targetNodeId, pathIds := range targetNodeToPathIdsMapping {
			splitNodeId, _ := extract.NumberFromString(splitNode)
			finalString = fmt.Sprintf("%d,%d,%d,", splitNodeId, targetNodeId, len(pathIds))
			for pathIndex, pathId := range pathIds {
				if pathIndex != (len(pathIds) - 1) {
					finalString += fmt.Sprintf("%d,", pathId)
				} else {
					finalString += fmt.Sprintf("%d", pathId)
				}
			}
			g.RelationshipMapping[splitNode] = append(g.RelationshipMapping[splitNode], finalString)
		}
	}
	return nil
}

func (g *Graph) PrintRelationshipMapping() error {
	fmt.Println()
	for splitNode, relationshipList := range g.RelationshipMapping {
		fmt.Printf("----------------- split Node %s -----------------\n", splitNode)
		for _, relationship := range relationshipList {
			fmt.Println(relationship)
		}
		fmt.Printf("----------------- split Node %s -----------------\n", splitNode)
	}
	return nil
}
