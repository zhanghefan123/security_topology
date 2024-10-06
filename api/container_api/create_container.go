package container_api

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/position"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateContainer 创建 container
func CreateContainer(client *docker.Client, node *node.AbstractNode) error {
	var err error = nil
	if node.Type == types.NetworkNodeType_NormalSatellite {
		sat, _ := node.ActualNode.(*satellite.NormalSatellite)
		err = CreateNormalSatellite(client, sat)
	} else if node.Type == types.NetworkNodeType_ConsensusSatellite {
		sat, _ := node.ActualNode.(*satellite.ConsensusSatellite)
		err = CreateConsensusSatellite(client, sat)
	} else if node.Type == types.NetworkNodeType_EtcdService {
		etcdNode, _ := node.ActualNode.(*etcd.EtcdNode)
		err = CreateEtcdNode(client, etcdNode)
	} else if node.Type == types.NetworkNodeType_PositionService {
		positionService, _ := node.ActualNode.(*position.PositionService)
		err = CreatePositionService(client, positionService)
	} else {
		err = fmt.Errorf("not supported type")
	}
	return err
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

// CreateNormalSatellite 创建普通卫星容器
func CreateNormalSatellite(client *docker.Client, satellite *satellite.NormalSatellite) error {
	// 1. 检查状态
	if satellite.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("normal satellite not in logic status cannot create")
	}

	// 2. 创建 sysctls
	sysctls := map[string]string{
		// ipv4 的相关网络配置
		"net.ipv4.ip_forward":          "1",
		"net.ipv4.conf.all.forwarding": "1",

		// ipv6 的相关网络配置
		"net.ipv6.conf.default.disable_ipv6":     "0",
		"net.ipv6.conf.all.disable_ipv6":         "0",
		"net.ipv6.conf.all.forwarding":           "1",
		"net.ipv6.conf.default.seg6_enabled":     "1",
		"net.ipv6.conf.eth0.seg6_enabled":        "1",
		"net.ipv6.conf.lo.seg6_enabled":          "1",
		"net.ipv6.conf.all.seg6_enabled":         "1",
		"net.ipv6.conf.all.keep_addr_on_down":    "1",
		"net.ipv6.route.skip_notify_on_dev_down": "1",
		"net.ipv4.conf.all.rp_filter":            "0",
		"net.ipv6.seg6_flowlabel":                "1",
	}

	// 3. 创建容器
	//容器数据卷映射
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, satellite.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", satellite.ContainerName)),
	}

	// 4. 环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", satellite.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", satellite.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
	}

	// 5. containerConfig
	containerConfig := &container.Config{
		Image: satellite.ImageName,
		Tty:   true,
		Env:   envs,
	}

	// 6. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:      volumes,
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: true,
		Sysctls:    sysctls,
	}

	// 7. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		satellite.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create satellite container failed %v", err)
	}

	satellite.ContainerId = response.ID

	// 8. 状态转换
	satellite.Status = types.NetworkNodeStatus_Created

	return nil
}

// CreateConsensusSatellite 创建共识卫星容器
func CreateConsensusSatellite(client *docker.Client, satellite *satellite.ConsensusSatellite) error {
	// 1. 检查状态
	if satellite.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("consensus satellite not in logic status cannot create")
	}

	// 2. 创建 sysctls
	sysctls := map[string]string{
		// ipv4 的相关网络配置
		"net.ipv4.ip_forward":          "1",
		"net.ipv4.conf.all.forwarding": "1",

		// ipv6 的相关网络配置
		"net.ipv6.conf.default.disable_ipv6":     "0",
		"net.ipv6.conf.all.disable_ipv6":         "0",
		"net.ipv6.conf.all.forwarding":           "1",
		"net.ipv6.conf.default.seg6_enabled":     "1",
		"net.ipv6.conf.eth0.seg6_enabled":        "1",
		"net.ipv6.conf.lo.seg6_enabled":          "1",
		"net.ipv6.conf.all.seg6_enabled":         "1",
		"net.ipv6.conf.all.keep_addr_on_down":    "1",
		"net.ipv6.route.skip_notify_on_dev_down": "1",
		"net.ipv4.conf.all.rp_filter":            "0",
		"net.ipv6.seg6_flowlabel":                "1",
	}

	// 3. 创建容器
	// 容器数据卷映射
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, satellite.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr

	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", satellite.ContainerName)),
	}

	// 4. 环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", satellite.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", satellite.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
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
		Binds:      volumes,
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: true,
		Sysctls:    sysctls,
	}

	// 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		satellite.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create consensus satellite container failed %v", err)
	}

	satellite.ContainerId = response.ID

	// 3. 状态转换
	satellite.Status = types.NetworkNodeStatus_Created

	return nil
}
