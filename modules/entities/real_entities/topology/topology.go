package topology

import (
	"github.com/c-robinson/iplib/v2"
	"github.com/coreos/etcd/clientv3"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/services/http/params"
)

var (
	topologyLogger = logger.GetLogger(logger.ModuleTopology)
)

type Topology struct {
	client         *docker.Client
	etcdClient     *clientv3.Client
	TopologyParams *params.TopologyParams
	Ipv4SubNets    []iplib.Net4
	Ipv6SubNets    []iplib.Net6

	Routers        []*nodes.Router
	NormalNodes    []*normal_node.NormalNode
	ConsensusNodes []*nodes.ConsensusNode
	MaliciousNodes []*nodes.MaliciousNode

	RouterAbstractNodes    []*node.AbstractNode
	NormalAbstractNodes    []*node.AbstractNode
	ConsensusAbstractNodes []*node.AbstractNode
	MaliciousAbstractNodes []*node.AbstractNode
	AllAbstractNodes       []*node.AbstractNode

	Links    []*link.AbstractLink
	LinksMap map[string]map[string]*link.AbstractLink // map[sourceContainerName][targetContainerName]*link.AbstractLink

	topologyInitSteps  map[string]struct{} // 拓扑初始化步骤
	topologyStartSteps map[string]struct{} // 拓扑启动步骤
	topologyStopSteps  map[string]struct{} // 拓扑停止步骤
}

// NewTopology 创建新的拓扑
func NewTopology(client *docker.Client, etcdClient *clientv3.Client, params *params.TopologyParams) *Topology {
	topology := &Topology{
		client:             client,
		etcdClient:         etcdClient,
		TopologyParams:     params,
		Routers:            make([]*nodes.Router, 0),
		NormalNodes:        make([]*normal_node.NormalNode, 0),
		ConsensusNodes:     make([]*nodes.ConsensusNode, 0),
		MaliciousNodes:     make([]*nodes.MaliciousNode, 0),
		Links:              make([]*link.AbstractLink, 0),
		LinksMap:           make(map[string]map[string]*link.AbstractLink),
		topologyInitSteps:  make(map[string]struct{}),
		topologyStartSteps: make(map[string]struct{}),
		topologyStopSteps:  make(map[string]struct{}),
	}
	topologyLogger.Infof("create new images")
	return topology
}
