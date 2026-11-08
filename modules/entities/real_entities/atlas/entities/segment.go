package entities

import (
	"fmt"
	"strconv"
	"strings"
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

// SegmentToString 这个是为了进行唯一标识的
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

// SegmentToIndexString 这个是为了方便进行读取的
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

func (s *Segment) BeautifulPrint() string {
	path := make([]string, 0, len(s.Path))
	for _, node := range s.Path {
		path = append(path, strconv.Itoa(node.Index))
	}
	pathStr := strings.Join(path, " → ")
	if len(path) == 0 {
		pathStr = "(empty)"
	}
	const width = 54
	border := strings.Repeat("─", width)
	return fmt.Sprintf(
		"┌%s┐\n"+
			"│  SEGMENT %-44d │\n"+
			"├%s┤\n"+
			"  ID           : %d\n"+
			"  Destination  : %d\n"+
			"  Depth        : %d\n"+
			"  Parent       : %d\n"+
			"  Path Length  : %d nodes\n"+
			"\n"+
			"  PATH\n"+
			"  %s\n"+
			"└%s┘",
		border,
		s.Id,
		border,
		s.Id,
		s.FinalDestinationIndex,
		s.Depth,
		s.ParentId,
		len(s.Path),
		pathStr,
		border,
	)
}

func (s *Segment) String() string {
	return s.BeautifulPrint()
}
