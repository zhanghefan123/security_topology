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

func CreateFiscoBcosNode(client *docker.Client, fiscoBcosNode *nodes.FiscoBcosNode, graphNodeId int) error {
	if fiscoBcosNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("fisco bcos node not in logic status cannot create")
	}

	// 1 获取第一个接口
	firstInterfaceName := fiscoBcosNode.Interfaces[0].IfName
	firstInterfaceAddress := fiscoBcosNode.Interfaces[0].SourceIpv4Addr[:len(fiscoBcosNode.Interfaces[0].SourceIpv4Addr)-3]

	// 2 创建 sysctls
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

	// 3. 创建容器卷映射

	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, fiscoBcosNode.ContainerName)

	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", fiscoBcosNode.ContainerName)),
	}

	// 4. 配置环境变量
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	webPort := configs.TopConfiguration.ServicesConfig.WebConfig.StartPort + graphNodeId
	envs := []string{
		// zhf 添加的环境变量
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_NAME", firstInterfaceName),
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_ADDR", firstInterfaceAddress),
		fmt.Sprintf("%s=%d", "NODE_ID", fiscoBcosNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", fiscoBcosNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(fiscoBcosNode.Type), fiscoBcosNode.Id, 1)),
		fmt.Sprintf("%s=%d", "WEB_SERVER_LISTEN_PORT", webPort),
	}

	// 5. 资源限制
	cpuLimit := configs.TopConfiguration.ResourcesConfig.CpuLimit
	memoryLimit := configs.TopConfiguration.ResourcesConfig.MemoryLimit
	resourcesLimit := container.Resources{
		NanoCPUs: int64(cpuLimit * 1e9),
		Memory:   int64(memoryLimit * 1024 * 1024),
	}

	// 6. 端口映射 (现在暂时没有端口映射)
	exposedPorts := nat.PortSet{}

	portBindings := nat.PortMap{}

	// 7. 创建容器配置
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.FiscoBcosImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
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
		fiscoBcosNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create fisco bcos failed %v", err)
	}

	fiscoBcosNode.ContainerId = response.ID

	return nil
}
