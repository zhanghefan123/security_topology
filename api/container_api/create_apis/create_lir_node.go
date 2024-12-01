package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateLirNode 创建 LiRNode
func CreateLirNode(client *docker.Client, lirNode *nodes.LiRNode, graphNodeId int) error {
	// 1. 检查状态
	if lirNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("path_validation node not in logic status cannot create")
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
	httpPortInteger := 9000 + lirNode.Id
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
	nodeDir := filepath.Join(simulationDir, lirNode.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	// 4.2 获取布隆过滤器相关配置
	effectiveBits := configs.TopConfiguration.PathValidationConfig.EffectiveBits
	hashSeed := configs.TopConfiguration.PathValidationConfig.HashSeed
	numberOfHashFunctions := configs.TopConfiguration.PathValidationConfig.NumberOfHashFunctions
	// 4.3 获取路由表类型
	routingTableType := configs.TopConfiguration.PathValidationConfig.RoutingTableType

	// 5. 创建容器卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", lirNode.ContainerName)),
	}

	// 6. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", lirNode.Id),
		fmt.Sprintf("%s=%d", "GRAPH_NODE_ID", graphNodeId),
		fmt.Sprintf("%s=%d", "LISTEN_PORT", httpPortInteger),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", lirNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(lirNode.Type), lirNode.Id, 1)),
		fmt.Sprintf("%s=%d", "EFFECTIVE_BITS", effectiveBits),
		fmt.Sprintf("%s=%d", "HASH_SEED", hashSeed),
		fmt.Sprintf("%s=%d", "NUMBER_OF_HASH_FUNCTIONS", numberOfHashFunctions),
		fmt.Sprintf("%s=%d", "ROUTING_TABLE_TYPE", routingTableType),
	}
	// 7. 容器配置
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.LirNodeImageName,
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
		lirNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create path_validation node failed: %v", err)
	}

	lirNode.ContainerId = response.ID

	// 10. 状态转换
	lirNode.Status = types.NetworkNodeStatus_Created

	return nil
}
