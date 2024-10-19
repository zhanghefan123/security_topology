package topology

import (
	"github.com/c-robinson/iplib/v2"
	"github.com/coreos/etcd/clientv3"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/services/http"
)

var (
	topologyLogger = logger.GetLogger(logger.ModuleTopology)
)

type Topology struct {
	client         *docker.Client
	etcdClient     *clientv3.Client
	topologyParams *http.TopologyParams
	Nodes          []*node.AbstractNode
	Links          []*link.AbstractLink
	Ipv4SubNets    []iplib.Net4
	Ipv6SubNets    []iplib.Net6

	topologyInitSteps  map[string]struct{} // 拓扑初始化步骤
	topologyStartSteps map[string]struct{} // 拓扑启动步骤
}

// NewTopology 创建新的拓扑
func NewTopology(client *docker.Client, etcdClient *clientv3.Client, params *http.TopologyParams) *Topology {
	topology := &Topology{
		client:         client,
		etcdClient:     etcdClient,
		topologyParams: params,
		Nodes:          make([]*node.AbstractNode, 0),
		Links:          make([]*link.AbstractLink, 0),

		topologyInitSteps:  make(map[string]struct{}),
		topologyStartSteps: make(map[string]struct{}),
	}
	topologyLogger.Infof("create new topology")
	return topology
}
