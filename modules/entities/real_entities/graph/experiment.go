package graph

import (
	"fmt"
)

// ExperimentWithDifferentPaths 在不同多路径数下的 lip 和 Atlas 的各项性能测试
func ExperimentWithDifferentPaths() {
	var lipHeaderSizeList []float64
	var lipSourceMacList []float64
	var lipOnpathMacList []float64
	var lipDestMacList []float64

	var atlasHeaderSizeList []float64
	var atlasSourceMacList []float64
	var atlasOnpathMacList []float64
	var atlasDestMacList []float64

	var optHeaderSizeList []float64
	var optSourceMacList []float64
	var optOnpathMacList []float64
	var optDestMacList []float64

	hcu := &HeaderCalculationUnit{}
	mcu := &MacsCalculationUnit{}

	fpr := 0.00001
	// 现在要求后续的 path 只能在之前的 path 之上进行递增(前面的path必须保留)
	for numberOfPaths := range 20 {
		optHeaderSizeTotal := 0.0
		optSourceMacTotal := 0.0
		optOnPathMacTotal := 0.0
		optDestMacTotal := 0.0

		lipHeaderSizeTotal := 0.0
		lipSourceMacTotal := 0.0
		lipOnPathMacTotal := 0.0
		lipDestMacTotal := 0.0

		atlasHeaderSizeTotal := 0.0
		atlasSourceMacTotal := 0.0
		atlasOnPathMacTotal := 0.0
		atlasDestMacTotal := 0.0
		iterations := 100
		for _ = range iterations {
			finalSegments = []*Segment{}
			// 进行图的初始化
			atlasComplexTopologyFilePath := "C:\\Users\\zhf\\Desktop\\zhf_projects\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
			atlasComplexGraph := CreateGraph(atlasComplexTopologyFilePath)
			err := atlasComplexGraph.Init()
			if err != nil {
				fmt.Printf("create atlas simple graph failed: %v\n", err)
			}
			atlasComplexGraph.GraphParams.NumberOfPaths = numberOfPaths + 1
			err = atlasComplexGraph.CalculateKShortestPaths()
			if err != nil {
				fmt.Printf("calculate shortest paths failed: %v\n", err)
			}

			PrintPaths(atlasComplexGraph.KShortestPaths)
			// 计算 atlas 的首部大小
			HierarchyDivision(atlasComplexGraph.KShortestPaths, 0)
			headerSize := hcu.AtlasCalculateHeaderSize(finalSegments)
			atlasHeaderSizeTotal += float64(headerSize)
			// 计算 atlas 的各个节点的 mac 计算次数
			atlasSourceMacs := mcu.AtlasCalculateNumberOfSourceMacs(finalSegments)
			atlasSourceMacTotal += float64(atlasSourceMacs)
			createdGraph, multipathNodeMapping, _ := CreateNewGraphFromRealPaths(atlasComplexGraph.KShortestPaths)
			CalculateIndegree(createdGraph, multipathNodeMapping)
			atlasAverageOnPathRouterMacs := mcu.AtlasCalculateNumberOfOnPathRouterMacs(multipathNodeMapping,
				atlasComplexGraph.GraphParams.Source,
				atlasComplexGraph.GraphParams.Destination)
			atlasOnPathMacTotal += atlasAverageOnPathRouterMacs
			atlasDestinationMacs := mcu.AtlasCalculateNumberOfDestinationMacs(finalSegments[0])
			for index, finalSegment := range finalSegments {
				finalString := ""
				for _, node := range finalSegment.Path {
					name := node.NodeName
					finalString += fmt.Sprintf("%s->", name)
				}
				fmt.Printf("segment-:%d | depth: %d | ->%s\n", index, finalSegment.Depth, finalString)
			}
			atlasDestMacTotal += float64(atlasDestinationMacs)

			// 计算 lip 的首部大小
			lipHeaderSize := hcu.LiPCalculateHeaderSize(atlasComplexGraph.KShortestPaths, fpr)
			lipHeaderSizeTotal += float64(lipHeaderSize)

			lipSourceMacs := mcu.LiPCalculateNumberOfSourceMacs(atlasComplexGraph.KShortestPaths)
			lipSourceMacTotal += float64(lipSourceMacs)
			lipOnPathRouterMacs := mcu.LiPCalculateNumberOfOnPathRouterMacs()
			lipOnPathMacTotal += float64(lipOnPathRouterMacs)
			lipDestinationMacs := mcu.LiPCalculateNumberOfDestinationMacs()
			lipDestMacTotal += float64(lipDestinationMacs)

			optHeaderSize := hcu.OptCalculateHeaderSize(atlasComplexGraph.KShortestPaths)
			optHeaderSizeTotal += optHeaderSize
			optSourceMacs := mcu.OptCalculateNumberOfSourceMacs(atlasComplexGraph.KShortestPaths)
			optSourceMacTotal += float64(optSourceMacs)
			optOnpathMacs := mcu.OptCalculateNumberOfOnPathRouterMacs()
			optOnPathMacTotal += float64(optOnpathMacs)
			optDestMacs := mcu.OptCalculateNumberOfDestinationMacs(atlasComplexGraph.KShortestPaths)
			optDestMacTotal += optDestMacs
		}
		atlasHeaderSizeList = append(atlasHeaderSizeList, atlasHeaderSizeTotal/float64(iterations))
		atlasSourceMacList = append(atlasSourceMacList, atlasSourceMacTotal/float64(iterations))
		atlasOnpathMacList = append(atlasOnpathMacList, atlasOnPathMacTotal/float64(iterations))
		atlasDestMacList = append(atlasDestMacList, atlasDestMacTotal/float64(iterations))

		lipHeaderSizeList = append(lipHeaderSizeList, lipHeaderSizeTotal/float64(iterations))
		lipSourceMacList = append(lipSourceMacList, lipSourceMacTotal/float64(iterations))
		lipOnpathMacList = append(lipOnpathMacList, lipOnPathMacTotal/float64(iterations))
		lipDestMacList = append(lipDestMacList, lipDestMacTotal/float64(iterations))

		optHeaderSizeList = append(optHeaderSizeList, optHeaderSizeTotal/float64(iterations))
		optSourceMacList = append(optSourceMacList, optSourceMacTotal/float64(iterations))
		optOnpathMacList = append(optOnpathMacList, optOnPathMacTotal/float64(iterations))
		optDestMacList = append(optDestMacList, optDestMacTotal/float64(iterations))
	}

	fmt.Printf("opt header size: %v\n", ListToStringSimple(optHeaderSizeList))
	fmt.Printf("lip headers size: %v\n", ListToStringSimple(lipHeaderSizeList))
	fmt.Printf("atlas headers size: %v\n", ListToStringSimple(atlasHeaderSizeList))

	fmt.Println("lip source macs: ", ListToStringSimple(lipSourceMacList))
	fmt.Printf("atlas source macs: %v\n", ListToStringSimple(atlasSourceMacList))
	fmt.Printf("opt source macs: %v\n", ListToStringSimple(optSourceMacList))

	fmt.Println("lip on-path router macs", ListToStringSimple(lipOnpathMacList))
	fmt.Println("atlas on-path router macs: ", ListToStringSimple(atlasOnpathMacList))
	fmt.Println("opt on-path router macs:", ListToStringSimple(optOnpathMacList))

	fmt.Println("lip destination macs", ListToStringSimple(lipDestMacList))
	fmt.Println("atlas destination macs: ", ListToStringSimple(atlasDestMacList))
	fmt.Println("opt destination macs: ", ListToStringSimple(optDestMacList))
}

// ListToStringSimple 使用 fmt.Sprintf 简化
func ListToStringSimple[T any](list []T) string {
	if len(list) == 0 {
		return ""
	}

	result := ""
	for i, v := range list {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%v", v)
	}
	return result
}

func Start() {
	//ExperimentWithDifferentPaths()
	GeneratePaths(3)
}
