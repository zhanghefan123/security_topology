package topo

import "zhanghefan123/security_topology/services/http_server/params/node_and_link"

// TopologyParams 拓扑相关参数
type TopologyParams struct {
	Nodes        []node_and_link.NodeParam `json:"nodes"`
	Links        []node_and_link.LinkParam `json:"links"`
	PerLinkDelay float64                   `json:"per_link_delay"`
}
