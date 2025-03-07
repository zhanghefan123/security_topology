package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateRealTimePositionService 创建位置更新/延迟更新服务
func CreateRealTimePositionService(client *docker.Client, positionService *position.PositionService) error {
	// 1. 检查状态
	if positionService.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("position service status is %s", positionService.Status)
	}

	ContainerName := "realtime_position"

	// 2. 创建环境变量
	envs := []string{
		fmt.Sprintf("%s=%s", "ETCD_LISTEN_ADDR", positionService.EtcdListenAddr),
		fmt.Sprintf("%s=%d", "ETCD_CLIENT_PORT", positionService.EtcdClientPort),
		fmt.Sprintf("%s=%s", "ETCD_ISLS_PREFIX", positionService.EtcdISLsPrefix),
		fmt.Sprintf("%s=%s", "ETCD_GSLS_PREFIX", positionService.EtcdGSLsPrefix),
		fmt.Sprintf("%s=%s", "ETCD_SATELLITES_PREFIX", positionService.EtcdSatellitesPrefix),
		fmt.Sprintf("%s=%s", "ETCD_GROUND_STATIONS_PREFIX", positionService.EtcdGroundStationsPrefix),
		fmt.Sprintf("%s=%s", "CONSTELLATION_START_TIME", positionService.ConstellationStartTime),
		fmt.Sprintf("%s=%d", "UPDATE_INTERVAL", positionService.UpdateInterval),
		fmt.Sprintf("%s=%d", "SATELLITE_AVAILABLE_GSLS", configs.TopConfiguration.ConstellationConfig.SatelliteAvailableGSLs),
		fmt.Sprintf("%s=%s", "TIME_STEP_KEY", configs.TopConfiguration.ConstellationConfig.TimeStepKey),
		fmt.Sprintf("%s=%s", "MINIMUM_ELEVATION_ANGLE_KEY", configs.TopConfiguration.ConstellationConfig.MinimumElevationAngleKey),
		fmt.Sprintf("%s=%t", "UPDATE_POSITION", configs.TopConfiguration.ConstellationConfig.UpdatePosition),
	}

	// 3. 创建 containerConfig
	containerConfig := &container.Config{
		Image: configs.TopConfiguration.ImagesConfig.PositionServiceImageName,
		Env:   envs,
		Tty:   true,
	}

	// 4. 创建 hostConfig
	hostConfig := &container.HostConfig{
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: true,
	}

	// 5. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		ContainerName)
	if err != nil {
		return fmt.Errorf("create container err: %v", err)
	}

	// 6. 从 response 之中获取 ID
	positionService.ContainerId = response.ID

	// 7. 进行状态的转换
	positionService.Status = types.NetworkNodeStatus_Created
	return nil
}
