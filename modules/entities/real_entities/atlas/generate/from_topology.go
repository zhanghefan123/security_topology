package generate

import (
	"fmt"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/entities"
	atlasinit "zhanghefan123/security_topology/modules/entities/real_entities/atlas/init"
)

// GenerateMultipathSelirMultipaths 生成多路径
func GenerateMultipathSelirMultipaths(numberOfPath int) ([]*entities.Path, map[string]*entities.Node, string, string) {
	// 1. 创建图
	resourcesPath := configs.TopConfiguration.PathConfig.ResourcesPath
	topologyFilePath := filepath.Join(resourcesPath, "./multipath/atlas_simple_topology.json")
	atlasComplexGraph := atlasinit.CreateGraph(topologyFilePath, "")
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

	// 4. 进行路的返回
	return atlasComplexGraph.KShortestPaths, atlasComplexGraph.NameToNodeMapping, atlasComplexGraph.GraphParams.Source, atlasComplexGraph.GraphParams.Destination
}

// GenerateAtlasPathsAndSegments 生成分段
func GenerateAtlasPathsAndSegments(numberOfPath int) ([]*entities.Path, []*entities.Segment, map[string][]*entities.Segment, *atlasinit.Params) {
	// 1. 创建图
	resourcesPath := configs.TopConfiguration.PathConfig.ResourcesPath
	topologyFilePath := filepath.Join(resourcesPath, fmt.Sprintf("./multipath/%s.json", configs.TopConfiguration.StartTopologyParams.PathValidationParams.MultipathConfig.ValidationTopology))
	atlasComplexGraph := atlasinit.CreateGraph(topologyFilePath, "")
	// 2. 图初始化
	err := atlasComplexGraph.Init()
	if err != nil {
		fmt.Printf("create graph error: %v", err)
	}
	fmt.Printf("number of paths: %d\n", numberOfPath)

	// 3. 计算 length 条最短路
	atlasComplexGraph.GraphParams.NumberOfPaths = numberOfPath
	err = atlasComplexGraph.CalculateKShortestPaths()
	if err != nil {
		fmt.Printf("calculate shortest paths failed: %v\n", err)
	}
	createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(atlasComplexGraph.KShortestPaths)
	atlasComplexGraph.DirectedGraph = createdGraph
	atlasinit.CalculateNoLoopKshortestPaths(atlasComplexGraph)
	fmt.Printf("shortest path length = %d\n", len(atlasComplexGraph.KShortestPaths))

	// 4. 进行所有的 path 的打印
	fmt.Println("----------------------------------------")
	for _, path := range atlasComplexGraph.KShortestPaths {
		fmt.Println(path)
	}
	fmt.Println("----------------------------------------")
	// 4. 进行分段
	var finalSegmentsTmp []*entities.Segment
	HierarchyDivision(atlasComplexGraph.KShortestPaths, 0, &finalSegmentsTmp)

	// 5. 给所有的 segments 加上编号
	idToSegmentMapping := make(map[int]*entities.Segment)
	maxDepth := -1
	for index, segment := range finalSegmentsTmp {
		segmentId := index + 1
		segment.Id = segmentId
		segment.PathStr = entities.SegmentToString(segment)
		idToSegmentMapping[segment.Id] = segment
		if segment.Depth > maxDepth {
			maxDepth = segment.Depth
		}
	}

	// 遍历所有 segments 进行最终目的节点的设置
	for _, segment := range finalSegmentsTmp {
		segment.FinalDestinationIndex = atlasComplexGraph.GraphParams.DestinationIndex
	}

	// 6. 进行所有的 segments 的遍历找到他们的 parent
	for index, segment := range finalSegmentsTmp {
		if index == 0 {
			segment.ParentId = segment.Id
		} else {
			for innerIndex := index; innerIndex >= 0; innerIndex-- {
				if segment.Depth == (finalSegmentsTmp[innerIndex].Depth + 1) {
					segment.ParentId = finalSegmentsTmp[innerIndex].Id
					break
				}
			}
		}
	}

	// 7. 按照节点的不同对 segment 进行划分
	nodeSegmentsMapping := make(map[string][]*entities.Segment)

	for nodeName, _ := range atlasComplexGraph.NameToNodeMapping {
		for _, segment := range finalSegmentsTmp {
			// 进行 segment 内部的遍历
			for _, node := range segment.Path {
				if node.NodeName == nodeName {
					nodeSegmentsMapping[nodeName] = append(nodeSegmentsMapping[nodeName], segment)
				}
			}
		}
	}

	// 8. 返回结果
	return atlasComplexGraph.KShortestPaths, finalSegmentsTmp, nodeSegmentsMapping, atlasComplexGraph.GraphParams
}
