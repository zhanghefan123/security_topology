package apis

import (
	"context"
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"zhanghefan123/security_topology/api/etcd_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/constellation"
	"zhanghefan123/security_topology/modules/entities/types"
)

// GetConstellationState 获取星座的状态
func GetConstellationState(c *gin.Context) {
	if constellation.ConstellationInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "down",
		})
	} else {
		// 通过 groundStations 拿到所有的 groundStationContainerNames
		groundStationRealNames := make([]string, 0)
		for _, groundStation := range constellation.ConstellationInstance.GroundStations {
			groundStationRealNames = append(groundStationRealNames, groundStation.RealName)
		}

		constellationParams := map[string]interface{}{
			"orbit_number":            constellation.ConstellationInstance.OrbitNumber,
			"satellite_per_orbit":     constellation.ConstellationInstance.SatellitePerOrbit,
			"ground_stations":         groundStationRealNames,
			"minimum_elevation_angle": constellation.ConstellationInstance.MinimumElevationAngle,
			"time_step":               constellation.ConstellationInstance.TimeStep,
		}

		c.JSON(http.StatusOK, gin.H{
			"state":                "up",
			"constellation_params": constellationParams,
		})
	}
}

// GetInstancesPositions 获取所有实例的位置
func GetInstancesPositions(c *gin.Context) {
	// 1. 如果还没有创建星座实例 -> 那么就直接进行错误的返回
	if constellation.ConstellationInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "down",
			"message": "constellation already stopped",
		})
		return
	}

	// 进行星间链路的整合
	// -----------------------------------------------------------------------------
	isls := map[int][]string{}
	for _, isl := range constellation.ConstellationInstance.AllSatelliteLinks {
		if isl.Type == types.NetworkLinkType_IntraOrbitSatelliteLink {
			isls[isl.Id] = make([]string, 0)
			isls[isl.Id] = append(isls[isl.Id], isl.SourceContainerName)
			isls[isl.Id] = append(isls[isl.Id], isl.TargetContainerName)
			isls[isl.Id] = append(isls[isl.Id], "intra_orbit")
		} else if isl.Type == types.NetworkLinkType_InterOrbitSatelliteLink {
			isls[isl.Id] = make([]string, 0)
			isls[isl.Id] = append(isls[isl.Id], isl.SourceContainerName)
			isls[isl.Id] = append(isls[isl.Id], isl.TargetContainerName)
			isls[isl.Id] = append(isls[isl.Id], "inter_orbit")
		}
	}
	// -----------------------------------------------------------------------------

	// 进行星地链路的整合
	// -----------------------------------------------------------------------------
	gsls := map[int][]string{}
	for _, gsl := range constellation.ConstellationInstance.AllGroundSatelliteLinks {
		if gsl.Status {
			gsls[gsl.Id] = make([]string, 0)
			gsls[gsl.Id] = append(gsls[gsl.Id], gsl.SourceContainerName)
			gsls[gsl.Id] = append(gsls[gsl.Id], gsl.TargetContainerName)
		}
	}
	// -----------------------------------------------------------------------------

	c.JSON(http.StatusOK, gin.H{
		"positions": constellation.ConstellationInstance.ContainerNameToPosition,
		"isls":      isls,
		"gsls":      gsls,
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

	// 看是否能进行成功的绑定 -> 将反序列化的参数打印一下
	// -------------------------------------------------------------------------------
	fmt.Printf("orbitNumber: %d | satellitePerOrbit: %d\n",
		constellationParams.OrbitNumber,
		constellationParams.SatellitePerOrbit)
	for _, groundStationParam := range constellationParams.GroundStationsParams {
		fmt.Printf("ground station name: %s | longitude: %f | latitude: %f\n",
			groundStationParam.Name,
			groundStationParam.Longitude,
			groundStationParam.Latitude)
	}
	// -------------------------------------------------------------------------------

	// 处理逻辑 -> 应该只需要更新卫星数量和每个轨道的卫星数量即可
	err = startConstellationInner(constellationParams)
	if err != nil {
		// 不止进行星座启动
		fmt.Printf("start constellation inner error: %v", err)
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
	// 创建错误
	var err error
	// 创建 docker 客户端
	var dockerClient *docker.Client
	// 初始化本地配置
	err = configs.InitLocalConfig()
	// 看是否存在错误
	if err != nil {
		// 如果错误存在, 就进行返回
		return fmt.Errorf("init local config err: %v", err)
	}
	// 初始化 dockerClient
	dockerClient, err = client.NewDockerClient() // 创建新的 docker client
	if err != nil {
		// 如果存在错误就进行返回
		return fmt.Errorf("create docker client err: %v", err)
	}
	// 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	// 获取星座启动时间
	startTime := configs.TopConfiguration.ConstellationConfig.GoStartTime
	// 创建星座实例
	constellation.ConstellationInstance = constellation.NewConstellation(dockerClient, etcdClient, startTime, constellationParams)
	// 进行星座的初始化
	err = constellation.ConstellationInstance.Init()
	if err != nil {
		return fmt.Errorf("init constellation err: %w", err)
	}
	// 进行星座的启动
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

// ChangeTimeStamp 进行时间戳的改变
func ChangeTimeStamp(c *gin.Context) {
	timeStampParams := &constellation.TimeStampParameter{}
	err := c.ShouldBindJSON(timeStampParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("bindjson err: %v", err),
		})
		return
	}

	// 将结果存储到 etcd 之中
	timeStepKey := configs.TopConfiguration.ConstellationConfig.TimeStepKey
	_, err = constellation.ConstellationInstance.EtcdClient.Put(context.Background(),
		timeStepKey, strconv.Itoa(timeStampParams.TimeStep))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("store to etcd error: %v", err),
		})
		return
	}

	// 进行结果的返回
	c.JSON(http.StatusOK, gin.H{
		"message": "change time step successfully",
	})
}
