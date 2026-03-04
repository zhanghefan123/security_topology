package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

func TestInitFromPaths(t *testing.T) {
	// 1. 创建图
	atlasComplexGraph := graph.CreateGraph("", "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_4.txt")
	// 2. 从路径之中进行初始化
	err := atlasComplexGraph.InitFromPathsFile()
	if err != nil {
		fmt.Printf("init graph from paths failed due to %v", err)
	}
	// 3. 打印路径
	for index, path := range atlasComplexGraph.KShortestPaths {
		entities.PrintPath(path, index)
	}
}

func TestCreateSegmentsFromPaths(t *testing.T) {
	paths, segments, _, _ := graph.GenerateAtlasPathsAndSegmentsViaPathsFile("C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_10.txt")
	for index, path := range paths {
		entities.PrintPath(path, index)
	}
	entities.PrintSegments(segments)
}
