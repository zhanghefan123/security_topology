package graph

import (
	"gonum.org/v1/gonum/graph/simple"
	"sort"
)

// CalculateExcessValue 计算流经的路径的数量
func CalculateExcessValue(multipaths []*Path, multipathNodeMapping map[string]*MultipathGraphNode) {
	for _, multipathNode := range multipathNodeMapping {
		excessValue := 0
		// 进行所有的路径的遍历
		for _, path := range multipaths {
			for _, nodeInPath := range path.NodeList {
				if nodeInPath.NodeName == multipathNode.NodeName {
					excessValue += 1
					break
				}
			}
		}
		multipathNode.ExcessValue = excessValue
	}
}

// FindNodesWithExcessValueSameAsSource 找到所有和源节点具有相同的 ExcessValue 的节点
func FindNodesWithExcessValueSameAsSource(multipaths []*Path, source *MultipathGraphNode) []*MultipathGraphNode {
	// 已经添加的节点
	alreadyAddedNodes := map[string]struct{}{}
	// 最终结果
	var nodeList []*MultipathGraphNode
	// 进行所有的 path 的遍历
	for _, path := range multipaths {
		for _, node := range path.NodeList {
			if _, ok := alreadyAddedNodes[node.NodeName]; (!ok) && (node.ExcessValue == source.ExcessValue) {
				nodeList = append(nodeList, node)
				alreadyAddedNodes[node.NodeName] = struct{}{}
			}
		}
	}
	return nodeList
}

// FindEdgeWithinSegmentButNotInGraph 找到存在于 Segment 但是不在 Graph 之中的
func FindEdgeWithinSegmentButNotInGraph(segment *Segment, multipathGraph *simple.DirectedGraph) []*DirectedEdge {
	var virtualEdges []*DirectedEdge
	for index, node := range segment.Path {
		if index == (len(segment.Path) - 1) {
			break
		}
		nextNode := segment.Path[index+1]
		if !multipathGraph.HasEdgeFromTo(node.ID(), nextNode.ID()) {
			virtualEdges = append(virtualEdges, CreateDirectedEdge(node, nextNode))
		}
	}
	return virtualEdges
}

// FindPathsInVirtualEdge 找到在 virtual Edge 之中的 paths
func FindPathsInVirtualEdge(directedEdge *DirectedEdge, paths []*Path) []*Path {
	var finalPaths []*Path
	for _, path := range paths {
		subPath := &Path{}
		recording := false
		for _, node := range path.NodeList {
			// 从这个节点开始
			if directedEdge.From.NodeName == node.NodeName {
				recording = true
			}

			// recording == true 就进行记录
			if recording {
				subPath.NodeList = append(subPath.NodeList, node)
			}

			// 到这个节点结束
			if directedEdge.To.NodeName == node.NodeName {
				recording = false
			}
		}
		finalPaths = append(finalPaths, subPath)
	}
	return finalPaths
}

func FindConvergencePoints(graphNodeMapping map[string]*MultipathGraphNode) []*MultipathGraphNode {
	var covergencePoints []*MultipathGraphNode
	for _, node := range graphNodeMapping {
		if node.ExcessValue >= 2 {
			covergencePoints = append(covergencePoints, node)
		}
	}
	// 将这些节点按照 ExcessValue 从小到大进行排序

	// 按照 ExcessValue 从小到大进行排序
	sort.Slice(covergencePoints, func(i, j int) bool {
		return covergencePoints[i].ExcessValue < covergencePoints[j].ExcessValue
	})

	return covergencePoints
}

func FindSubsetPathIncludingANode(paths []*Path, node *MultipathGraphNode) ([]*Path, []*Path) {
	var selectedPaths []*Path
	var ignoredPaths []*Path
	for _, path := range paths {
		selectPath := false
		for _, tmpNode := range path.NodeList {
			if tmpNode.NodeName == node.NodeName {
				selectPath = true
				break
			}
		}
		if selectPath {
			selectedPaths = append(selectedPaths, path)
		} else {
			ignoredPaths = append(ignoredPaths, path)
		}
	}
	return selectedPaths, ignoredPaths
}

func RemoveVirtualEdgeDestinationFromConvergencePoints(convergencePoints []*MultipathGraphNode, destination *MultipathGraphNode) []*MultipathGraphNode {
	result := make([]*MultipathGraphNode, 0, len(convergencePoints))
	for _, node := range convergencePoints {
		if node.NodeName != destination.NodeName {
			result = append(result, node)
		}
	}
	return result
}
