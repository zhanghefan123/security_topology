package http

import (
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"net/http"
	"zhanghefan123/security_topology/api/etcd_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
)

var (
	topologyInstance *topology.Topology
)

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

// StartTopology 进行拓扑的启动
func StartTopology(c *gin.Context) {
	if topologyInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "",
			"message": "topology already created",
		})
	}

	// 反序列化
	topologyParams := &TopologyParams{}
	err := c.BindJSON(topologyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "down",
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}

	// 核心处理逻辑
	startTopologyInner(topologyParams)

	c.JSON(http.StatusOK, gin.H{
		"status":  "up",
		"message": "successfully start the constellation",
	})
}

// startTopologyInner 实际的拓扑启动逻辑
func startTopologyInner(topologyParams *TopologyParams) {
	var err error
	var dockerClient *docker.Client
	// 初始化本地配置
	err = configs.InitLocalConfig()
	if err != nil {
		HttpServiceLogger.Errorf("init local configuration failed: %v", err)
	}
	// 初始化 dockerClient
	dockerClient, err = client.NewDockerClient()
	if err != nil {
		HttpServiceLogger.Errorf("create docker client failed: %v", err)
	}
	// 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	// 创建拓扑实例
	topologyInstance = topology.NewTopology(dockerClient, etcdClient, topologyParams)

	pretty.Println(*topologyParams)
}

// StopTopology 进行拓扑的删除
func StopTopology(c *gin.Context) {
	if topologyInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": "topology already stopped",
		})
	}

	stopTopologyInner()

	c.JSON(http.StatusOK, gin.H{
		"status":  "down",
		"message": "successfully stop the topology",
	})
}

// stopTopologyInner 实际的拓扑销毁逻辑
func stopTopologyInner() {
	topologyInstance = nil
}
