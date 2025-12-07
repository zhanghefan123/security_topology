package graph

type SourceAndDest struct {
	Source      *Node
	Destination *Node
}

func CreateSourceAndDest(source, destination *Node) *SourceAndDest {
	return &SourceAndDest{
		Source:      source,
		Destination: destination,
	}
}
