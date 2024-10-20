package params

type NodeParam struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

type LinkParam struct {
	SourceNode NodeParam `json:"source_node"`
	TargetNode NodeParam `json:"target_node"`
}

type TopologyParams struct {
	Nodes []NodeParam `json:"nodes"`
	Links []LinkParam `json:"links"`
}
