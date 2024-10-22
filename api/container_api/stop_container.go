package container_api

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// StopContainer 进行容器的停止
func StopContainer(client *docker.Client, node *node.AbstractNode) error {
	// 1. 从抽象节点之中进行普通节点的获取
	var err error = nil
	normalNode, err := node.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get normal node from abstract node error: %w", err)
	}

	// 2. 判断节点状态
	if normalNode.Status != types.NetworkNodeStatus_Started {
		return fmt.Errorf("not in started state, cannot stop")
	}

	// 3. 获取容器id 进行容器的停止
	containerId := normalNode.ContainerId
	err = client.ContainerStop(
		context.Background(),
		containerId,
		container.StopOptions{})
	if err != nil {
		return fmt.Errorf("stop container error: %w", err)
	}

	// 4. 进行节点的状态转换
	normalNode.Status = types.NetworkNodeStatus_STOPPED

	return nil
}
