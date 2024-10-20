package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/types"
)

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
		Image: configs.TopConfiguration.ImagesConfig.EtcdServiceImageName,
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
