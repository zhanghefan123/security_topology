package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/types"
)

func CreateEntrance(client *docker.Client, entrance *nodes.Entrance) error {
	// 1. 检查状态
	if entrance.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("entrance not in logic status cannot create")
	}

	// 2. 创建 sysctls -> 注意 entrance 是 host 网路模式
	var sysctls map[string]string

	// 3. 获取配置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, entrance.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr

	// 4. 创建容器卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", entrance.ContainerName)),
	}

	// 5. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", entrance.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", entrance.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(entrance.Type), entrance.Id, 1)),
	}

	// 6. containerConfig
	containerConfig := &container.Config{
		Image: configs.TopConfiguration.ImagesConfig.EntranceImageName,
		Tty:   true,
		Env:   envs,
	}

	// 7. hostConfig
	var hostConfig *container.HostConfig
	hostConfig = &container.HostConfig{
		// 容器数据卷映射
		Binds:       volumes,
		CapAdd:      []string{"NET_ADMIN"},
		Privileged:  true,
		Sysctls:     sysctls,
		NetworkMode: "host",
	}

	// 8. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		entrance.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create entrance failed %v", err)
	}

	entrance.ContainerId = response.ID

	// 9. 状态转换
	entrance.Status = types.NetworkNodeStatus_Created

	return nil
}
