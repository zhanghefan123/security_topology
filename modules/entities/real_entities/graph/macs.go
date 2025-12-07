package graph

import (
	"fmt"
)

type MacsCalculationUnit struct {
}

// ------------------------------------ opt 计算函数 ------------------------------------

func (mcu *MacsCalculationUnit) OptCalculateNumberOfSourceMacs(paths []*Path) int {
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

func (mcu *MacsCalculationUnit) OptCalculateNumberOfOnPathRouterMacs() int {
	return 2
}

func (mcu *MacsCalculationUnit) OptCalculateNumberOfDestinationMacs(paths []*Path) float64 {
	total := 0.0
	for _, path := range paths {
		total += float64(len(path.NodeList))
	}
	return total / float64(len(paths))
}

// ------------------------------------ opt 计算函数 ------------------------------------

// ------------------------------------ atlas 计算函数 ------------------------------------

// AtlasCalculateNumberOfSourceMacs atlas 计算源节点计算的 MAC 的次数
func (mcu *MacsCalculationUnit) AtlasCalculateNumberOfSourceMacs(segments []*Segment) int {
	finalResult := 0
	for _, singleSegment := range segments {
		finalResult += (len(singleSegment.Path) - 1) * 2
	}
	return finalResult
}

// AtlasCalculateNumberOfOnPathRouterMacs 计算中间节点的 MAC 次数
func (mcu *MacsCalculationUnit) AtlasCalculateNumberOfOnPathRouterMacs(multipathNodeMapping map[string]*Node, source, destination string) float64 {
	// 如果构建的图的入度大于2的话就需要做2次 MAC 操作, 其他只需要进行一次 MAC 操作。
	macTotal := 0
	for _, node := range multipathNodeMapping {
		if node.NodeName != source && node.NodeName != destination {
			if node.Indegree >= 2 {
				macTotal += 3
			} else {
				macTotal += 2
			}
		}
	}
	averageMacs := float64(macTotal) / (float64(len(multipathNodeMapping)) - 2)

	return averageMacs
}

// AtlasCalculateNumberOfDestinationMacs 计算目的节点的 MAC 次数
func (mcu *MacsCalculationUnit) AtlasCalculateNumberOfDestinationMacs(firstSegment *Segment) int {
	return len(firstSegment.Path) + 1 - 1
}

// ------------------------------------ atlas 计算函数 ------------------------------------

// ------------------------------------ lip 计算函数 ------------------------------------

func (mcu *MacsCalculationUnit) LiPCalculateNumberOfSourceMacs(paths []*Path) int {
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

func (mcu *MacsCalculationUnit) LiPCalculateNumberOfDestinationMacs() int {
	return 1
}

func (mcu *MacsCalculationUnit) LiPCalculateNumberOfOnPathRouterMacs() int {
	return 1
}

// ------------------------------------ lip 计算函数 ------------------------------------
