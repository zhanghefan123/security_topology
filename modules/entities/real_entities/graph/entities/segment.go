package entities

import (
	"fmt"
	"strconv"
)

type Segment struct {
	Id                    int
	FinalDestinationIndex int     // 这个段所属的最终目的节点
	Source                *Node   // 段起点
	Destination           *Node   // 段终点
	Depth                 int     // 段的深度
	Path                  []*Node // 部分的路径
	PathStr               string
	ParentId              int
}

func CreateSegment(source *Node, destination *Node, depth int,
	path []*Node) *Segment {
	return &Segment{
		Source:      source,
		Destination: destination,
		Depth:       depth,
		Path:        path,
	}
}

func SegmentToString(segment *Segment) string {
	finalString := "id:" + strconv.Itoa(segment.Id) + " depth: " + strconv.Itoa(segment.Depth) + " parent segment id: " + strconv.Itoa(segment.ParentId) + " path:"
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			finalString = finalString + node.NodeName + "->"
		} else {
			finalString = finalString + node.NodeName
		}
	}
	return finalString
}

func SegmentToIndexString(segment *Segment) string {
	finalString := fmt.Sprintf("%d,%d,%d,%d,%d,", segment.Id, segment.FinalDestinationIndex, segment.Depth, segment.ParentId, len(segment.Path))
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			finalString += fmt.Sprintf("%d,", node.Index)
		} else {
			finalString += fmt.Sprintf("%d", node.Index)
		}
	}
	return finalString
}

func PrintSegment(segment *Segment) {
	fmt.Println(SegmentToString(segment))
}

func PrintSegments(segments []*Segment) {
	for _, segment := range segments {
		PrintSegment(segment)
	}
}
