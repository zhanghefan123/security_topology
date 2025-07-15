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
func CreateFabricOrderNode(client *docker.Client, fabricOrderNode *nodes.FabricOrderNode, graphNodeId int) error {
	// 1. 检查状态
	if fabricOrderNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("fabric orderer node not in logic status cannot create")
	}

	firstInterfaceName := fabricOrderNode.Interfaces[0].IfName
	firstInterfaceAddress := fabricOrderNode.Interfaces[0].SourceIpv4Addr[:len(fabricOrderNode.Interfaces[0].SourceIpv4Addr)-3]
	fmt.Printf("Node Name: %s Addr: %s \n", fabricOrderNode.ContainerName, firstInterfaceAddress)

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
	orderGeneralListenPort := configs.TopConfiguration.FabricConfig.OrderGeneralListenStartPort + fabricOrderNode.Id
	orderAdminListenPort := configs.TopConfiguration.FabricConfig.OrderAdminListenStartPort + fabricOrderNode.Id
	orderOperationListenPort := configs.TopConfiguration.FabricConfig.OrderOperationListenStartPort + fabricOrderNode.Id
	orderPprofListenPort := configs.TopConfiguration.FabricConfig.PprofOrdererStartListenPort + fabricOrderNode.Id
	enablePprof := configs.TopConfiguration.FabricConfig.EnablePprof
	enableRoutine := configs.TopConfiguration.FabricConfig.EnableRoutine
	enableAdvancedMessageHandler := configs.TopConfiguration.FabricConfig.EnableAdvancedMessageHandler
	webPort := configs.TopConfiguration.ServicesConfig.WebConfig.StartPort + graphNodeId
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
		// zhf add code
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_NAME", firstInterfaceName),
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_ADDR", firstInterfaceAddress),
		fmt.Sprintf("%s=%d", "NODE_ID", fabricOrderNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", fabricOrderNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(fabricOrderNode.Type), fabricOrderNode.Id, 1)),
		fmt.Sprintf("%s=%s", "FABRIC_LOGGING_SPEC", "ERROR"),
		fmt.Sprintf("%s=%s", "ORDERER_GENERAL_LISTENADDRESS", "0.0.0.0"),
		fmt.Sprintf("%s=%d", "ORDERER_GENERAL_LISTENPORT", orderGeneralListenPort),
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
		// 现在的版本
		fmt.Sprintf("%s=%s", "ORDERER_ADMIN_LISTENADDRESS", fmt.Sprintf("0.0.0.0:%d", orderAdminListenPort)), // 这个地址需要设置成 0.0.0.0 供宿主机访问
		fmt.Sprintf("%s=%s", "ORDERER_OPERATIONS_LISTENADDRESS", fmt.Sprintf("orderer%d.example.com:%d", fabricOrderNode.Id, orderOperationListenPort)),
		fmt.Sprintf("%s=%s", "ORDERER_METRICS_PROVIDER", "prometheus"),
		// zhf add code
		fmt.Sprintf("%s=%t", "ENABLE_ROUTINE", enableRoutine),
		fmt.Sprintf("%s=%t", "ENABLE_ADVANCED_MESSAGE_HANDLER", enableAdvancedMessageHandler),
		fmt.Sprintf("%s=%t", "ENABLE_PPROF", enablePprof),
		fmt.Sprintf("%s=%d", "WEB_SERVER_LISTEN_PORT", webPort),
		fmt.Sprintf("%s=%d", "PPROF_ORDERER_LISTEN_PORT", orderPprofListenPort),
	}

	// 6. 资源限制
	resourcesLimit := container.Resources{
		NanoCPUs: int64(cpuLimit * 1e9),
		Memory:   int64(memoryLimit * 1024 * 1024), // memoryLimit 的单位是 MB
	}

	// 7. 创建端口映射
	generalPort := nat.Port(fmt.Sprintf("%d/tcp", orderGeneralListenPort))
	adminPort := nat.Port(fmt.Sprintf("%d/tcp", orderAdminListenPort))
	operationPort := nat.Port(fmt.Sprintf("%d/tcp", orderOperationListenPort))
	webServerPort := nat.Port(fmt.Sprintf("%d/tcp", webPort))
	pprofListenPort := nat.Port(fmt.Sprintf("%d/tcp", orderPprofListenPort))

	exposedPorts := nat.PortSet{
		generalPort:     {},
		adminPort:       {},
		operationPort:   {},
		webServerPort:   {},
		pprofListenPort: {},
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
		webServerPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(webServerPort),
			},
		},
		pprofListenPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(pprofListenPort),
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
	}
	// 9. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:        volumes,
		CapAdd:       []string{"NET_ADMIN"},
		Privileged:   true,
		Sysctls:      sysctls,
		PortBindings: portBindings,
		Resources:    resourcesLimit,
		//指定宿主机作为DNS服务器
		DNS: []string{configs.TopConfiguration.NetworkConfig.LocalNetworkAddress},
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
