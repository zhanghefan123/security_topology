package graph

import (
	"math"
	"sort"
	"strconv"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

var finalSegments []*Segment
var finallyFuck bool = false
var fuckResult []*Path

// HierarchyDivision 进行分层切割 (paths 代表多路径, depth 代表深度)
func HierarchyDivision(paths []*Path, depth int) {
	// step 5 Build directed mulitpath graph using paths
	createdGraph, multipathNodeMapping, sourceAndDest := CreateNewGraphFromRealPaths(paths)
	// step 6-7 traverse each node in graph and calculate their excess value
	CalculateExcessValue(paths, multipathNodeMapping)
	//fmt.Println("------------------------------------------------------")
	//for _, node := range multipathNodeMapping {
	//	fmt.Printf("multipath node: %s excess value: %v\n", node.NodeName, node.ExcessValue)
	//}
	//fmt.Println("------------------------------------------------------")
	// step 8-9 进行 segment 的创建
	segment := CreateSegment(sourceAndDest.Source, sourceAndDest.Destination, depth, nil)
	// step 10-12 找到和源具有相同 excess value 的节点并入到其中
	nodeListWithSameExcessValue := FindNodesWithExcessValueSameAsSource(paths, sourceAndDest.Source)
	//fmt.Println("------------------------------------------------------")
	//for _, node := range nodeListWithSameExcessValue {
	//	fmt.Printf("node with same excess value: %v\n", node.NodeName)
	//}
	//fmt.Println("------------------------------------------------------")
	// 将 nodeList 按照到目的节点的距离进行排序
	sortedHighLevelNodes := SortNodeListBasedOnDistanceToDestination(nodeListWithSameExcessValue, createdGraph, sourceAndDest.Source)
	//fmt.Println("------------------------------------------------------")
	//finalString := ""
	//for index, node := range sortedHighLevelNodes {
	//	if index != len(sortedHighLevelNodes)-1 {
	//		finalString += node.NodeName + "->"
	//	} else {
	//		finalString += node.NodeName
	//	}
	//}
	//fmt.Printf("high-level path: %v\n", finalString)
	//fmt.Println("------------------------------------------------------")
	segment.Path = sortedHighLevelNodes
	// step 13 将 segment 并入到 finalSegments 之中
	finalSegments = append(finalSegments, segment)
	// step 14 进行所有在 segment 之中的但是不在 graph 之中的边的遍历
	virtualEdges := FindVirtualEdgesInHighLevelPath(segment, createdGraph)
	//fmt.Println("------------------------------------------------------")
	//for _, edge := range virtualEdges {
	//	fmt.Printf("virtual edge: %s <-> %s\n", edge.From.NodeName, edge.To.NodeName)
	//}
	//fmt.Println("------------------------------------------------------")
	for _, virtualEdge := range virtualEdges {
		// step 15 找到 virtual edge 的连接的子路径
		subpaths := FindPathsInVirtualEdge(virtualEdge, paths)
		// step 16 通过这些 subpaths 构建一个 subgraph
		subGraph, subGraphNodeMapping, _ := CreateNewGraphFromRealPaths(subpaths)

		// step 17 找到大于等于2 的入度的节点
		CalculateExcessValue(subpaths, subGraphNodeMapping)
		//convergencePoints := FindConvergencePointsUsingExcessValue(subGraphNodeMapping)
		convergencePoints := FindConvergencePointsUsingIndegree(subGraph, subGraphNodeMapping)
		if len(convergencePoints) == 0 {
			finallyFuck = true
			fuckResult = subpaths
		}
		//fmt.Println()
		//fmt.Println("------------------------------------------------------")
		//fmt.Printf("virtual edge: %s <-> %s\n", virtualEdge.From.NodeName, virtualEdge.To.NodeName)
		//PrintPaths(subpaths)
		//for _, node := range convergencePoints {
		//	fmt.Printf("convergence point: %s\n", node.NodeName)
		//}
		//fmt.Println("------------------------------------------------------")
		// step 19 判断是否是空集
		for {
			if len(convergencePoints) == 0 {
				for _, subpath := range subpaths {
					// step 21-23 创建一个新的 segment:
					subPathSegment := CreateSegment(virtualEdge.From, virtualEdge.To, depth+1, subpath.NodeList)
					finalSegments = append(finalSegments, subPathSegment)
				}
				break
			}
			firstConvergencePoint := convergencePoints[0]
			// step 26
			if len(convergencePoints) > 1 {
				convergencePoints = convergencePoints[1:]
			} else {
				convergencePoints = []*Node{}
				for _, subpath := range subpaths {
					// step 21-23 创建一个新的 segment:
					subPathSegment := CreateSegment(virtualEdge.From, virtualEdge.To, depth+1, subpath.NodeList)
					finalSegments = append(finalSegments, subPathSegment)
				}
				break
			}
			//fmt.Println("------------------------------------------------------")
			//fmt.Println("first convergence point:", firstConvergencePoint.NodeName)
			//fmt.Println("------------------------------------------------------")
			// step 27 可能前面的节点把后面的节点的所在的 path 给选完了，造成一个空path集合, 但是可能还存在点没被选到，我们要一直选择到所有的节点都选择完为止
			selectedPaths, ignoredPaths := FindSubsetPathIncludingTheConvergencePoint(subpaths, firstConvergencePoint) // 找到包含某个聚合点的所有路径
			if len(selectedPaths) == 0 {
				//fmt.Println("No selected paths, ignored paths:", ignoredPaths)
				//for _, singlePath := range subpaths {
				//	fmt.Printf("singleSubPath: %v\n", singlePath)
				//}
				continue // 可能有节点还没选完
			}
			// step 28 进行递归调用
			HierarchyDivision(selectedPaths, depth+1)
			// step 29 调用完成后将节点移除, 路径也进行移除
			//fmt.Println("convergencePoints length: ", len(convergencePoints))

			subpaths = ignoredPaths // 剩下的这个路径还有用到吗？
			// 视情况进行 break
			if len(convergencePoints) == 0 {
				break
			}
		}
	}
}

// CalculateExcessValue 计算流经的路径的数量
func CalculateExcessValue(multipaths []*Path, multipathNodeMapping map[string]*Node) {
	// 遍历每个节点
	for _, multipathNode := range multipathNodeMapping {
		// 初始的 excessValue == 0
		excessValue := 0
		// 进行所有的路径的遍历
		for _, SinglePath := range multipaths {
			for _, nodeInPath := range SinglePath.NodeList {
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
func FindNodesWithExcessValueSameAsSource(multipaths []*Path, source *Node) map[string]*Node {
	// 已经添加的节点
	nodesWithExcessValueSameAsSource := map[string]*Node{}
	// 进行所有的 path 的遍历
	for _, singlePath := range multipaths {
		for _, node := range singlePath.NodeList {
			if _, ok := nodesWithExcessValueSameAsSource[node.NodeName]; (!ok) && (node.ExcessValue == source.ExcessValue) {
				nodesWithExcessValueSameAsSource[node.NodeName] = node
			}
		}
	}
	return nodesWithExcessValueSameAsSource
}

// SortNodeListBasedOnDistanceToDestination 根据到目的节点的举例将 nodeList 进行排序
func SortNodeListBasedOnDistanceToDestination(nodesMapping map[string]*Node, graphTmp *simple.DirectedGraph, source *Node) []*Node {
	var finalNodeList []*Node
	dijkstraResult := path.DijkstraFrom(source, graphTmp)
	for _, node := range nodesMapping {
		distance := dijkstraResult.WeightTo(node.ID())
		node.Distance = distance
		finalNodeList = append(finalNodeList, node)
	}
	// 将 nodeList 按照 Distance 进行从小到大的排列
	sort.Slice(finalNodeList, func(i, j int) bool {
		return finalNodeList[i].Distance < finalNodeList[j].Distance
	})
	// 最终结果
	return finalNodeList
}

// FindVirtualEdgesInHighLevelPath 一种情况是高级节点之间没有边连接, 一种情况是高级节点间不但存在直接边连接，还存在间接的边的连接
func FindVirtualEdgesInHighLevelPath(segment *Segment, multipathGraph *simple.DirectedGraph) []*DirectedEdge {
	// 找到的虚链路
	var virtualEdges []*DirectedEdge
	// segment 之中的高层节点的顺序可能是不对的, 我们需要通过最短路径算法算一遍得到应该的顺序是什么样的
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			nextNode := segment.Path[index+1]
			// 第一种情况 -> 不包含边
			if !(multipathGraph.HasEdgeBetween(node.ID(), nextNode.ID())) {
				virtualEdges = append(virtualEdges, &DirectedEdge{
					From: node,
					To:   nextNode,
				})
			}
			// 第二种情况 -> 不但存在直接边连接，还存在间接的边的连接
			if multipathGraph.HasEdgeBetween(node.ID(), nextNode.ID()) {
				// 将边暂时进行删除, 然后看看是否能够抵达
				multipathGraph.RemoveEdge(node.ID(), nextNode.ID())
				spf := path.DijkstraFrom(node, multipathGraph)
				_, weight := spf.To(nextNode.ID())
				if weight != math.Inf(1) {
					virtualEdges = append(virtualEdges, &DirectedEdge{
						From: node,
						To:   nextNode,
					})
				}
				// 重新进行边的创建
				newEdge := multipathGraph.NewEdge(node, nextNode)
				multipathGraph.SetEdge(newEdge)
			}
		}
	}
	// 将结果进行返回
	return virtualEdges
}

// FindEdgeWithinSegmentButNotInGraph 找到存在于 Segment 但是不在 Graph 之中的
func FindEdgeWithinSegmentButNotInGraph(segment *Segment, multipathGraph *simple.DirectedGraph) []*DirectedEdge {
	// 找到的虚链路
	var virtualEdges []*DirectedEdge
	// segment 之中的高层节点的顺序可能是不对的, 我们需要通过最短路径算法算一遍得到应该的顺序是什么样的
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			nextNode := segment.Path[index+1]
			if !(multipathGraph.HasEdgeBetween(node.ID(), nextNode.ID())) {
				virtualEdges = append(virtualEdges, &DirectedEdge{
					From: node,
					To:   nextNode,
				})
			}
		}
	}
	// 将结果进行返回
	return virtualEdges
}

// FindPathsInVirtualEdge 找到在 virtual Edge 之中的 paths
func FindPathsInVirtualEdge(directedEdge *DirectedEdge, paths []*Path) []*Path {
	var finalPaths []*Path
	var subpathMap = make(map[string]*Path)
	for _, singlePath := range paths {
		subPath := &Path{}
		subPathString := ""
		recording := false
		for _, node := range singlePath.NodeList {
			// 从这个节点开始
			if directedEdge.From.NodeName == node.NodeName {
				recording = true
			}

			// recording == true 就进行记录
			if recording {
				subPath.NodeList = append(subPath.NodeList, node)
				subPathString = subPathString + node.NodeName + "->"
			}

			// 到这个节点结束
			if directedEdge.To.NodeName == node.NodeName {
				recording = false
			}
		}

		// 从 subpaths 到 path 可能发生重复需要进行去重
		if _, ok := subpathMap[subPathString]; !ok {
			finalPaths = append(finalPaths, subPath)
			subpathMap[subPathString] = subPath
		}
	}

	return finalPaths
}

// FindConvergencePointsUsingExcessValue 这个会把分散点也考虑进去, 所以只能使用基于 Indegree 的策略
/*
func FindConvergencePointsUsingExcessValue(graphNodeMapping map[string]*Node) []*Node {
	var finalResult []*Node
	for _, node := range graphNodeMapping {
		if node.ExcessValue >= 2 {
			finalResult = append(finalResult, node)
		}
	}
	return finalResult
}
*/

// FindConvergencePointsUsingIndegree 找到 ExcessValue >= 2 的汇聚点
func FindConvergencePointsUsingIndegree(directedGraph *simple.DirectedGraph, graphNodeMapping map[string]*Node) []*Node {
	var convergencePoints []*Node
	alreadyAddedPoints := make(map[string]struct{})
	edgeIterator := directedGraph.Edges()
	for _, node := range graphNodeMapping {
		node.Indegree = 0
	}
	for {
		if !(edgeIterator.Next()) {
			break
		}
		edge := edgeIterator.Edge()
		graphNodeMapping[edge.To().(*Node).NodeName].Indegree++
		if graphNodeMapping[edge.To().(*Node).NodeName].Indegree >= 2 {
			if _, ok := alreadyAddedPoints[edge.To().(*Node).NodeName]; !ok {
				alreadyAddedPoints[edge.To().(*Node).NodeName] = struct{}{}
				convergencePoints = append(convergencePoints, edge.To().(*Node))
			}
		}
	}
	// 按照 ExcessValue 从小到大进行排序
	sort.Slice(convergencePoints, func(i, j int) bool {
		return convergencePoints[i].ExcessValue < convergencePoints[j].ExcessValue
	})

	finalString := ""
	for _, point := range convergencePoints {
		finalString = finalString + strconv.Itoa(point.ExcessValue) + "->"
	}
	// fmt.Printf("ConvergencePoints: %s\n", finalString)

	return convergencePoints
}

func CalculateIndegree(directedGraph *simple.DirectedGraph, graphNodeMapping map[string]*Node) {
	edgeIterator := directedGraph.Edges()
	for _, node := range graphNodeMapping {
		node.Indegree = 0
	}
	for {
		if !(edgeIterator.Next()) {
			break
		}
		edge := edgeIterator.Edge()
		graphNodeMapping[edge.To().(*Node).NodeName].Indegree++
	}
}

// FindSubsetPathIncludingTheConvergencePoint 在一个大的路径集合之中找到包含某个节点的路径子集，以及被排除的路径的集合
func FindSubsetPathIncludingTheConvergencePoint(paths []*Path, node *Node) ([]*Path, []*Path) {
	// 选择的路径
	var selectedPaths []*Path
	// 忽略的路径
	var ignoredPaths []*Path
	// 进行所有的路径遍历
	for _, singlePath := range paths {
		// 一开始判断是不包含的
		selectPath := false
		for _, tmpNode := range singlePath.NodeList {
			if tmpNode.NodeName == node.NodeName {
				selectPath = true
				break
			}
		}
		if selectPath { // 如果判定包含则加入 selectedPaths 之中
			selectedPaths = append(selectedPaths, singlePath)
		} else { // 如果判定不包含则加入 ignoredPaths 之中
			ignoredPaths = append(ignoredPaths, singlePath)
		}
	}
	return selectedPaths, ignoredPaths
}

// RemoveVirtualEdgeDestinationFromConvergencePoints 将虚链路的目的节点从汇聚点集合之中移除
func RemoveVirtualEdgeDestinationFromConvergencePoints(convergencePoints []*Node, destination *Node) []*Node {
	result := make([]*Node, 0, len(convergencePoints))
	for _, node := range convergencePoints {
		if node.NodeName != destination.NodeName {
			result = append(result, node)
		}
	}
	return result
}
