package create_container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/server/router"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateContainer 创建 container
func CreateContainer(client *docker.Client, node *node.AbstractNode) error {
	var err error = nil
	switch node.Type {
	case types.NetworkNodeType_NormalSatellite:
		sat, _ := node.ActualNode.(*satellites.NormalSatellite)
		err = CreateNormalSatellite(client, sat)
		if err != nil {
			return fmt.Errorf("CreateNormalSatellite err: %w", err)
		}
	case types.NetworkNodeType_ConsensusSatellite:
		sat, _ := node.ActualNode.(*satellites.ConsensusSatellite)
		err = CreateConsensusSatellite(client, sat)
		if err != nil {
			return fmt.Errorf("CreateConsensusSatellite err: %w", err)
		}
	case types.NetworkNodeType_EtcdService:
		etcdNode, _ := node.ActualNode.(*etcd.EtcdNode)
		err = CreateEtcdNode(client, etcdNode)
		if err != nil {
			return fmt.Errorf("CreateEtcdNode err: %w", err)
		}
	case types.NetworkNodeType_PositionService:
		positionService, _ := node.ActualNode.(*position.PositionService)
		err = CreatePositionService(client, positionService)
		if err != nil {
			return fmt.Errorf("CreatePositionService err: %w", err)
		}
	case types.NetworkNodeType_Router:
		router, _ := node.ActualNode.(*router.Router)
		err =
	}
	return nil
}


// CreatePositionService 创建位置更新/延迟更新服务
func CreatePositionService(client *docker.Client, positionService *position.PositionService) error {
	// 1. 检查状态
	if positionService.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("position service status is %s", positionService.Status)
	}

	ContainerName := "position_service"

	// 2. 创建环境变量
	envs := []string{
		fmt.Sprintf("%s=%s", "ETCD_LISTEN_ADDR", positionService.EtcdListenAddr),
		fmt.Sprintf("%s=%d", "ETCD_CLIENT_PORT", positionService.EtcdClientPort),
		fmt.Sprintf("%s=%s", "ETCD_ISLS_PREFIX", positionService.EtcdISLsPrefix),
		fmt.Sprintf("%s=%s", "ETCD_SATELLITES_PREFIX", positionService.EtcdSatellitesPrefix),
		fmt.Sprintf("%s=%s", "CONSTELLATION_START_TIME", positionService.ConstellationStartTime),
		fmt.Sprintf("%s=%d", "UPDATE_INTERVAL", positionService.UpdateInterval),
	}

	// 3. 创建 containerConfig
	containerConfig := &container.Config{
		Image: positionService.ImageName,
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

// CreateEtcdNode 创建 etcd 容器
func CreateEtcdNode(client *docker.Client, etcdNode *etcd.EtcdNode) error {
	// 1. 检查状态
	if etcdNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("etcd node status is %s", etcdNode.Status)
	}

	// 2. 创建变量
	EtcdAdvertiseClientUrls := fmt.Sprintf("http://0.0.0.0:%d", etcdNode.ClientPort)
	EtcdListenClientUrls := EtcdAdvertiseClientUrls
	EtcdListenPeerUrls := fmt.Sprintf("http://0.0.0.0:%d", etcdNode.PeerPort)
	EtcdInitialAdvertisePeerUrls := EtcdListenPeerUrls
	EtcdName := etcdNode.EtcdName
	EtcdInitialCluster := fmt.Sprintf("%s=%s", EtcdName, EtcdListenPeerUrls)
	EtcdDataDir := etcdNode.DataDir
	ContainerName := "etcd_service"

	// 3. 环境变量
	envs := []string{
		fmt.Sprintf("%s=%s", "ETCD_ADVERTISE_CLIENT_URLS", EtcdAdvertiseClientUrls),
		fmt.Sprintf("%s=%s", "ETCD_LISTEN_CLIENT_URLS", EtcdListenClientUrls),
		fmt.Sprintf("%s=%s", "ETCD_LISTEN_PEER_URLS", EtcdListenPeerUrls),
		fmt.Sprintf("%s=%s", "ETCD_INITIAL_ADVERTISE_PEER_URLS", EtcdInitialAdvertisePeerUrls),
		fmt.Sprintf("%s=%s", "ALLOW_NONE_AUTHENTICATION", "yes"),
		fmt.Sprintf("%s=%s", "ETCD_INITIAL_CLUSTER", EtcdInitialCluster),
		fmt.Sprintf("%s=%s", "ETCD_NAME", EtcdName),
		fmt.Sprintf("%s=%s", "ETCD_DATA_DIR", EtcdDataDir),
	}

	// 4. 暴露的端口
	clientPort := nat.Port(fmt.Sprintf("%d/tcp", etcdNode.ClientPort))
	peerPort := nat.Port(fmt.Sprintf("%d/tcp", etcdNode.PeerPort))

	// 5. 创建 containerConfig
	containerConfig := &container.Config{
		// 容器暴露的端口
		ExposedPorts: nat.PortSet{
			// rpc 端口
			clientPort: {},
			peerPort:   {},
		},
		Image: etcdNode.ImageName,
		Env:   envs,
		Tty:   true,
	}

	// 6. 创建 hostConfig
	hostConfig := &container.HostConfig{
		// 端口映射
		PortBindings: nat.PortMap{
			clientPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: string(clientPort),
				},
			},
			peerPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: string(peerPort),
				},
			},
		},
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
		ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create etcd container failed %v", err)
	}

	// 6. 从 response 之中获取 ID
	etcdNode.ContainerId = response.ID

	// 7. 进行状态的转换
	etcdNode.Status = types.NetworkNodeStatus_Created
	return nil
}


