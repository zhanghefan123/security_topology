package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/types"
)

// CreateChainMakerNode 创建长安链容器
func CreateChainMakerNode(client *docker.Client, chainMakerNode *nodes.ChainmakerNode) error {
	// 1. 检查状态
	if chainMakerNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("consensus node not in logic status cannot create")
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

	// 3. 获取配置
	var cpuLimit float64
	var memoryLimit float64
	chainMakerConfig := configs.TopConfiguration.ChainMakerConfig
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, chainMakerNode.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	absOfMultiNode := path.Join(chainMakerConfig.ChainMakerGoProjectPath, "scripts/docker/multi_node")
	cpuLimit = configs.TopConfiguration.ResourcesConfig.CpuLimit
	memoryLimit = configs.TopConfiguration.ResourcesConfig.MemoryLimit

	// 4. 创建容器卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", chainMakerNode.ContainerName)),

		fmt.Sprintf("%s:%s",
			fmt.Sprintf("%s/config/node%d", absOfMultiNode, chainMakerNode.Id),
			fmt.Sprintf("/chainmaker-go/config/wx-org%d.chainmaker.org", chainMakerNode.Id)),

		fmt.Sprintf("%s:%s",
			fmt.Sprintf("%s/data/data%d", absOfMultiNode, chainMakerNode.Id),
			"/chainmaker-go/data"),

		fmt.Sprintf("%s:%s",
			fmt.Sprintf("%s/log/log%d", absOfMultiNode, chainMakerNode.Id),
			"/chainmaker-go/log"),
	}

	// 5. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", chainMakerNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", chainMakerNode.ContainerName),
		fmt.Sprintf("%s=%d", "HTTP_PORT", chainMakerConfig.HttpStartPort+chainMakerNode.Id-1),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(chainMakerNode.Type), chainMakerNode.Id, 1)),
		fmt.Sprintf("%s=%s", "LISTEN_ADDR", chainMakerNode.Interfaces[0].Ipv4Addr),
		fmt.Sprintf("%s=%t", "SPEED_CHECK", chainMakerConfig.SpeedCheck),
		fmt.Sprintf("%s=%t", "ENABLE_BROADCAST_DEFENCE", chainMakerConfig.EnableBroadcastDefence),
		fmt.Sprintf("%s=%f", "DDOS_WARNING_RATE", chainMakerConfig.DdosWarningRate),
		fmt.Sprintf("%s=%t", "DIRECT_REMOVE", chainMakerConfig.DirectRemoveAttackedNode),
		fmt.Sprintf("%s=%d", "BLOCKS_PER_PROPOSER", chainMakerConfig.BlocksPerProposer),
	}

	// 6. 资源限制
	resourcesLimit := container.Resources{
		NanoCPUs: int64(cpuLimit * 1e9),
		Memory:   int64(memoryLimit * 1024 * 1024), // memoryLimit 的单位是 MB
	}

	// 7. 创建端口映射
	rpcPort := nat.Port(fmt.Sprintf("%d/tcp", chainMakerConfig.RpcStartPort+chainMakerNode.Id-1))
	httpPort := nat.Port(fmt.Sprintf("%d/tcp", chainMakerConfig.HttpStartPort+chainMakerNode.Id-1))

	exposedPorts := nat.PortSet{
		rpcPort:  {},
		httpPort: {},
	}

	portBindings := nat.PortMap{
		rpcPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(rpcPort),
			},
		},
		httpPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(httpPort),
			},
		},
	}

	// 8. 创建容器配置
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.ChainMakerNodeImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
		Cmd: []string{
			"./chainmaker",
			"start",
			"-c",
			fmt.Sprintf("../config/wx-org%d.chainmaker.org/chainmaker.yml", chainMakerNode.Id),
		},
	}

	// 8. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:        volumes,
		CapAdd:       []string{"NET_ADMIN"},
		Privileged:   true,
		Sysctls:      sysctls,
		PortBindings: portBindings,
		Resources:    resourcesLimit,
	}

	// 9. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		chainMakerNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create consensus node failed %v", err)
	}

	chainMakerNode.ContainerId = response.ID

	// 10. 状态转换
	chainMakerNode.Status = types.NetworkNodeStatus_Created

	return nil
}
