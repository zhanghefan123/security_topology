package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateConsensusSatellite 创建共识卫星容器
func CreateConsensusSatellite(client *docker.Client, consensusSatellite *satellites.ConsensusSatellite) error {
	// 1. 检查状态
	if consensusSatellite.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("consensus consensusSatellite not in logic status cannot create")
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
	nodeDir := filepath.Join(simulationDir, consensusSatellite.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr

	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", consensusSatellite.ContainerName)),
	}

	// 4. 环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", consensusSatellite.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", consensusSatellite.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(consensusSatellite.Type), consensusSatellite.Id, 1)),
	}

	// 暴露的端口
	rpcPort := nat.Port(fmt.Sprintf("%d/tcp", consensusSatellite.StartRpcPort+consensusSatellite.Id-1)) // 起始的端口为 11301 那么第一个共识卫星应该监听的端口为 11301
	p2pPort := nat.Port(fmt.Sprintf("%d/tcp", consensusSatellite.StartP2PPort+consensusSatellite.Id-1))

	// containerConfig
	containerConfig := &container.Config{
		Image: configs.TopConfiguration.ImagesConfig.ConsensusSatelliteImageName,
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
		consensusSatellite.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create consensus consensusSatellite container failed %v", err)
	}

	consensusSatellite.ContainerId = response.ID

	// 3. 状态转换
	consensusSatellite.Status = types.NetworkNodeStatus_Created

	return nil
}
