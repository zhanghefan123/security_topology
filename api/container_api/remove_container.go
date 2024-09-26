package container_api

import (
	"context"
	"fmt"
	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// RemoveContainer 进行容器的删除
func RemoveContainer(client *docker.Client, node *node.AbstractNode) error {
	var err error = nil
	normalNode, err := node.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get normal node from abstract node error: %w", err)
	}

	// 1. 判断节点状态
	if normalNode.Status != types.NetworkNodeStatus_STOPPED {
		return fmt.Errorf("not in stopped state, cannot remove")
	}

	// 2. 获取容器id 进行容器的停止
	containerId := normalNode.ContainerId
	err = client.ContainerRemove(
		context.Background(),
		containerId,
		dockerTypes.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("remove container error: %w", err)
	}

	// 3. 进行节点的状态转换
	normalNode.Status = types.NetworkNodeStatus_Logic

	return nil
}
