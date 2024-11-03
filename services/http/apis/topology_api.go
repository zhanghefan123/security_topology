package apis

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
	"zhanghefan123/security_topology/services/http/params"
)

var (
	TopologyInstance *topology.Topology
)

// GetTopologyState 进行拓扑状态的获取
func GetTopologyState(c *gin.Context) {
	if TopologyInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "down",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"state":           "up",
			"topology_params": TopologyInstance.TopologyParams, // 如果已经创建完成了, 还需要进行创建的参数的返回
		})
	}
}

// StartTopology 进行拓扑的启动
func StartTopology(c *gin.Context) {
	if TopologyInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "topology already created",
		})
	}

	// 反序列化
	topologyParams := &params.TopologyParams{}
	err := c.BindJSON(topologyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}
	fmt.Println(topologyParams) // 打印拓扑

	// 核心处理逻辑
	err = startTopologyInner(topologyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("start topology err: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully start the topology",
	})
}

// startTopologyInner 实际的拓扑启动逻辑
func startTopologyInner(topologyParams *params.TopologyParams) error {
	var err error
	var dockerClient *docker.Client
	// 初始化本地配置
	err = configs.InitLocalConfig()
	if err != nil {
		return fmt.Errorf("init local config err: %w", err)
	}
	// 初始化 dockerClient
	dockerClient, err = client.NewDockerClient()
	if err != nil {
		return fmt.Errorf("create docker client err: %w", err)
	}
	// 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	// 创建拓扑实例
	TopologyInstance = topology.NewTopology(dockerClient, etcdClient, topologyParams)

	// 进行 init
	err = TopologyInstance.Init()
	if err != nil {
		return fmt.Errorf("init topology err: %w", err)
	}
	err = TopologyInstance.Start()
	if err != nil {
		return fmt.Errorf("start topology err: %w", err)
	}
	pretty.Println(*topologyParams)
	return nil
}

// StopTopology 进行拓扑的删除
func StopTopology(c *gin.Context) {
	if TopologyInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": "images already stopped",
		})
		return
	}

	err := stopTopologyInner()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": fmt.Sprintf("stop topology err: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "down",
		"message": "successfully stop the images",
	})
}

// stopTopologyInner 实际的拓扑销毁逻辑
func stopTopologyInner() error {
	err := TopologyInstance.Remove()
	defer func() {
		TopologyInstance = nil
	}()
	if err != nil {
		return fmt.Errorf("remove topology err: %v", err)
	}
	return nil
}
