package graph

import (
	"fmt"
	"math"
)

const (
	DATA_HASH_SIZE  = 16
	SESSION_ID_SIZE = 16
	TIMESTAMP_SIZE  = 4

	START_TAG_SIZE = 1
	END_TAG_SIZE   = 1
	PVF_SIZE       = 16
	OPV_SIZE       = 16

	HVF_SIZE = 16
	DVF_SIZE = 16
)

type HeaderCalculationUnit struct {
}

// ------------------------------------ opt 计算函数 ------------------------------------

func (hcu *HeaderCalculationUnit) OptCalculateNumberOfOpvs(paths []*Path) int {
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

func (hcu *HeaderCalculationUnit) OptCalculateHeaderSize(paths []*Path) float64 {
	opvCount := hcu.OptCalculateNumberOfOpvs(paths)
	return float64(DATA_HASH_SIZE + SESSION_ID_SIZE + TIMESTAMP_SIZE + opvCount*OPV_SIZE + PVF_SIZE)
}

// ------------------------------------ opt 计算函数 ------------------------------------

// ------------------------------------ atlas 计算函数 ------------------------------------

// AtlasCalculateHeaderSize 基于 segments 进行报头的大小的计算
func (hcu *HeaderCalculationUnit) AtlasCalculateHeaderSize(segments []*Segment) int {
	metaDataSize := DATA_HASH_SIZE + SESSION_ID_SIZE + TIMESTAMP_SIZE
	validationFieldSize := 0
	for _, segment := range segments {
		validationFieldSize += hcu.AtlasCalculateHeaderSizeForSingleSegment(segment)
	}
	return metaDataSize + validationFieldSize
}

func (hcu *HeaderCalculationUnit) AtlasCalculateNumberOfOpvs(segments []*Segment) int {
	finalResult := 0
	for _, segment := range segments {
		finalResult += len(segment.Path) - 1
	}
	return finalResult
}

func (hcu *HeaderCalculationUnit) AtlasCalculateHeaderSizeForSingleSegment(segment *Segment) int {
	return START_TAG_SIZE + PVF_SIZE + (len(segment.Path)-1)*OPV_SIZE + END_TAG_SIZE
}

// ------------------------------------ atlas 计算函数 ------------------------------------

// ------------------------------------ lip 计算函数 ------------------------------------

func (hcu *HeaderCalculationUnit) LiPCalculateHeaderSize(paths []*Path, fpr float64) int {
	fixedPartSize := TIMESTAMP_SIZE + DATA_HASH_SIZE + SESSION_ID_SIZE + HVF_SIZE + DVF_SIZE
	hasHVFs := map[string]struct{}{}
	hvfCount := 0
	// 进行所有的路径的遍历
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
	}
	variableLengthLpf := float64(-hvfCount) * math.E * math.Log(fpr)
	return fixedPartSize + int(math.Ceil(variableLengthLpf/8))
}

// ------------------------------------ lip 计算函数 ------------------------------------
