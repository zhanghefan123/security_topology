package container_api

import (
	"context"
	dockerTypes "github.com/docker/docker/api/types"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

// StartContainer 启动容器
func StartContainer(node *node.AbstractNode) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		normalSat, _ := node.ActualNode.(*satellite.NormalSatellite)
		StartNormalSatellite(normalSat)
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		consensusSat, _ := node.ActualNode.(*satellite.ConsensusSatellite)
		StartConsensusSatellite(consensusSat)
	} else {
		moduleContainerManagerLogger.Errorf("unsupported node type")
	}
}

// StartNormalSatellite 进行普通卫星的启动
func StartNormalSatellite(sat *satellite.NormalSatellite) {
	// 1. 进行容器的启动
	containerId := sat.ContainerId
	err := client.DockerClient.ContainerStart(
		context.Background(),
		containerId,
		dockerTypes.ContainerStartOptions{},
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to start the normal satellite")
		return
	}

	// 2. 进行信息 (i.e., pid) 的获取
	info, err := client.DockerClient.ContainerInspect(
		context.Background(),
		sat.ContainerId,
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to inspect the normal satellite")
		return
	}
	sat.Pid = info.State.Pid
}

// StartConsensusSatellite 进行共识卫星的启动
func StartConsensusSatellite(sat *satellite.ConsensusSatellite) {
	// 1. 进行容器的启动
	containerId := sat.ContainerId
	err := client.DockerClient.ContainerStart(
		context.Background(),
		containerId,
		dockerTypes.ContainerStartOptions{},
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to start the consensus satellite")
		return
	}

	// 2. 进行信息 (i.e., pid) 的获取
	info, err := client.DockerClient.ContainerInspect(
		context.Background(),
		sat.ContainerId,
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to inspect the normal satellite")
		return
	}
	sat.Pid = info.State.Pid
}
