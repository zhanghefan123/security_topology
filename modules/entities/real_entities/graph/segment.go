package graph

import "strconv"

type Segment struct {
	Source      *Node   // 段起点
	Destination *Node   // 段终点
	Depth       int     // 段的深度
	Path        []*Node // 部分的路径
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

func PrintSegment(segment *Segment) {
	finalString := "depth: " + strconv.Itoa(segment.Depth) + " path:"
	for index, node := range segment.Path {
		if index != (len(segment.Path) - 1) {
			finalString = finalString + node.NodeName + "->"
		} else {
			finalString = finalString + node.NodeName
		}
	}
}

func PrintSegments(segments []*Segment) {
	for _, segment := range segments {
		PrintSegment(segment)
	}
}
