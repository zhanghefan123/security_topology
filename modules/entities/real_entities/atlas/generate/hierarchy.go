package generate

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"

	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/entities"
	"zhanghefan123/security_topology/utils/extract"
)

// HierarchyDivision 将同源同目的的路径集合递归分解为不同层级的 segment。
// paths 应包含至少一条非空路径；depth 表示当前层级。
// 结果追加到 finalSegments，当前层 segment 先于其递归生成的子 segment。
func HierarchyDivision(paths []*entities.Path, depth int, finalSegments *[]*entities.Segment) {
	// 1. 构建当前路径集合的图，统计经过各节点的路径数（ExcessValue）。
	multipathGraph, nodeMapping, sourceAndDest := entities.CreateNewGraphFromRealPaths(paths)
	CalculateExcessValue(paths, nodeMapping)

	// 2. 用与源节点 ExcessValue 相同的节点组成当前层路径。
	// 在同源路径集合中，这些节点是所有路径的公共节点。
	segment := entities.CreateSegment(sourceAndDest.Source, sourceAndDest.Destination, depth, nil)
	highLevelNodes := FindNodesWithExcessValueSameAsSource(paths, sourceAndDest.Source)
	segment.Path = SortNodeListBasedOnDistanceToDestination(highLevelNodes, multipathGraph, sourceAndDest.Destination)
	*finalSegments = append(*finalSegments, segment)

	// 3. 展开当前层路径中的虚拟边，分别处理其两端之间的子路径。
	virtualEdges := FindVirtualEdgesInHighLevelPath(segment, multipathGraph)
	for _, virtualEdge := range virtualEdges {
		remainingPaths := FindPathsInVirtualEdge(virtualEdge, paths)
		subgraph, subgraphNodes, _ := entities.CreateNewGraphFromRealPaths(remainingPaths)

		// 汇聚点按入度识别，按 ExcessValue、节点编号依次升序排列。
		// 此列表只在递归处理前计算一次，后续仅消费列表，不重新排序。
		CalculateExcessValue(remainingPaths, subgraphNodes)
		convergencePoints := FindConvergencePointsUsingIndegree(subgraph, subgraphNodes)

		// 4. 依次递归处理经过汇聚点的路径子集。
		// 保留原有终止规则：最后一个汇聚点不参与递归分组。
		for len(convergencePoints) > 1 {
			convergencePoint := convergencePoints[0]
			convergencePoints = convergencePoints[1:]

			selectedPaths, unselectedPaths := FindSubsetPathIncludingTheConvergencePoint(remainingPaths, convergencePoint)
			if len(selectedPaths) == 0 {
				// 经过此点的路径可能已被前面的汇聚点选走，继续检查后续点。
				continue
			}

			HierarchyDivision(selectedPaths, depth+1, finalSegments)
			// 已处理的路径退出本轮候选集，避免被后续汇聚点重复选中。
			remainingPaths = unselectedPaths
		}

		// 5. 无汇聚点或仅剩最后一个汇聚点时，将每条剩余路径直接生成子 segment。
		for _, remainingPath := range remainingPaths {
			subsegment := entities.CreateSegment(virtualEdge.From, virtualEdge.To, depth+1, remainingPath.NodeList)
			*finalSegments = append(*finalSegments, subsegment)
		}
	}
}

// CalculateExcessValue 更新每个节点的 ExcessValue，即包含该节点的路径数。
// 按节点名称匹配，同一条路径即使多次经过该节点，也只计数一次。
func CalculateExcessValue(multipaths []*entities.Path, multipathNodeMapping map[string]*entities.Node) {
	for _, multipathNode := range multipathNodeMapping {
		excessValue := 0
		for _, singlePath := range multipaths {
			for _, nodeInPath := range singlePath.NodeList {
				if nodeInPath.NodeName == multipathNode.NodeName {
					excessValue += 1
					break
				}
			}
		}
		multipathNode.ExcessValue = excessValue
	}
}

// FindNodesWithExcessValueSameAsSource 选出与源节点 ExcessValue 相同的节点。
// 调用前需先计算 ExcessValue；返回结果按节点名称去重。
func FindNodesWithExcessValueSameAsSource(multipaths []*entities.Path, source *entities.Node) map[string]*entities.Node {
	nodesWithExcessValueSameAsSource := map[string]*entities.Node{}
	for _, singlePath := range multipaths {
		for _, node := range singlePath.NodeList {
			if _, ok := nodesWithExcessValueSameAsSource[node.NodeName]; (!ok) && (node.ExcessValue == source.ExcessValue) {
				nodesWithExcessValueSameAsSource[node.NodeName] = node
			}
		}
	}
	return nodesWithExcessValueSameAsSource
}

// SortNodeListBasedOnDistanceToDestination 按到目的节点的最短距离降序排列节点，
// 并将计算结果写入各节点的 Distance 字段，用于组织当前层路径的节点顺序。
func SortNodeListBasedOnDistanceToDestination(nodesMapping map[string]*entities.Node, multipathGraph *simple.DirectedGraph, destination *entities.Node) []*entities.Node {
	var sortedNodes []*entities.Node
	for _, node := range nodesMapping {
		shortestPaths := path.DijkstraFrom(node, multipathGraph)
		distance := shortestPaths.WeightTo(destination.ID())
		node.Distance = distance
		sortedNodes = append(sortedNodes, node)
	}

	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].Distance > sortedNodes[j].Distance
	})

	return sortedNodes
}

// FindVirtualEdgesInHighLevelPath 找出当前层路径中需要展开子路径的相邻节点对。
// 两种情况视为虚拟边：节点间没有直接边，或有直接边且还存在其他可达路径。
func FindVirtualEdgesInHighLevelPath(segment *entities.Segment, multipathGraph *simple.DirectedGraph) []*entities.DirectedEdge {
	var virtualEdges []*entities.DirectedEdge
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			nextNode := segment.Path[index+1]
			// 没有直接边时，这对相邻高层节点之间的连接由下层路径表示。
			if !(multipathGraph.HasEdgeBetween(node.ID(), nextNode.ID())) {
				virtualEdges = append(virtualEdges, &entities.DirectedEdge{
					From: node,
					To:   nextNode,
				})
			}

			if multipathGraph.HasEdgeBetween(node.ID(), nextNode.ID()) {
				// 暂时移除直接边：若仍可达，则还存在需要展开的间接路径。
				multipathGraph.RemoveEdge(node.ID(), nextNode.ID())
				shortestPaths := path.DijkstraFrom(node, multipathGraph)
				_, weight := shortestPaths.To(nextNode.ID())
				if weight != math.Inf(1) {
					virtualEdges = append(virtualEdges, &entities.DirectedEdge{
						From: node,
						To:   nextNode,
					})
				}
				// 无论是否找到间接路径，都恢复刚才移除的边。
				newEdge := multipathGraph.NewEdge(node, nextNode)
				multipathGraph.SetEdge(newEdge)
			}
		}
	}
	return virtualEdges
}

// FindPathsInVirtualEdge 从各条路径中提取虚拟边两端之间的子路径（包含端点），
// 按节点名称序列去重，并保留各子路径首次出现的顺序。
func FindPathsInVirtualEdge(directedEdge *entities.DirectedEdge, paths []*entities.Path) []*entities.Path {
	var uniquePaths []*entities.Path
	var pathsByNodeSequence = make(map[string]*entities.Path)
	for _, singlePath := range paths {
		subpath := &entities.Path{}
		nodeSequence := ""
		recording := false
		for _, node := range singlePath.NodeList {
			if directedEdge.From.NodeName == node.NodeName {
				recording = true
			}

			if recording {
				subpath.NodeList = append(subpath.NodeList, node)
				nodeSequence = nodeSequence + node.NodeName + "->"
			}

			// 先记录终点再关闭标记，使提取结果包含虚拟边的两个端点。
			if directedEdge.To.NodeName == node.NodeName {
				recording = false
			}
		}

		// 不同完整路径可能共享同一段子路径，该子路径仅保留一次。
		if _, ok := pathsByNodeSequence[nodeSequence]; !ok {
			uniquePaths = append(uniquePaths, subpath)
			pathsByNodeSequence[nodeSequence] = subpath
		}
	}

	return uniquePaths
}

// FindConvergencePointsUsingIndegree 更新节点入度并返回入度至少为 2 的汇聚点。
// ExcessValue 仅用于排序，不能用于识别汇聚点，因为分叉点也可能被多条路径经过。
// 返回顺序为 ExcessValue 升序；相同时按节点名称中提取的编号升序。
func FindConvergencePointsUsingIndegree(directedGraph *simple.DirectedGraph, graphNodeMapping map[string]*entities.Node) []*entities.Node {
	var convergencePoints []*entities.Node
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
		graphNodeMapping[edge.To().(*entities.Node).NodeName].Indegree++
		if graphNodeMapping[edge.To().(*entities.Node).NodeName].Indegree >= 2 {
			// 入度超过 2 时仍只加入一次，避免同一汇聚点被重复处理。
			if _, ok := alreadyAddedPoints[edge.To().(*entities.Node).NodeName]; !ok {
				alreadyAddedPoints[edge.To().(*entities.Node).NodeName] = struct{}{}
				convergencePoints = append(convergencePoints, graphNodeMapping[edge.To().(*entities.Node).NodeName])
			}
		}
	}

	sort.Slice(convergencePoints, func(i, j int) bool {
		if convergencePoints[i].ExcessValue < convergencePoints[j].ExcessValue {
			return true
		} else if convergencePoints[i].ExcessValue > convergencePoints[j].ExcessValue {
			return false
		} else {
			firstNodeIndex, _ := extract.NumberFromString(convergencePoints[i].NodeName)
			secondNodeIndex, _ := extract.NumberFromString(convergencePoints[j].NodeName)
			return firstNodeIndex < secondNodeIndex
		}
	})

	return convergencePoints
}

// FindSubsetPathIncludingTheConvergencePoint 按节点名称将路径分为包含和不包含该点的两组。
// 两组均保留输入中的相对顺序，路径及节点对象不复制。
func FindSubsetPathIncludingTheConvergencePoint(paths []*entities.Path, node *entities.Node) ([]*entities.Path, []*entities.Path) {
	var selectedPaths []*entities.Path
	var ignoredPaths []*entities.Path
	for _, singlePath := range paths {
		containsNode := false
		for _, pathNode := range singlePath.NodeList {
			if pathNode.NodeName == node.NodeName {
				containsNode = true
				break
			}
		}
		if containsNode {
			selectedPaths = append(selectedPaths, singlePath)
		} else {
			ignoredPaths = append(ignoredPaths, singlePath)
		}
	}
	return selectedPaths, ignoredPaths
}
