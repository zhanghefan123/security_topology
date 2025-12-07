package graph

import (
	"fmt"
	"strings"
)

// GeneratePaths 进行指定数量的路径的 segment 的生成
func GeneratePaths(numberOfPath int) {
	// 1. 创建图
	atlasComplexTopologyFilePath := "C:\\Users\\zhf\\Desktop\\zhf_projects\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
	atlasComplexGraph := CreateGraph(atlasComplexTopologyFilePath)
	// 2. 图初始化
	err := atlasComplexGraph.Init()
	if err != nil {
		fmt.Printf("create graph error: %v", err)
	}
	// 3. 计算 length 条最短路
	atlasComplexGraph.GraphParams.NumberOfPaths = numberOfPath
	err = atlasComplexGraph.CalculateKShortestPaths()
	if err != nil {
		fmt.Printf("calculate shortest paths failed: %v\n", err)
	}
	// 4. 进行分段
	HierarchyDivision(atlasComplexGraph.KShortestPaths, 0)
	// 5. 将分段进行打印
	fmt.Printf("------------------------------------------\n")
	for index, segment := range finalSegments {
		segmentId := index + 1
		segment.Id = segmentId
		segment.PathStr = SegmentToString(segment)
		fmt.Println(segment.PathStr)
	}
	fmt.Printf("--------------------------------------------\n")
	// 6. 将 segments 按节点进行分类
	nodeSegments := map[string][]*Segment{}
	for nodeName, _ := range atlasComplexGraph.NameToNodeMapping {
		for _, segment := range finalSegments {
			if strings.Contains(segment.PathStr, nodeName) {
				nodeSegments[nodeName] = append(nodeSegments[nodeName], segment)
			}
		}
	}
	// 7. 将 nodeSegments 进行打印
	for nodeName, segments := range nodeSegments {
		fmt.Printf("---------------------%s---------------------\n", nodeName)
		for _, segment := range segments {
			fmt.Println(segment.PathStr)
		}
		fmt.Printf("---------------------%s---------------------\n", nodeName)
	}
}
