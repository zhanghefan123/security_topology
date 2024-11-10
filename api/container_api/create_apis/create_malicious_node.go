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

func CreateMaliciousNode(client *docker.Client, maliciousNode *nodes.MaliciousNode) error {
	// 1. 检查状态
	if maliciousNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("malicious node not in logic status cannot create")
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

	// 创建端口映射
	httpPortInteger := 8080 + maliciousNode.Id
	httpPort := nat.Port(fmt.Sprintf("%d/tcp", 8080+maliciousNode.Id))

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

	// 3. 获取配置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, maliciousNode.ContainerName)
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr

	// 4. 创建容器卷映射
	volumes := []string{
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", maliciousNode.ContainerName)),
	}

	// 5. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", maliciousNode.Id),
		fmt.Sprintf("%s=%d", "LISTEN_PORT", httpPortInteger),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", maliciousNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(maliciousNode.Type), maliciousNode.Id, 1)),
	}

	// 6. containerConfig
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.MaliciousNodeImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
	}

	// 7. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:        volumes,
		CapAdd:       []string{"NET_ADMIN"},
		Privileged:   true,
		Sysctls:      sysctls,
		PortBindings: portBindings,
	}

	// 8. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		maliciousNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create malicious node failed %v", err)
	}

	maliciousNode.ContainerId = response.ID

	// 9. 状态转换
	maliciousNode.Status = types.NetworkNodeStatus_Created

	return nil
}
