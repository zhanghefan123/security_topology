package container_api

import (
	"context"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

func RemoveContainer(node *node.AbstractNode) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		sat, _ := node.ActualNode.(*satellite.NormalSatellite)
		RemoveNormalSatellite(sat)
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		sat, _ := node.ActualNode.(*satellite.ConsensusSatellite)
		RemoveConsensusSatellite(sat)
	} else {
		moduleContainerManagerLogger.Errorf("unsupported node type")
	}
}

func RemoveConsensusSatellite(sat *satellite.ConsensusSatellite) {
	// 1. 判断节点状态
	if sat.Status != types.NetworkNodeStatus_STOPPED {
		moduleContainerManagerLogger.Errorf("consensus satellite not in stopped state")
		return
	}

	// 2. 进行容器的停止
	containerId := sat.ContainerId
	err := client.DockerClient.ContainerStop(
		context.Background(),
		containerId,
		container.StopOptions{},
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to remove consensus satellite")
		return
	}

	// 3. 进行节点的状态转换
	sat.Status = types.NetworkNodeStatus_Logic
}

func RemoveNormalSatellite(sat *satellite.NormalSatellite) {
	// 1. 判断节点状态
	if sat.Status != types.NetworkNodeStatus_STOPPED {
		moduleContainerManagerLogger.Errorf("normal satellite not in stopped state")
		return
	}

	// 2. 进行容器的停止
	containerId := sat.ContainerId
	err := client.DockerClient.ContainerRemove(
		context.Background(),
		containerId,
		dockerTypes.ContainerRemoveOptions{},
	)
	if err != nil {
		moduleContainerManagerLogger.Errorf("fail to remove normal satellite")
		return
	}

	// 3. 进行节点的状态转换
	sat.Status = types.NetworkNodeStatus_Logic
}
