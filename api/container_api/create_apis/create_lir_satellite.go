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

// CreateLiRSatellite 创建 lir satellite
func CreateLiRSatellite(client *docker.Client, lirSatellite *satellites.LiRSatellite, graphNodeId int) error {
	// 1. 检查状态
	if lirSatellite.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("lir satellite not in logic status cannot create")
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

	// 3. 创建端口映射
	httpPortInteger := 9000 + lirSatellite.Id
	httpPort := nat.Port(fmt.Sprintf("%d/tcp", httpPortInteger))

	exposedPorts := nat.PortSet{
		httpPort: {},
	}

	portBindings := nat.PortMap{
		httpPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(httpPort),
			},
		},
	}

	// 4. 获取配置
	// 4.1 获取一般配置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, lirSatellite.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	// 4.2 获取布隆过滤器相关配置
	bfEffectiveBits := configs.TopConfiguration.PathValidationConfig.BfEffectiveBits
	pvfEffectiveBits := configs.TopConfiguration.PathValidationConfig.PVFEffectiveBits
	hashSeed := configs.TopConfiguration.PathValidationConfig.HashSeed
	numberOfHashFunctions := configs.TopConfiguration.PathValidationConfig.NumberOfHashFunctions
	// 4.3 获取路由表类型
	routingTableType := configs.TopConfiguration.PathValidationConfig.RoutingTableType
	// 4.4 获取 serverport
	ipv6ServerPort := configs.TopConfiguration.AppsConfig.IPv6Config.ServerPort

	// 5. 创建容器卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", lirSatellite.ContainerName)),
	}

	// 6. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", lirSatellite.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", lirSatellite.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(lirSatellite.Type), lirSatellite.Id, 1)),
		fmt.Sprintf("%s=%d", "IPV6_SERVER_PORT", ipv6ServerPort),
		fmt.Sprintf("%s=%d", "GRAPH_NODE_ID", graphNodeId),
		fmt.Sprintf("%s=%d", "LISTEN_PORT", httpPortInteger),
		fmt.Sprintf("%s=%d", "BF_EFFECTIVE_BITS", bfEffectiveBits),
		fmt.Sprintf("%s=%d", "PVF_EFFECTIVE_BITS", pvfEffectiveBits),
		fmt.Sprintf("%s=%d", "HASH_SEED", hashSeed),
		fmt.Sprintf("%s=%d", "NUMBER_OF_HASH_FUNCTIONS", numberOfHashFunctions),
		fmt.Sprintf("%s=%d", "ROUTING_TABLE_TYPE", routingTableType),
		fmt.Sprintf("%s=%d", "LIR_SINGLE_TIME_ENCODING_COUNT", configs.TopConfiguration.PathValidationConfig.LiRSingleTimeEncodingCount),
		fmt.Sprintf("%s=%t", "ENABLE_SRV6", configs.TopConfiguration.NetworkConfig.EnableSRv6),
	}

	fmt.Println(fmt.Sprintf("%s=%d", "GRAPH_NODE_ID", graphNodeId))

	// 7. 容器配置
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.LiRSatelliteImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
	}

	// 8. 宿主机配置
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:        volumes,
		CapAdd:       []string{"NET_ADMIN"},
		Privileged:   true,
		Sysctls:      sysctls,
		PortBindings: portBindings,
	}

	// 9. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		lirSatellite.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create lirSatellite failed: %v", err)
	}

	lirSatellite.ContainerId = response.ID

	// 10. 状态转换
	lirSatellite.Status = types.NetworkNodeStatus_Created

	return nil
}
