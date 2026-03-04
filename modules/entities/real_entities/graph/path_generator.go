package graph

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

// GeneratePaths 进行指定数量的路径的 segment 的生成
func GeneratePaths(numberOfPath int) {
	// 1. 创建图
	atlasComplexTopologyFilePath := "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
	atlasComplexGraph := CreateGraph(atlasComplexTopologyFilePath, "")
	// 2. 图初始化
	err := atlasComplexGraph.Init()
	if err != nil {
		fmt.Printf("create graph error: %v", err)
	}

	// 3. 计算 length 条最短路
	atlasComplexGraph.GraphParams.NumberOfPaths = numberOfPath
	// ---------------------------首先计算一次最短路径---------------------------------------
	err = atlasComplexGraph.CalculateKShortestPaths()
	if err != nil {
		fmt.Printf("calculate shortest paths failed: %v\n", err)
	}
	// ---------------------------首先计算一次最短路径---------------------------------------
	createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(atlasComplexGraph.KShortestPaths)
	atlasComplexGraph.DirectedGraph = createdGraph
	CalculateNoLoopKshortestPaths(atlasComplexGraph)
	for _, path := range atlasComplexGraph.KShortestPaths {
		entities.PrintPath(path, 0)
	}

	// 4. 进行分段
	var finalSegmentsTmp []*entities.Segment
	HierarchyDivision(atlasComplexGraph.KShortestPaths, 0, &finalSegmentsTmp)
	// 5. 将分段进行打印
	fmt.Printf("------------------------------------------\n")
	for index, segment := range finalSegmentsTmp {
		segmentId := index + 1
		segment.Id = segmentId
		segment.PathStr = entities.SegmentToString(segment)
		fmt.Printf("segment, %s\n", segment.PathStr)
	}
	fmt.Printf("--------------------------------------------\n")
}
