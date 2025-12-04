package graph

type Segment struct {
	Source      *MultipathGraphNode   // 段起点
	Destination *MultipathGraphNode   // 段终点
	Depth       int                   // 段的深度
	Path        []*MultipathGraphNode // 部分的路径
}

func CreateSegment(source *MultipathGraphNode, destination *MultipathGraphNode, depth int,
	path []*MultipathGraphNode) *Segment {
	return &Segment{
		Source:      source,
		Destination: destination,
		Depth:       depth,
		Path:        path,
	}
}
