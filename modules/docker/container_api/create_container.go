package container_api

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"zhanghefan123/security_topology/modules/config/system"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleContainerManagerLogger = logger.GetLogger(logger.ModuleContainerManager)

// CreateContainer 创建 container
func CreateContainer(node *node.AbstractNode) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		sat, _ := node.ActualNode.(*satellite.NormalSatellite)
		CreateNormalSatellite(sat)
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		sat, _ := node.ActualNode.(*satellite.ConsensusSatellite)
		CreateConsensusSatellite(sat)
	} else {
		moduleContainerManagerLogger.Errorf("unsupported node type")
	}
}

// CreateNormalSatellite 创建普通卫星容器
func CreateNormalSatellite(satellite *satellite.NormalSatellite) {
	// 1. 检查状态
	if satellite.Status != types.NetworkNodeStatus_Logic {
		moduleContainerManagerLogger.Errorf("consensus satellite not in logic status")
		return
	}

	// 2. 创建容器
	// 容器数据卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", system.TopConfiguration.PathConfig.FrrPath.FrrHostPath,
			system.TopConfiguration.PathConfig.FrrPath.FrrContainerPath),
	}

	// 环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", satellite.Id),
	}

	// containerConfig
	containerConfig := &container.Config{
		Image: satellite.ImageName,
		Tty:   true,
		Env:   envs,
	}

	// hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds: volumes,
	}

	// 进行容器的创建
	response, err := client.DockerClient.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		satellite.ContainerName,
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("create satellite container failed")
		return
	}

	satellite.ContainerId = response.ID

	// 3. 状态转换
	satellite.Status = types.NetworkNodeStatus_Created
}

func CreateConsensusSatellite(satellite *satellite.ConsensusSatellite) {
	// 1. 检查状态
	if satellite.Status != types.NetworkNodeStatus_Logic {
		moduleContainerManagerLogger.Errorf("consensus satellite not in logic status")
		return
	}

	// 2. 创建容器
	// 容器数据卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", system.TopConfiguration.PathConfig.FrrPath.FrrHostPath,
			system.TopConfiguration.PathConfig.FrrPath.FrrContainerPath),
	}

	// 环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", satellite.Id),
	}

	// 暴露的端口
	rpcPort := nat.Port(fmt.Sprintf("%d/tcp", satellite.StartRpcPort+satellite.Id))
	p2pPort := nat.Port(fmt.Sprintf("%d/tcp", satellite.StartP2PPort+satellite.Id))

	// containerConfig
	containerConfig := &container.Config{
		Image: satellite.ImageName,
		// 容器暴露的端口
		ExposedPorts: nat.PortSet{
			// rpc 端口
			rpcPort: {},
			p2pPort: {},
		},
		Tty: true,
		Env: envs,
	}

	// hostConfig
	hostConfig := &container.HostConfig{
		// 端口映射
		PortBindings: nat.PortMap{
			rpcPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: string(rpcPort),
				},
			},
			p2pPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: string(p2pPort),
				},
			},
		},
		// 容器数据卷映射
		Binds: volumes,
	}

	// 进行容器的创建
	response, err := client.DockerClient.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		satellite.ContainerName,
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("create satellite container failed")
		return
	}

	satellite.ContainerId = response.ID

	// 3. 状态转换
	satellite.Status = types.NetworkNodeStatus_Created
}
