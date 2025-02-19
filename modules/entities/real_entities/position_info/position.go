package position_info

type Position struct {
	NodeType  string  `json:"node_type"` // 节点类型
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
	Altitude  float64 `json:"altitude"`  // 高度
}
