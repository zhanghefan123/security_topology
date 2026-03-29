package calculation

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

func Test() {
	for pathCount := 2; pathCount < 10; pathCount += 2 {
		// 1. 创建图
		atlasComplexTopologyFilePath := "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
		atlasComplexGraph := graph.CreateGraph(atlasComplexTopologyFilePath, "")
		// 2. 图初始化
		err := atlasComplexGraph.Init()
		if err != nil {
			fmt.Printf("create graph error: %v", err)
		}
		atlasComplexGraph.GraphParams.NumberOfPaths = pathCount
		err = atlasComplexGraph.CalculateKShortestPaths()
		if err != nil {
			fmt.Printf("calculate shortest paths failed: %v\n", err)
		}
		fmt.Printf("previous paths length = %d\n", len(atlasComplexGraph.KShortestPaths))
		// 4. 基于创建的路径进行图的生成
		createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(atlasComplexGraph.KShortestPaths)
		atlasComplexGraph.DirectedGraph = createdGraph
		var finalSegmentsTmp []*entities.Segment
		// 6. 打印 edges

		graph.HierarchyDivision(atlasComplexGraph.KShortestPaths, 0, &finalSegmentsTmp)
		//edgeIterator := atlasComplexGraph.RealGraph.Edges()
		//for {
		//	if !(edgeIterator.Next()) {
		//		break
		//	}
		//	edge := edgeIterator.Edge()
		//	fmt.Println(edge.From().(*entities.Node).NodeName, edge.To().(*entities.Node).NodeName)
		//}
	}

}

func GenerateDifferentPaths() error {

	// 3. 计算 length 条最短路
	experimentCount := 100
	headerResultMapping := map[int]map[string][]int{}
	sourceMacResultMapping := map[int]map[string][]int{}
	intermediateMacResultMapping := map[int]map[string][]float64{}
	destinationMacResultMapping := map[int]map[string][]float64{}
	validationProtocolList := []string{"atlas", "opt", "lip"}
	pathLimit := 30
	interval := 2
	arraySize := pathLimit / interval
	for index := 0; index < experimentCount; index++ {
		headerResultMapping[index] = make(map[string][]int)
		sourceMacResultMapping[index] = make(map[string][]int)
		intermediateMacResultMapping[index] = make(map[string][]float64)
		destinationMacResultMapping[index] = make(map[string][]float64)
		// 4. 进行遍历生成
		for innerIndex := 2; innerIndex <= pathLimit; innerIndex += interval {
			// 1. 创建图
			atlasComplexTopologyFilePath := "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
			atlasComplexGraph := graph.CreateGraph(atlasComplexTopologyFilePath, "")
			// 2. 图初始化
			err := atlasComplexGraph.Init()
			if err != nil {
				fmt.Printf("create graph error: %v", err)
			}
			atlasComplexGraph.GraphParams.NumberOfPaths = innerIndex
			err = atlasComplexGraph.CalculateKShortestPaths()
			if err != nil {
				fmt.Printf("calculate shortest paths failed: %v\n", err)
			}
			fmt.Printf("previous paths length = %d\n", len(atlasComplexGraph.KShortestPaths))
			var finalSegmentsTmp []*entities.Segment
			// 4. 基于创建的路径进行图的生成
			createdGraph, nodeMapping, _ := entities.CreateNewGraphFromRealPaths(atlasComplexGraph.KShortestPaths)
			atlasComplexGraph.DirectedGraph = createdGraph

			// 6. 头部结果的计算
			// 5. 进行 segment 的生成
			multipathOptHeaderSize := OptCalculateHeaderSize(atlasComplexGraph.KShortestPaths)
			headerResultMapping[index]["opt"] = append(headerResultMapping[index]["opt"], multipathOptHeaderSize)
			lipHeaderSize := LiPCalculateHeaderSize(atlasComplexGraph.KShortestPaths, 0.00001)
			headerResultMapping[index]["lip"] = append(headerResultMapping[index]["lip"], lipHeaderSize)

			multipathOptSourceMacs := OptCalculateNumberOfSourceMacs(atlasComplexGraph.KShortestPaths)
			sourceMacResultMapping[index]["opt"] = append(sourceMacResultMapping[index]["opt"], multipathOptSourceMacs)
			lipSourceMacs := LiPCalculateNumberOfSourceMacs(atlasComplexGraph.KShortestPaths)
			sourceMacResultMapping[index]["lip"] = append(sourceMacResultMapping[index]["lip"], lipSourceMacs)
			// 8. 进行中间节点的 mac 计算
			atlasIntermediateMacs := AtlasCalculateNumberOfOnPathRouterMacs(atlasComplexGraph.KShortestPaths, atlasComplexGraph.DirectedGraph, nodeMapping, atlasComplexGraph.GraphParams.Source, atlasComplexGraph.GraphParams.Destination)
			intermediateMacResultMapping[index]["atlas"] = append(intermediateMacResultMapping[index]["atlas"], atlasIntermediateMacs)
			multipathOptIntermediateMacs := OptCalculateNumberOfOnPathRouterMacs()
			intermediateMacResultMapping[index]["opt"] = append(intermediateMacResultMapping[index]["opt"], float64(multipathOptIntermediateMacs))
			lipIntermediateMacs := LiPCalculateNumberOfOnPathRouterMacs()
			intermediateMacResultMapping[index]["lip"] = append(intermediateMacResultMapping[index]["lip"], float64(lipIntermediateMacs))
			// 9. 进行目的节点的 mac 计算

			atlasDestinationMacs := AtlasCalculateNumberOfDestinationMacs(atlasComplexGraph.KShortestPaths, atlasComplexGraph.GraphParams.Destination, atlasComplexGraph.DirectedGraph, nodeMapping)
			destinationMacResultMapping[index]["atlas"] = append(destinationMacResultMapping[index]["atlas"], float64(atlasDestinationMacs))
			multipathOptDestinationMacs := OptCalculateNumberOfDestinationMacs(atlasComplexGraph.KShortestPaths)
			destinationMacResultMapping[index]["opt"] = append(destinationMacResultMapping[index]["opt"], multipathOptDestinationMacs)
			lipDestinationMacs := LiPCalculateNumberOfDestinationMacs()
			destinationMacResultMapping[index]["lip"] = append(destinationMacResultMapping[index]["lip"], float64(lipDestinationMacs))

			// 一旦 HierarchyDivision 会对于 path 内部的 node List 进行修改, 也随之修改了原始的 graph 之中的 node list
			graph.HierarchyDivision(atlasComplexGraph.KShortestPaths, 0, &finalSegmentsTmp)
			atlasHeaderSize := AtlasCalculateHeaderSize(finalSegmentsTmp)
			headerResultMapping[index]["atlas"] = append(headerResultMapping[index]["atlas"], atlasHeaderSize)
			atlasSourceMacs := AtlasCalculateNumberOfSourceMacs(finalSegmentsTmp)
			sourceMacResultMapping[index]["atlas"] = append(sourceMacResultMapping[index]["atlas"], atlasSourceMacs)
		}
	}
	// 进行结果汇总, 计算平均值
	// -------------------------------------------------------------------------------------------------------------
	averageHeaderResultMapping := make(map[string][]float64)
	averageSourceMacResultMapping := make(map[string][]float64)
	averageIntermediateMacResultMapping := make(map[string][]float64)
	averageDestinationMacResultMapping := make(map[string][]float64)
	for _, validationProtocol := range validationProtocolList {
		for hopIndex := 0; hopIndex < arraySize; hopIndex += 1 {
			totalHeaderSize := 0
			totalSourceMacs := 0
			totalIntermediateMacs := 0.0
			totalDestinationMacs := 0.0
			for experimentIndex := 0; experimentIndex < experimentCount; experimentIndex++ {
				totalHeaderSize += headerResultMapping[experimentIndex][validationProtocol][hopIndex]
				totalSourceMacs += sourceMacResultMapping[experimentIndex][validationProtocol][hopIndex]
				totalIntermediateMacs += intermediateMacResultMapping[experimentIndex][validationProtocol][hopIndex]
				totalDestinationMacs += destinationMacResultMapping[experimentIndex][validationProtocol][hopIndex]
			}
			averageHeaderSize := float64(totalHeaderSize) / float64(experimentCount)
			averageSourceMacs := float64(totalSourceMacs) / float64(experimentCount)
			averageIntermediateMacs := totalIntermediateMacs / float64(experimentCount)
			averageDestinationMacs := totalDestinationMacs / float64(experimentCount)
			averageHeaderResultMapping[validationProtocol] = append(averageHeaderResultMapping[validationProtocol], averageHeaderSize)
			averageSourceMacResultMapping[validationProtocol] = append(averageSourceMacResultMapping[validationProtocol], averageSourceMacs)
			averageIntermediateMacResultMapping[validationProtocol] = append(averageIntermediateMacResultMapping[validationProtocol], averageIntermediateMacs)
			averageDestinationMacResultMapping[validationProtocol] = append(averageDestinationMacResultMapping[validationProtocol], averageDestinationMacs)
		}
	}
	// -------------------------------------------------------------------------------------------------------------

	// 进行结果的打印
	// -------------------------------------------------------------------------------------------------------------
	for validationProtocol, averageHeaderSizeList := range averageHeaderResultMapping {
		finalString := fmt.Sprintf("validation protocol: %s --> average header size: [", validationProtocol)
		for index, averageHeaderSize := range averageHeaderSizeList {
			if index != (len(averageHeaderSizeList) - 1) {
				finalString += fmt.Sprintf("%f,", averageHeaderSize)
			} else {
				finalString += fmt.Sprintf("%f", averageHeaderSize)
			}
		}
		finalString += "]"
		fmt.Println(finalString)
	}
	// -------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------
	for validationProtocol, averageSourceMacs := range averageSourceMacResultMapping {
		finalString := fmt.Sprintf("validation protocol: %s --> source macs: [", validationProtocol)
		for index, averageSourceMac := range averageSourceMacs {
			if index != (len(averageSourceMacs) - 1) {
				finalString += fmt.Sprintf("%f,", averageSourceMac)
			} else {
				finalString += fmt.Sprintf("%f", averageSourceMac)
			}
		}
		finalString += "]"
		fmt.Println(finalString)
	}
	// -------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------
	for validationProtocol, averageIntermediateMacs := range averageIntermediateMacResultMapping {
		finalString := fmt.Sprintf("validation protocol: %s --> intermediate macs: [", validationProtocol)
		for index, averageIntermediateMac := range averageIntermediateMacs {
			if index != (len(averageIntermediateMacs) - 1) {
				finalString += fmt.Sprintf("%f,", averageIntermediateMac)
			} else {
				finalString += fmt.Sprintf("%f", averageIntermediateMac)
			}
		}
		finalString += "]"
		fmt.Println(finalString)
	}
	// -------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------
	for validationProtocol, averageDestinationMacs := range averageDestinationMacResultMapping {
		finalString := fmt.Sprintf("validation protocol: %s --> destination macs: [", validationProtocol)
		for index, averageDestinationMac := range averageDestinationMacs {
			if index != (len(averageDestinationMacs) - 1) {
				finalString += fmt.Sprintf("%f,", averageDestinationMac)
			} else {
				finalString += fmt.Sprintf("%f", averageDestinationMac)
			}
		}
		finalString += "]"
		fmt.Println(finalString)
	}
	// -------------------------------------------------------------------------------------------------------------

	return nil
}

func CalculateIndices() {
	var atlasHeaderSizeList []int
	var multipathOptHeaderSizeList []int
	var lipHeaderSizeList []int
	for index := 2; index <= 20; index += 2 {
		filePath := fmt.Sprintf("C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_%d.txt", index)
		paths, segments, _, _ := graph.GenerateAtlasPathsAndSegmentsViaPathsFile(filePath)
		atlasHeaderSize := AtlasCalculateHeaderSize(segments)
		atlasHeaderSizeList = append(atlasHeaderSizeList, atlasHeaderSize)
		multipathOptHeaderSize := OptCalculateHeaderSize(paths)
		multipathOptHeaderSizeList = append(multipathOptHeaderSizeList, multipathOptHeaderSize)
		lipHeaderSize := LiPCalculateHeaderSize(paths, 0.00001)
		lipHeaderSizeList = append(lipHeaderSizeList, lipHeaderSize)
	}
}
