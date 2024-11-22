package topology

import (
	"github.com/c-robinson/iplib/v2"
	docker "github.com/docker/docker/client"
	"go.etcd.io/etcd/client/v3"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/services/http/params"
)

var (
	TopologyInstance *Topology
	topologyLogger   = logger.GetLogger(logger.ModuleTopology)
)

type Topology struct {
	client         *docker.Client
	etcdClient     *clientv3.Client
	TopologyParams *params.TopologyParams
	Ipv4SubNets    []iplib.Net4
	Ipv6SubNets    []iplib.Net6
	TopologyGraph  *simple.DirectedGraph

	Routers         []*nodes.Router
	NormalNodes     []*normal_node.NormalNode
	ConsensusNodes  []*nodes.ConsensusNode
	ChainmakerNodes []*nodes.ChainmakerNode
	MaliciousNodes  []*nodes.MaliciousNode
	LirNodes        []*nodes.LiRNode

	RouterAbstractNodes     []*node.AbstractNode
	NormalAbstractNodes     []*node.AbstractNode
	ConsensusAbstractNodes  []*node.AbstractNode
	ChainMakerAbstractNodes []*node.AbstractNode
	MaliciousAbstractNodes  []*node.AbstractNode
	LirAbstractNodes        []*node.AbstractNode
	AllAbstractNodes        []*node.AbstractNode
	AbstractNodesMap        map[string]*node.AbstractNode

	Links       []*link.AbstractLink
	AllLinksMap map[string]map[string]*link.AbstractLink // map[sourceContainerName][targetContainerName]*link.AbstractLink

	topologyInitSteps  map[string]struct{} // 拓扑初始化步骤
	topologyStartSteps map[string]struct{} // 拓扑启动步骤
	topologyStopSteps  map[string]struct{} // 拓扑停止步骤
}

// NewTopology 创建新的拓扑
func NewTopology(client *docker.Client, etcdClient *clientv3.Client, params *params.TopologyParams) *Topology {
	topology := &Topology{
		client:         client,
		etcdClient:     etcdClient,
		TopologyParams: params,
		TopologyGraph:  simple.NewDirectedGraph(),

		Routers:         make([]*nodes.Router, 0),
		NormalNodes:     make([]*normal_node.NormalNode, 0),
		ConsensusNodes:  make([]*nodes.ConsensusNode, 0),
		ChainmakerNodes: make([]*nodes.ChainmakerNode, 0),
		MaliciousNodes:  make([]*nodes.MaliciousNode, 0),
		LirNodes:        make([]*nodes.LiRNode, 0),

		RouterAbstractNodes:     make([]*node.AbstractNode, 0),
		NormalAbstractNodes:     make([]*node.AbstractNode, 0),
		ConsensusAbstractNodes:  make([]*node.AbstractNode, 0),
		ChainMakerAbstractNodes: make([]*node.AbstractNode, 0),
		MaliciousAbstractNodes:  make([]*node.AbstractNode, 0),
		LirAbstractNodes:        make([]*node.AbstractNode, 0),

		AllAbstractNodes: make([]*node.AbstractNode, 0),
		AbstractNodesMap: make(map[string]*node.AbstractNode),

		Links:              make([]*link.AbstractLink, 0),
		AllLinksMap:        make(map[string]map[string]*link.AbstractLink),
		topologyInitSteps:  make(map[string]struct{}),
		topologyStartSteps: make(map[string]struct{}),
		topologyStopSteps:  make(map[string]struct{}),
	}
	topologyLogger.Infof("create new images")
	return topology
}
