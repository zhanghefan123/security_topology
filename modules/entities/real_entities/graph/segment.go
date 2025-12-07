package graph

import (
	"fmt"
	"strconv"
)

type Segment struct {
	Id          int
	Source      *Node   // 段起点
	Destination *Node   // 段终点
	Depth       int     // 段的深度
	Path        []*Node // 部分的路径
	PathStr     string
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
	finalString := "id:" + strconv.Itoa(segment.Id) + " depth: " + strconv.Itoa(segment.Depth) + " path:"
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			finalString = finalString + node.NodeName + "->"
		} else {
			finalString = finalString + node.NodeName
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
