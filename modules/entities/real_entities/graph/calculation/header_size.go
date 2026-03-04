package calculation

import (
	"fmt"
	"math"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

const (
	DataHashSize  = 16
	SessionIdSize = 16
	TimestampSize = 4

	StartTagSize = 1
	EndTagSize   = 1
	PvfSize      = 16
	OpvSize      = 16

	HvfSize = 16
	DvfSize = 16
)

// ------------------------------------ opt 计算函数 ------------------------------------

func OptCalculateNumberOfOpvs(paths []*entities.Path) int {
	hasOpvs := map[string]struct{}{}
	opvCount := 0
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
	return opvCount
}

func OptCalculateHeaderSize(paths []*entities.Path) int {
	opvCount := OptCalculateNumberOfOpvs(paths)
	return DataHashSize + SessionIdSize + TimestampSize + opvCount*OpvSize + PvfSize
}

// ------------------------------------ opt 计算函数 ------------------------------------

// ------------------------------------ atlas 计算函数 ------------------------------------

// AtlasCalculateHeaderSize 基于 segments 进行报头的大小的计算
func AtlasCalculateHeaderSize(segments []*entities.Segment) int {
	metaDataSize := DataHashSize + SessionIdSize + TimestampSize
	validationFieldSize := 0
	for _, segment := range segments {
		validationFieldSize += AtlasCalculateHeaderSizeForSingleSegment(segment)
	}
	return metaDataSize + validationFieldSize
}

func AtlasCalculateHeaderSizeForSingleSegment(segment *entities.Segment) int {
	return StartTagSize + PvfSize + (len(segment.Path)-1)*OpvSize + EndTagSize
}

// ------------------------------------ atlas 计算函数 ------------------------------------

// ------------------------------------ lip 计算函数 ------------------------------------

func LiPCalculateHeaderSize(paths []*entities.Path, fpr float64) int {
	fixedPartSize := TimestampSize + DataHashSize + SessionIdSize + HvfSize + DvfSize
	hasHVFs := map[string]struct{}{}
	hvfCount := 0
	// 进行所有的路径的遍历
	longestPathLength := 0
	for _, path := range paths {
		currentPathString := ""
		for index, node := range path.NodeList { //  1 2 3
			if index < (len(path.NodeList) - 2) {
				nextIndex := index + 1
				nextNode := path.NodeList[nextIndex]
				currentPathString += fmt.Sprintf("%s->%s->", node.NodeName, nextNode.NodeName)
				if _, ok := hasHVFs[currentPathString]; !ok {
					hasHVFs[currentPathString] = struct{}{}
					hvfCount += 1
				}
			}
		}
		if len(path.NodeList) > longestPathLength {
			longestPathLength = len(path.NodeList)
		}
	}
	variableLengthLpf := float64(-hvfCount) * math.E * math.Log(fpr)
	return fixedPartSize + int(math.Ceil(variableLengthLpf/8)) + longestPathLength*4
}

// ------------------------------------ lip 计算函数 ------------------------------------
