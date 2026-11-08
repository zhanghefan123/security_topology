package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/generate"
	atlasinit "zhanghefan123/security_topology/modules/entities/real_entities/atlas/init"
)

func TestInitFromPaths(t *testing.T) {
	// 1. 创建图
	atlasComplexGraph := atlasinit.CreateGraph("", "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_4.txt")
	// 2. 从路径之中进行初始化
	err := atlasComplexGraph.InitFromPathsFile()
	if err != nil {
		fmt.Printf("init graph from paths failed due to %v", err)
	}
	// 3. 打印路径
	for _, path := range atlasComplexGraph.KShortestPaths {
		fmt.Println(path)
	}
}

func TestCreateSegmentsFromPaths(t *testing.T) {
	paths, segments, _, _ := generate.GenerateAtlasPathsAndSegmentsViaPathsFile("C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_10.txt")
	for _, path := range paths {
		fmt.Println(path)
	}
	for _, segment := range segments {
		fmt.Println(segment)
	}
}
