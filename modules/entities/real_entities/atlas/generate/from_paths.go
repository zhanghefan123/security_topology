package generate

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/entities"
	atlasinit "zhanghefan123/security_topology/modules/entities/real_entities/atlas/init"
	"zhanghefan123/security_topology/utils/extract"
)

// GenerateMultipathSelirMultipathsViaPathsFile 通过路径文件生成 fast selir 多路径
func GenerateMultipathSelirMultipathsViaPathsFile(pathsFile string) (*atlasinit.Graph, []*entities.Path, map[string]*entities.Node, string, string) {
	// 1. 创建图
	multipathSelirComplexGraph := atlasinit.CreateGraph("", pathsFile)
	// 2. 从路径之中进行初始化
	err := multipathSelirComplexGraph.InitFromPathsFile()
	if err != nil {
		fmt.Printf("init graph from paths failed due to %v", err)
	}
	// 3. source 和 destination
	source := multipathSelirComplexGraph.KShortestPaths[0].NodeList[0]
	destination := multipathSelirComplexGraph.KShortestPaths[0].NodeList[len(multipathSelirComplexGraph.KShortestPaths[0].NodeList)-1]
	// 4. 结果返回
	return multipathSelirComplexGraph, multipathSelirComplexGraph.KShortestPaths, multipathSelirComplexGraph.NameToNodeMapping, source.NodeName, destination.NodeName
}

// GenerateAtlasPathsAndSegmentsViaPathsFile 通过路径文件生成 Atlas 多路径
func GenerateAtlasPathsAndSegmentsViaPathsFile(pathsFile string) ([]*entities.Path, []*entities.Segment, map[string][]*entities.Segment, *atlasinit.Params) {
	// 1. 创建图
	atlasComplexGraph := atlasinit.CreateGraph("", pathsFile)
	// 2. 从路径之中进行初始化
	err := atlasComplexGraph.InitFromPathsFile()
	if err != nil {
		fmt.Printf("init graph from paths failed due to %v", err)
	}
	// 3. 进行分段
	var finalSegmentsTmp []*entities.Segment
	HierarchyDivision(atlasComplexGraph.KShortestPaths, 0, &finalSegmentsTmp)
	// 4. 进行源头和目的节点的获取
	source := atlasComplexGraph.KShortestPaths[0].NodeList[0]
	destination := atlasComplexGraph.KShortestPaths[0].NodeList[len(atlasComplexGraph.KShortestPaths[0].NodeList)-1]
	atlasComplexGraph.GraphParams.Source = source.NodeName
	atlasComplexGraph.GraphParams.Destination = destination.NodeName
	sourceIndex, _ := extract.NumberFromString(source.NodeName)
	destinationIndex, _ := extract.NumberFromString(destination.NodeName)
	atlasComplexGraph.GraphParams.SourceIndex = sourceIndex
	atlasComplexGraph.GraphParams.DestinationIndex = destinationIndex
	// 5. 给所有的 segments 加上编号 (并且进行最大的深度的查找)
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
	// 6. 遍历所有 segments 进行最终目的节点的设置
	for _, segment := range finalSegmentsTmp {
		segment.FinalDestinationIndex = atlasComplexGraph.GraphParams.DestinationIndex
	}
	// 7. 进行所有的 segments 的遍历找到他们的 parent
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
	// 8. 按照节点的不同对 segment 进行划分
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
	// 9. 返回结果
	return atlasComplexGraph.KShortestPaths, finalSegmentsTmp, nodeSegmentsMapping, atlasComplexGraph.GraphParams
}
