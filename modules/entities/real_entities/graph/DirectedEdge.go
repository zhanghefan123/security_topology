package graph

type DirectedEdge struct {
	From *MultipathGraphNode
	To   *MultipathGraphNode
}

func CreateDirectedEdge(from, to *MultipathGraphNode) *DirectedEdge {
	return &DirectedEdge{
		From: from,
		To:   to,
	}
}
