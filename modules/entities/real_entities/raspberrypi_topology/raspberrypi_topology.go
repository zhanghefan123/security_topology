package raspberrypi_topology

import (
	"github.com/c-robinson/iplib/v2"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	RaspberrypiTopologyInstance = NewRaspeberryTopology()
	raspberrypiTopologyLogger   = logger.GetLogger(logger.ModuleRaspberrypiConfigure)
)

type RaspberrypiTopology struct {
	Routers  []*nodes.Router
	LirNodes []*nodes.LiRNode

	RouterAbstractNodes []*node.AbstractNode
	LirAbstractNodes    []*node.AbstractNode

	TopologyGraph *simple.DirectedGraph

	AllAbstractNodes []*node.AbstractNode
	AbstractNodesMap map[string]*node.AbstractNode

	Links       []*link.AbstractLink
	AllLinksMap map[string]map[string]*link.AbstractLink

	Ipv4SubNets []iplib.Net4 // IPv4 两主机子网列表
	Ipv6SubNets []iplib.Net6 // IPv6 两主机子网列表

	topologyInitSteps  map[string]struct{} // 拓扑初始化步骤
	topologyStartSteps map[string]struct{} // 拓扑启动步骤

	NetworkInterfaces int // 网络接口数量
}

// NewRaspeberryTopology 创建新的拓扑
func NewRaspeberryTopology() *RaspberrypiTopology {
	topology := &RaspberrypiTopology{
		TopologyGraph: simple.NewDirectedGraph(),

		Routers:  make([]*nodes.Router, 0),
		LirNodes: make([]*nodes.LiRNode, 0),

		RouterAbstractNodes: make([]*node.AbstractNode, 0),
		LirAbstractNodes:    make([]*node.AbstractNode, 0),

		AllAbstractNodes: make([]*node.AbstractNode, 0),
		AbstractNodesMap: make(map[string]*node.AbstractNode),

		Links:       make([]*link.AbstractLink, 0),
		AllLinksMap: make(map[string]map[string]*link.AbstractLink),

		topologyInitSteps:  make(map[string]struct{}),
		topologyStartSteps: make(map[string]struct{}),

		NetworkInterfaces: 0,
	}
	raspberrypiTopologyLogger.Infof("create new images")
	return topology
}
