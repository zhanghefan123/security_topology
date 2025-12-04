package graph

type SourceAndDest struct {
	Source      *MultipathGraphNode
	Destination *MultipathGraphNode
}

func CreateSourceAndDest(source, destination *MultipathGraphNode) *SourceAndDest {
	return &SourceAndDest{
		Source:      source,
		Destination: destination,
	}
}
