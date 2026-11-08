package atlasinit

import (
	"fmt"
	"strings"

	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/entities"
	"zhanghefan123/security_topology/utils/extract"
	"zhanghefan123/security_topology/utils/file"
)

// 路径文件初始化过程中的步骤名称。
const (
	LoadNodesAndPathsFromPathsFile       = "LoadNodesAndPathsFromPathsFile"
	LoadLinksFromPaths                   = "LoadLinksFromPaths"
	FindSplitPointsAndCorrespondingPaths = "FindSplitPointsAndCorrespondingPaths"
	PrintRelationshipMapping             = "PrintRelationshipMapping"
)

// InitFromPathsFile 按既定顺序从路径文件初始化图、路径和转发关系。
func (g *Graph) InitFromPathsFile() error {
	initSteps := []map[string]InitModule{
		{LoadNodesAndPathsFromPathsFile: {true, g.LoadNodesAndPathsFromPathsFile}},
		{LoadLinksFromPaths: {true, g.LoadLinksFromPaths}},
		{FindSplitPointsAndCorrespondingPaths: {true, g.FindSplitPointsAndCorrespondingPaths}},
		{PrintRelationshipMapping: {true, g.PrintRelationshipMapping}},
	}
	if err := g.initializeSteps(initSteps); err != nil {
		return fmt.Errorf("graph init failed: %w", err)
	}
	return nil
}

// LoadNodesAndPathsFromPathsFile 读取路径文件，并复用同名节点构建路径列表。
// 文件中的每一行表示一条路径，节点之间使用 "->" 分隔；路径编号从 1 开始。
func (g *Graph) LoadNodesAndPathsFromPathsFile() error {
	content, err := file.ReadFile(g.PathsFilePath)
	if err != nil {
		return fmt.Errorf("cannot find path: %v", err)
	}

	pathLines := strings.Split(content, "\n")
	var paths []*entities.Path

	for pathIndex, pathLine := range pathLines {
		currentPath := &entities.Path{PathId: pathIndex + 1}
		nodeNames := strings.Split(pathLine, "->")

		for _, nodeName := range nodeNames {
			var graphNode *entities.Node
			nodeIndex, _ := extract.NumberFromString(nodeName)

			// 同一个节点可能出现在多条路径中，只在首次遇到时加入图。
			if existingNode, ok := g.NameToNodeMapping[nodeName]; ok {
				graphNode = existingNode
			} else {
				graphNode = &entities.Node{
					Node:        nil,
					NodeName:    nodeName,
					ExcessValue: 0,
					Index:       nodeIndex,
				}
				newNode := g.DirectedGraph.NewNode()
				graphNode.Node = newNode
				g.DirectedGraph.AddNode(graphNode)
				g.NameToNodeMapping[nodeName] = graphNode
			}

			currentPath.NodeList = append(currentPath.NodeList, graphNode)
		}
		paths = append(paths, currentPath)
	}

	// 确保每条路径中的节点都持有名称映射中对应的底层图节点。
	for _, currentPath := range paths {
		for _, node := range currentPath.NodeList {
			node.Node = g.NameToNodeMapping[node.NodeName].Node
		}
	}

	g.KShortestPaths = paths
	return nil
}

// LoadLinksFromPaths 将每条路径中相邻的两个节点连接为有向边。
func (g *Graph) LoadLinksFromPaths() error {
	for _, currentPath := range g.KShortestPaths {
		for nodeIndex, currentNode := range currentPath.NodeList {
			if nodeIndex != len(currentPath.NodeList)-1 {
				nextNode := currentPath.NodeList[nodeIndex+1]
				directedEdge := g.DirectedGraph.NewEdge(currentNode, nextNode)
				g.DirectedGraph.SetEdge(directedEdge)
			}
		}
	}
	return nil
}

// FindSplitPointsAndCorrespondingPaths 找出出度至少为 2 的分叉节点，
// 并按下一跳节点归组经过该分叉的路径编号。
func (g *Graph) FindSplitPointsAndCorrespondingPaths() error {
	// 第一阶段：统计出度并识别分叉节点。
	splitNodes := make(map[string]*entities.Node)
	for _, node := range g.NameToNodeMapping {
		node.Outdegree = 0
	}

	edgeIterator := g.DirectedGraph.Edges()
	for edgeIterator.Next() {
		edge := edgeIterator.Edge()
		currentNode := g.NameToNodeMapping[edge.From().(*entities.Node).NodeName]
		currentNode.Outdegree++
		if currentNode.Outdegree >= 2 {
			splitNodes[currentNode.NodeName] = currentNode
		}
	}

	// 第二阶段：记录“分叉节点 -> 路径编号 -> 下一跳节点”的对应关系。
	splitNodeToPathTarget := make(map[string]map[int]string)
	for splitNodeName := range splitNodes {
		splitNodeToPathTarget[splitNodeName] = make(map[int]string)
		for _, currentPath := range g.KShortestPaths {
			for nodeIndex := range currentPath.NodeList {
				if nodeIndex != len(currentPath.NodeList)-1 {
					currentNode := currentPath.NodeList[nodeIndex]
					nextNode := currentPath.NodeList[nodeIndex+1]
					if splitNodeName == currentNode.NodeName {
						if _, exists := splitNodeToPathTarget[splitNodeName][currentPath.PathId]; !exists {
							splitNodeToPathTarget[splitNodeName][currentPath.PathId] = nextNode.NodeName
						}
					}
				}
			}
		}
	}

	// 第三阶段：将拥有相同下一跳的路径归为一条转发关系。
	for splitNodeName, pathToTargetNode := range splitNodeToPathTarget {
		targetNodeToPathIDs := make(map[int][]int)
		for pathID, targetNodeName := range pathToTargetNode {
			targetNodeID, _ := extract.NumberFromString(targetNodeName)
			targetNodeToPathIDs[targetNodeID] = append(targetNodeToPathIDs[targetNodeID], pathID)
		}

		for targetNodeID, pathIDs := range targetNodeToPathIDs {
			splitNodeID, _ := extract.NumberFromString(splitNodeName)
			relationship := &entities.Relationship{
				SplitNodeId:        splitNodeID,
				ForwardingTargetId: targetNodeID,
				NumberOfPaths:      len(pathIDs),
				PathIds:            &pathIDs,
			}
			g.RelationshipMapping[splitNodeName] = append(g.RelationshipMapping[splitNodeName], relationship)
		}
	}
	return nil
}

// PrintRelationshipMapping 按分叉节点打印其全部转发关系。
func (g *Graph) PrintRelationshipMapping() error {
	fmt.Println()
	for splitNodeName, relationships := range g.RelationshipMapping {
		fmt.Printf("----------------- split Node %s -----------------\n", splitNodeName)
		for _, relationship := range relationships {
			fmt.Println(relationship.BeautifulPrint())
		}
		fmt.Printf("----------------- split Node %s -----------------\n", splitNodeName)
	}
	return nil
}
