package calculation

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

// ------------------------------------ opt 计算函数 ------------------------------------

func OptCalculateNumberOfSourceMacs(paths []*entities.Path) int {
	pvfCount := 0
	calculatedPVF := map[string]struct{}{}
	for _, path := range paths {
		currentPathString := ""
		for index, node := range path.NodeList {
			if index < (len(path.NodeList) - 1) {
				currentPathString += fmt.Sprintf("%s->", node.NodeName)
				if _, ok := calculatedPVF[currentPathString]; !ok {
					calculatedPVF[currentPathString] = struct{}{}
					pvfCount += 1
				}
			}
		}
	}

	opvCount := 0
	hasOpvs := map[string]struct{}{}
	// 进行所有的路径的遍历
	for _, path := range paths {
		currentPathString := ""
		for index, node := range path.NodeList { //  1 2 3 1->2 2->3
			if index < (len(path.NodeList) - 1) {
				nextIndex := index + 1
				nextNode := path.NodeList[nextIndex]
				currentPathString += fmt.Sprintf("%s->%s->", node.NodeName, nextNode.NodeName)
				if _, ok := hasOpvs[currentPathString]; !ok {
					hasOpvs[currentPathString] = struct{}{}
					opvCount += 1
				}
			}
		}
	}

	return pvfCount + opvCount
}

func OptCalculateNumberOfOnPathRouterMacs() int {
	return 2
}

func OptCalculateNumberOfDestinationMacs(paths []*entities.Path) float64 {
	total := 0.0
	for _, path := range paths {
		total += float64(len(path.NodeList))
	}
	return total / float64(len(paths))
}

// ------------------------------------ opt 计算函数 ------------------------------------

// ------------------------------------ atlas 计算函数 ------------------------------------

// AtlasCalculateNumberOfSourceMacs atlas 计算源节点计算的 MAC 的次数
func AtlasCalculateNumberOfSourceMacs(segments []*entities.Segment) int {
	finalResult := 0
	for _, singleSegment := range segments {
		finalResult += (len(singleSegment.Path) - 1) * 2
	}
	return finalResult
}

// AtlasCalculateNumberOfOnPathRouterMacs 计算中间节点的 MAC 次数
func AtlasCalculateNumberOfOnPathRouterMacs(paths []*entities.Path, directedGraph *simple.DirectedGraph, graphNodeMapping map[string]*entities.Node, source, destination string) float64 {

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
	}

	// 对于每条路径进行一个平均的 mac 的计算
	pathMacAverageTotal := 0.0
	for _, path := range paths {
		pathMacTotal := 0.0
		for index, node := range path.NodeList {
			if (index != 0) && (index != (len(path.NodeList) - 1)) {
				if graphNodeMapping[node.NodeName].Indegree >= 2 {
					pathMacTotal += 3
				} else {
					pathMacTotal += 2
				}
			}
		}
		pathMacAvg := pathMacTotal / float64(len(path.NodeList)-2)
		pathMacAverageTotal += pathMacAvg
	}

	// 计算最终的结果
	finalAverage := pathMacAverageTotal / float64(len(paths))

	return finalAverage
}

// AtlasCalculateNumberOfDestinationMacs 计算目的节点的 MAC 次数
func AtlasCalculateNumberOfDestinationMacs(paths []*entities.Path, destinationNodeName string, directedGraph *simple.DirectedGraph, graphNodeMapping map[string]*entities.Node) int {
	//fmt.Println("mapping:", graphNodeMapping)

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
		fmt.Println(edge.From().(*entities.Node).NodeName, edge.To().(*entities.Node).NodeName)
	}

	// 看看目的节点的度数
	destinationNode := graphNodeMapping[destinationNodeName]
	if destinationNode.Indegree >= 2 {
		return 2
	} else {
		entities.PrintPaths(paths)
		return 1
	}
}

// ------------------------------------ atlas 计算函数 ------------------------------------

// ------------------------------------ lip 计算函数 ------------------------------------

func LiPCalculateNumberOfSourceMacs(paths []*entities.Path) int {
	macCount := 0
	calculatedMacs := map[string]struct{}{}
	// 进行所有的路径的遍历
	for _, path := range paths {
		currentPathString := ""
		for index, node := range path.NodeList {
			if index != len(path.NodeList)-1 {
				nextIndex := index + 1
				nextNode := path.NodeList[nextIndex]
				currentPathString += fmt.Sprintf("%s->%s->", node.NodeName, nextNode.NodeName)
				if _, ok := calculatedMacs[currentPathString]; !ok {
					calculatedMacs[currentPathString] = struct{}{}
					macCount += 1
				}
			}
		}
	}
	fmt.Println(calculatedMacs)
	return macCount
}

// 1->2->3->4 计算 HVF 1->2 2->3 3->4

func LiPCalculateNumberOfDestinationMacs() int {
	return 1
}

func LiPCalculateNumberOfOnPathRouterMacs() int {
	return 1
}

// ------------------------------------ lip 计算函数 ------------------------------------
