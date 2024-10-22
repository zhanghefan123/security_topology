package params

type NodeParam struct {
	Index int     `json:"index"`
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type LinkParam struct {
	SourceNode NodeParam `json:"source_node"`
	TargetNode NodeParam `json:"target_node"`
}

type TopologyParams struct {
	Nodes []NodeParam `json:"nodes"`
	Links []LinkParam `json:"links"`
}
