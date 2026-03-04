package entities

type SimNodeBase struct {
	NodeName string
	Index    int
}

func CreateSimNodeBase(NodeName string, NodeIndex int) *SimNodeBase {
	return &SimNodeBase{
		NodeName: NodeName,
		Index:    NodeIndex,
	}
}
