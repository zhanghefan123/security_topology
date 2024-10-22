package container_api

import (
	"context"
	"fmt"
	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// StartContainer 启动容器
func StartContainer(client *docker.Client, node *node.AbstractNode) error {
	// 1. 从抽象节点之中进行普通节点的获取
	var err error = nil
	normalNode, err := node.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get normal node from abstract node err: %w", err)
	}

	// 2. 判断节点状态
	if normalNode.Status != types.NetworkNodeStatus_Created {
		return fmt.Errorf("not in started state, cannot stop")
	}

	// 3. 获取容器 id 进行容器的停止
	containerId := normalNode.ContainerId
	err = client.ContainerStart(context.Background(),
		containerId,
		dockerTypes.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("start container err: %w", err)
	}

	// 4. 进行信息的获取
	info, err := client.ContainerInspect(
		context.Background(),
		normalNode.ContainerId,
	)
	if err != nil {
		return fmt.Errorf("fail to inspect the normal satellite")
	}
	normalNode.Pid = info.State.Pid

	// 5. 进行状态的转换
	normalNode.Status = types.NetworkNodeStatus_Started

	return nil
}
