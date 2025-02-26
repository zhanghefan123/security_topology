package apis

import (
	"context"
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/api/etcd_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/modules/performance_monitor"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/file"
	"zhanghefan123/security_topology/services/http/params"
)

// GetAllTopologyNames 获取所有拓扑的名称
func GetAllTopologyNames() ([]string, error) {
	// 拿到所有的可能的文件列表
	allTopologyNamesWithSuffix, err := dir.ListFileNames(configs.TopConfiguration.ResourcesConfig.TopologiesDir)
	if err != nil {
		return nil, fmt.Errorf("get all topology names err: %v", err)
	}
	fmt.Println(allTopologyNamesWithSuffix)

	// 将拿到的文件名全部去掉后缀
	allTopologyNames := make([]string, len(allTopologyNamesWithSuffix))
	for index, name := range allTopologyNamesWithSuffix {
		allTopologyNames[index] = strings.TrimSuffix(name, filepath.Ext(name))
	}

	return allTopologyNames, nil
}

// ChangeStartDefence 改变是否进行攻击
func ChangeStartDefence(c *gin.Context) {
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "topology not started",
		})
		return
	}

	// 准备进行参数的绑定
	startDefenceParams := &params.StartDefenceParameter{}
	err := c.ShouldBindJSON(startDefenceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}

	// 进行参数的设置
	topology.TopologyInstance.TopologyParams.StartDefence = startDefenceParams.StartDefence

	// 设置到 etcd 之中
	startDefenceKey := configs.TopConfiguration.ChainMakerConfig.StartDefenceKey
	if startDefenceParams.StartDefence {
		_, err = topology.TopologyInstance.EtcdClient.Put(context.Background(), startDefenceKey, "true")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("etcd put err: %v", err),
			})
		}
	} else {
		_, err = topology.TopologyInstance.EtcdClient.Put(context.Background(), startDefenceKey, "false")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("etcd put err: %v", err),
			})
		}
	}

	// 进行结果的返回
	c.JSON(http.StatusOK, gin.H{
		"message": "start_defence setting success",
	})
}

// GetTopologyDescription 获取拓扑描述
func GetTopologyDescription(c *gin.Context) {
	topologyParams := &topology.TopologyParameters{}
	err := c.ShouldBindJSON(topologyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": fmt.Sprintf("bind parameters error %v", err),
		})
		return
	}

	// 进行文件的读取
	filePath := filepath.Join(configs.TopConfiguration.ResourcesConfig.TopologiesDir, fmt.Sprintf("%s.txt", topologyParams.TopologyName))
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("read file error %v", err),
		})
	}

	// 进行内容的返回
	c.JSON(http.StatusOK, gin.H{
		"topology_description": string(content),
	})
}

// SaveTopology 进行拓扑的存储
func SaveTopology(c *gin.Context) {
	topologyParams := &topology.TopologyParameters{}
	err := c.ShouldBindJSON(topologyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": fmt.Sprintf("bind parameters error %v", err),
		})
		return
	}

	// 进行文件的写入, 写入的结果加上 txt
	completeFilePath := filepath.Join(configs.TopConfiguration.ResourcesConfig.TopologiesDir, fmt.Sprintf("%s.txt", topologyParams.TopologyName))
	err = file.WriteStringIntoFile(completeFilePath, topologyParams.TopologyDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("write into file error %v", err),
		})
		return
	}

	// 获取所有的 names
	allTopologyNames, err := GetAllTopologyNames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("get all topology names err: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             "up",
		"all_topology_names": allTopologyNames,
	})
}

// GetTopologyState 进行拓扑状态的获取
func GetTopologyState(c *gin.Context) {
	// 进行所有的 topology 的获取
	AllTopologyNames, err := GetAllTopologyNames()
	fmt.Println(AllTopologyNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": fmt.Sprintf("get all topology names err: %v", err),
		})
		return
	}
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"state":              "down",
			"all_topology_names": AllTopologyNames,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"state":              "up",
			"topology_params":    topology.TopologyInstance.TopologyParams, // 如果已经创建完成了, 还需要进行创建的参数的返回
			"all_topology_names": AllTopologyNames,
		})
	}
}

// StartTopology 进行拓扑的启动
func StartTopology(c *gin.Context) {
	// 1. 如果拓扑还没有启动, 那么直接返回
	if topology.TopologyInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "topology already created",
		})
		return
	}

	// 2. 进行拓扑参数的绑定
	topologyParams := &params.TopologyParams{}
	err := c.BindJSON(topologyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}
	fmt.Println(topologyParams) // 打印拓扑

	// 3. 核心处理逻辑
	err = startTopologyInner(topologyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("start topology err: %v", err),
		})
		return
	}

	// 4. 当一切正常的时候进行结果的返回
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully start the topology",
	})
}

// startTopologyInner 实际的拓扑启动逻辑
func startTopologyInner(topologyParams *params.TopologyParams) error {
	var err error
	var dockerClient *docker.Client
	// 1. 初始化本地配置
	err = configs.InitLocalConfig()
	if err != nil {
		return fmt.Errorf("init local config err: %w", err)
	}
	// 2. 进行资源限制的加载
	configs.TopConfiguration.ResourcesConfig.CpuLimit = topologyParams.ConsensusNodeCpu
	configs.TopConfiguration.ResourcesConfig.MemoryLimit = topologyParams.ConsensusNodeMemory
	fmt.Println("consensus node memory", configs.TopConfiguration.ResourcesConfig.MemoryLimit)
	// 3. 初始化 dockerClient
	dockerClient, err = client.NewDockerClient()
	if err != nil {
		return fmt.Errorf("create docker client err: %w", err)
	}
	// 4. 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	// 5. 创建拓扑实例
	topology.TopologyInstance = topology.NewTopology(dockerClient, etcdClient, topologyParams)
	// 6. 进行 init
	err = topology.TopologyInstance.Init()
	if err != nil {
		return fmt.Errorf("init topology err: %w", err)
	}
	// 7. 进行start
	err = topology.TopologyInstance.Start()
	if err != nil {
		return fmt.Errorf("start topology err: %w", err)
	}
	// 8. 打印拓扑参数信息
	_, err = pretty.Println(*topologyParams)
	if err != nil {
		return fmt.Errorf("pretty.Println(topologyParams) err: %w", err)
	}
	return nil
}

// StopTopology 进行拓扑的删除
func StopTopology(c *gin.Context) {
	if topology.TopologyInstance == nil {
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
	err := topology.TopologyInstance.Remove()
	defer func() {
		topology.TopologyInstance = nil
		performance_monitor.PerformanceMonitorMapping = make(map[string]*performance_monitor.PerformanceMonitor)
	}()
	if err != nil {
		return fmt.Errorf("remove topology err: %v", err)
	}
	return nil
}
