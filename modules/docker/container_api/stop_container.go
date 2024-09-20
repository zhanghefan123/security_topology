package container_api

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

func StopContainer(node *node.AbstractNode) {
	if node.Type == types.NetworkNodeType_NormalSatellite {
		normalSat, _ := node.ActualNode.(*satellite.NormalSatellite)
		StopNormalSatellite(normalSat)
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		consensusSat, _ := node.ActualNode.(*satellite.ConsensusSatellite)
		StopConsensusSatellite(consensusSat)
	} else {
		moduleContainerManagerLogger.Errorf("unsupported node type")
	}
}

func StopConsensusSatellite(sat *satellite.ConsensusSatellite) {
	// 1. 判断节点状态
	if sat.Status != types.NetworkNodeStatus_Started {
		moduleContainerManagerLogger.Errorf("consensus satellite not in started state")
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
		moduleContainerManagerLogger.Errorf("fail to stop consensus satellite")
		return
	}

	// 3. 进行节点的状态转换
	sat.Status = types.NetworkNodeStatus_STOPPED
}

func StopNormalSatellite(sat *satellite.NormalSatellite) {
	// 1. 判断节点状态
	if sat.Status != types.NetworkNodeStatus_Started {
		moduleContainerManagerLogger.Errorf("normal satellite not in started state")
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
		moduleContainerManagerLogger.Errorf("fail to stop the normal satellite")
		return
	}

	// 3. 进行节点的状态转换
	sat.Status = types.NetworkNodeStatus_STOPPED
}
