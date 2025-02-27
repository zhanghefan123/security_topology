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

// CreateFabricOrderNode 创建 CreateFabriOrderNode
func CreateFabricOrderNode(client *docker.Client, fabricOrderNode *nodes.FabricOrderNode) error {
	// 1. 检查状态
	if fabricOrderNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("fabric orderer node not in logic status cannot create")
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
	// simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	// nodeDir := filepath.Join(simulationDir, fabricPeerNode.ContainerName)
	var cpuLimit float64
	var memoryLimit float64
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	fabricNetwork := configs.TopConfiguration.FabricConfig.FabricNetworkPath
	orderGeneralListenStartPort := configs.TopConfiguration.FabricConfig.OrderGeneralListenStartPort + fabricOrderNode.Id
	orderAdminListenStartPort := configs.TopConfiguration.FabricConfig.OrderAdminListenStartPort + fabricOrderNode.Id
	orderOperationListenStartPort := configs.TopConfiguration.FabricConfig.OrderOperationListenStartPort + fabricOrderNode.Id
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, fabricOrderNode.ContainerName)
	// ipv4 := strings.Split(fabricOrdererNode.Interfaces[0].Ipv4Addr, "/")[0]

	// 4. 创建容器卷映射
	volumes := []string{
		// fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", fabricordererNode.ContainerName)),
		fmt.Sprintf("%s:%s", fmt.Sprintf("%s/organizations/ordererOrganizations/example.com/orderers/orderer%d.example.com/msp", fabricNetwork, fabricOrderNode.Id),
			"/var/hyperledger/orderer/msp"),
		fmt.Sprintf("%s:%s", fmt.Sprintf("%s/organizations/ordererOrganizations/example.com/orderers/orderer%d.example.com/tls/", fabricNetwork, fabricOrderNode.Id),
			"/var/hyperledger/orderer/tls"),
		fmt.Sprintf("%s:%s", fmt.Sprintf("orderer%d.example.com", fabricOrderNode.Id), "/var/hyperledger/production/orderer"),
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", fabricOrderNode.ContainerName)),
	}

	// 5. 配置环境变量
	envs := []string{
		fmt.Sprintf("%s=%d", "NODE_ID", fabricOrderNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", fabricOrderNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(fabricOrderNode.Type), fabricOrderNode.Id, 1)),
		//add
		fmt.Sprintf("%s=%s", "FABRIC_LOGGING_SPEC", "INFO"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_LISTENADDRESS", "0.0.0.0"),
		fmt.Sprintf("%s=%d", "ORDERER_GENERAL_LISTENPORT", orderGeneralListenStartPort),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_LOCALMSPID", "OrdererMSP"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_LOCALMSPDIR", "/var/hyperledger/orderer/msp"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_TLS_ENABLED", "true"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_TLS_PRIVATEKEY", "/var/hyperledger/orderer/tls/server.key"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_TLS_CERTIFICATE", "/var/hyperledger/orderer/tls/server.crt"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_TLS_ROOTCAS", "[/var/hyperledger/orderer/tls/ca.crt]"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE", "/var/hyperledger/orderer/tls/server.crt"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY", "/var/hyperledger/orderer/tls/server.key"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_CLUSTER_ROOTCAS", "[/var/hyperledger/orderer/tls/ca.crt]"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_BOOTSTRAPMETHOD", "none"),
		fmt.Sprintf("%s=%s", "ORDERER_CHANNELPARTICIPATION_ENABLED", "true"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_TLS_ENABLED", "true"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_TLS_CERTIFICATE", "/var/hyperledger/orderer/tls/server.crt"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_TLS_PRIVATEKEY", "/var/hyperledger/orderer/tls/server.key"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_TLS_ROOTCAS", "[/var/hyperledger/orderer/tls/ca.crt]"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_TLS_CLIENTROOTCAS", "[/var/hyperledger/orderer/tls/ca.crt]"),
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_LISTENADDRESS", fmt.Sprintf("0.0.0.0:%d", orderAdminListenStartPort)),
		fmt.Sprintf("%s=%s", "ORDERER_OPERATIONS_LISTENADDRESS", fmt.Sprintf("orderer%d.example.com:%d", fabricOrderNode.Id, orderOperationListenStartPort)),
		// fmt.Sprintf("%s=%s", "ORDERER_OPERATIONS_LISTENADDRESS", fmt.Sprintf("%s:%d", ipv4, orderOperationListenStartPort)),
		fmt.Sprintf("%s=%s", "ORDERER_METRICS_PROVIDER", "prometheus"),
	}

	// 6. 资源限制
	resourcesLimit := container.Resources{
		NanoCPUs: int64(cpuLimit * 1e9),
		Memory:   int64(memoryLimit * 1024 * 1024), // memoryLimit 的单位是 MB
	}

	// 7. 创建端口映射
	generalPort := nat.Port(fmt.Sprintf("%d/tcp", orderGeneralListenStartPort))
	adminPort := nat.Port(fmt.Sprintf("%d/tcp", orderAdminListenStartPort))
	operationPort := nat.Port(fmt.Sprintf("%d/tcp", orderOperationListenStartPort))

	exposedPorts := nat.PortSet{
		generalPort:   {},
		adminPort:     {},
		operationPort: {},
	}

	portBindings := nat.PortMap{
		generalPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(generalPort),
			},
		},
		adminPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(adminPort),
			},
		},
		operationPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(operationPort),
			},
		},
	}

	// 8. 创建容器配置
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.FabricOrderImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
		Hostname:     fmt.Sprintf(fmt.Sprintf("orderer%d", fabricOrderNode.Id)),
		Domainname:   fmt.Sprintf("example.com"),
		// Cmd: []string{
		//     "peer node start",
		// },
	}
	// 9. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:      volumes,
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: true,
		Sysctls:    sysctls,
		// ExtraHosts:   []string{fmt.Sprintf("orderer%d.example.com:%s", fabricOrdererNode.Id, ipv4)},
		PortBindings: portBindings,
		Resources:    resourcesLimit,
		//指定宿主机作为DNS服务器
		DNS: []string{"192.168.47.128"},
	}
	// 10. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		fabricOrderNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create fabric orderer node failed %v", err)
	}

	fabricOrderNode.ContainerId = response.ID

	// 9. 状态转换
	fabricOrderNode.Status = types.NetworkNodeStatus_Created

	return nil
}
