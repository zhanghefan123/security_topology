package apis

import (
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/api/etcd_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/constellation"
)

// GetConstellationState 获取星座的状态
func GetConstellationState(c *gin.Context) {
	if constellation.ConstellationInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "down",
		})
	} else {
		constellationParams := map[string]int{
			"orbit_number":        constellation.ConstellationInstance.OrbitNumber,
			"satellite_per_orbit": constellation.ConstellationInstance.SatellitePerOrbit,
		}

		c.JSON(http.StatusOK, gin.H{
			"state":                "up",
			"constellation_params": constellationParams,
		})
	}
}

// GetInstancePositions 获取所有实例的位置
func GetInstancePositions(c *gin.Context) {
	if constellation.ConstellationInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": "constellation already stopped",
		})
		return
	}

	links := map[int][]string{}

	for _, link := range constellation.ConstellationInstance.AllSatelliteLinks {
		links[link.Id] = make([]string, 0)
		links[link.Id] = append(links[link.Id], link.SourceContainerName)
		links[link.Id] = append(links[link.Id], link.TargetContainerName)
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": constellation.ConstellationInstance.ContainerNameToPosition,
		"links":     links,
	})
}

// StartConstellation 进行星座的启动
func StartConstellation(c *gin.Context) {
	// 如果已经存在实例之后就不要再创建了
	if constellation.ConstellationInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"message": "constellation already created",
		})
		return
	}

	// 反序列化
	constellationParams := &constellation.Parameters{}
	err := c.ShouldBindJSON(constellationParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "down",
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}
	//
	// 处理逻辑 -> 应该只需要更新卫星数量和每个轨道的卫星数量即可
	err = startConstellationInner(constellationParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "down",
			"message": fmt.Sprintf("startConstellationInner err: %v", err),
		})
		return
	}

	// 进行结果的返回
	c.JSON(http.StatusOK, gin.H{
		"status":  "up",
		"message": "successfully start the constellation",
	})
}

// startConstellationInner 实际的启动星座的逻辑
func startConstellationInner(constellationParams *constellation.Parameters) error {
	var err error // 创建错误
	var dockerClient *docker.Client
	// 初始化本地配置
	err = configs.InitLocalConfig()
	if err != nil {
		return fmt.Errorf("init local config err: %v", err)
	}
	// 初始化 dockerClient
	dockerClient, err = client.NewDockerClient() // 创建新的 docker client
	if err != nil {
		return fmt.Errorf("create docker client err: %v", err)
	}
	// 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	startTime := configs.TopConfiguration.ConstellationConfig.GoStartTime
	// 替换掉启动的卫星数量, 以及每个轨道的卫星数量
	configs.TopConfiguration.ConstellationConfig.OrbitNumber = constellationParams.OrbitNumber
	configs.TopConfiguration.ConstellationConfig.SatellitePerOrbit = constellationParams.SatellitePerOrbit
	// 创建星座实例
	constellation.ConstellationInstance = constellation.NewConstellation(dockerClient, etcdClient, startTime) // 创建一个星座, 使用的参数是 dockerClient
	err = constellation.ConstellationInstance.Init()                                                          // 进行星座的初始化
	if err != nil {
		return fmt.Errorf("init constellation err: %w", err)
	}
	err = constellation.ConstellationInstance.Start()
	if err != nil {
		return fmt.Errorf("start constellation err: %w", err)
	}
	return nil
}

// StopConstellation 进行星座的停止
func StopConstellation(c *gin.Context) {
	if constellation.ConstellationInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": "constellation already stopped",
		})
		return
	}

	// 没有参数, 直接进入处理逻辑
	err := stopConstellationInner()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "down",
			"message": fmt.Sprintf("stopConstellationInner err: %v", err),
		})
		return
	}

	// 进行成功的结果的返回
	c.JSON(http.StatusOK, gin.H{
		"status":  "down",
		"message": "successfully stop the constellation",
	})
}

// stopConstellationInner 停止星座的内部逻辑
func stopConstellationInner() error {
	err := constellation.ConstellationInstance.Remove()
	defer func() {
		constellation.ConstellationInstance = nil
	}()
	if err != nil {
		return fmt.Errorf("stop constellation error: %w", err)
	}
	return nil
}
