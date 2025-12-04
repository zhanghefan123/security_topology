package graph

import (
	"fmt"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
)

var finalSegments []*Segment

// HierarchyDivision 进行分层切割 (paths 代表多路径, depth 代表深度)
func HierarchyDivision(paths []*Path, depth int) {
	// step 5 Build directed mulitpath graph using paths
	createdGraph, multipathNodeMapping, sourceAndDest := CreateNewGraphFromRealPaths(paths)
	// step 6-7 traverse each node in graph and calculate their excess value
	CalculateExcessValue(paths, multipathNodeMapping)
	// step 8-9 进行 segment 的创建
	segment := CreateSegment(sourceAndDest.Source, sourceAndDest.Destination, depth, nil)
	// step 10-12 找到和源具有相同 excess value 的节点并入到其中
	nodeListWithSameExcessValue := FindNodesWithExcessValueSameAsSource(paths, sourceAndDest.Source)
	segment.Path = nodeListWithSameExcessValue
	// step 13 将 segment 并入到 finalSegments 之中
	finalSegments = append(finalSegments, segment)
	// step 14 进行所有在 segment 之中的但是不在 graph 之中的边的遍历
	virtualEdges := FindEdgeWithinSegmentButNotInGraph(segment, createdGraph)
	for _, virtualEdge := range virtualEdges {
		// step 15 找到 virtual edge 的连接的子路径
		subpaths := FindPathsInVirtualEdge(virtualEdge, paths)
		// step 16 通过这些 subpaths 构建一个 subgraph
		_, subGraphNodeMapping, _ := CreateNewGraphFromRealPaths(subpaths)
		// step 17 找到大于等于2 的入度的节点
		CalculateExcessValue(subpaths, subGraphNodeMapping)
		convergencePoints := FindConvergencePoints(subGraphNodeMapping)
		// step 18 移除掉最后的节点
		convergencePoints = RemoveVirtualEdgeDestinationFromConvergencePoints(convergencePoints, virtualEdge.To)
		// step 19 判断是否是空集
		if 0 == len(convergencePoints) {
			// step 20: 进行所有 subpaths 的遍历
			for _, subpath := range subpaths {
				// step 21-23 创建一个新的 segment:
				subPathSegment := CreateSegment(virtualEdge.From, virtualEdge.To, depth+1, subpath.NodeList)
				finalSegments = append(finalSegments, subPathSegment)
			}
		} else { // step 24-29
			for {
				// step 26
				firstConvergencePoint := convergencePoints[0]
				// step 27
				selectedPaths, ignoredPaths := FindSubsetPathIncludingANode(subpaths, firstConvergencePoint)
				// step 28 进行递归调用
				HierarchyDivision(selectedPaths, depth+1)
				// step 29 调用完成后将节点移除, 路径也进行移除
				convergencePoints = convergencePoints[1:]
				subpaths = ignoredPaths
				// 视情况进行 break
				if len(convergencePoints) == 0 {
					break
				}
			}
		}

	}
}

func Start() {
	multipathFilePath := filepath.Join(configs.TopConfiguration.PathConfig.ResourcesPath, "multipath/multipath.yml")
	pathsInStr, err := ResolveMultiPathFile(multipathFilePath)
	if err != nil {
		fmt.Printf("resolve multipath file failed: %v", err)
	}
	// 在这一步完成节点的创建以及边的创建
	paths := CreatePathsFromStrPaths(pathsInStr)
	HierarchyDivision(paths, 0)
}
