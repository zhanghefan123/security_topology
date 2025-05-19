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

func CreateFabricPeerNode(client *docker.Client, fabricPeerNode *nodes.FabricPeerNode) error {
	// 1. 检查状态
	if fabricPeerNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("fabric peer node not in logic status cannot create")
	}

	firstInterfaceName := fabricPeerNode.Interfaces[0].IfName
	firstInterfaceAddress := fabricPeerNode.Interfaces[0].SourceIpv4Addr[:len(fabricPeerNode.Interfaces[0].SourceIpv4Addr)-3]
	fmt.Printf("Node Name: %s Addr: %s \n", fabricPeerNode.ContainerName, firstInterfaceAddress)

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
	peerListenPort := configs.TopConfiguration.FabricConfig.PeerListenStartPort + fabricPeerNode.Id
	peerChainCodePort := configs.TopConfiguration.FabricConfig.PeerChaincodeStartPort + fabricPeerNode.Id
	peerOperationPort := configs.TopConfiguration.FabricConfig.PeerOperationStartPort + fabricPeerNode.Id
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, fabricPeerNode.ContainerName)
	// ipv4 := strings.Split(fabricPeerNode.Interfaces[0].Ipv4Addr, "/")[0]

	// 4. 创建容器卷映射
	volumes := []string{
		// fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", fabricPeerNode.ContainerName)),
		fmt.Sprintf("%s:%s", fmt.Sprintf("%s/organizations/peerOrganizations/org%d.example.com/peers/peer0.org%d.example.com", fabricNetwork,
			fabricPeerNode.Id, fabricPeerNode.Id),
			"/etc/hyperledger/fabric"),
		fmt.Sprintf("%s:%s", fmt.Sprintf("peer0.org%d.example.com", fabricPeerNode.Id),
			"/var/hyperledger/production"),
		fmt.Sprintf("%s:%s", fmt.Sprintf("%s/compose/docker/peercfg", fabricNetwork),
			"/etc/hyperledger/peercfg"),
		fmt.Sprintf("%s:%s", "/var/run/docker.sock", "/host/var/run/docker.sock"),
		fmt.Sprintf("%s:%s", nodeDir, fmt.Sprintf("/configuration/%s", fabricPeerNode.ContainerName)),
	}
	// fmt.Printf("intenfaceNum:%d\n", len(fabricPeerNode.Interfaces))
	// fmt.Printf("Ipv4:%s\n", fabricPeerNode.Interfaces[0].Ipv4Addr)
	// 5. 配置环境变量
	envs := []string{
		// zhf 添加的环境变量
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_NAME", firstInterfaceName),
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_ADDR", firstInterfaceAddress),

		fmt.Sprintf("%s=%d", "NODE_ID", fabricPeerNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", fabricPeerNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%s", "INTERFACE_NAME", fmt.Sprintf("%s%d_idx%d", types.GetPrefix(fabricPeerNode.Type), fabricPeerNode.Id, 1)),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		//add
		fmt.Sprintf("%s=%s", "FABRIC_CFG_PATH", "/etc/hyperledger/peercfg"),
		fmt.Sprintf("%s=%s", "FABRIC_LOGGING_SPEC", "INFO"),
		fmt.Sprintf("%s=%s", "CORE_PEER_TLS_ENABLED", "true"),
		fmt.Sprintf("%s=%s", "CORE_PEER_PROFILE_ENABLED", "false"),
		fmt.Sprintf("%s=%s", "CORE_PEER_TLS_CERT_FILE", "/etc/hyperledger/fabric/tls/server.crt"),
		fmt.Sprintf("%s=%s", "CORE_PEER_TLS_KEY_FILE", "/etc/hyperledger/fabric/tls/server.key"),
		fmt.Sprintf("%s=%s", "CORE_PEER_TLS_ROOTCERT_FILE", "/etc/hyperledger/fabric/tls/ca.crt"),
		fmt.Sprintf("%s=%s", "CORE_PEER_ID", fmt.Sprintf("peer0.org%d.example.com", fabricPeerNode.Id)),
		fmt.Sprintf("%s=%s", "CORE_PEER_ADDRESS", fmt.Sprintf("%s:%d", "0.0.0.0", peerListenPort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_LISTENADDRESS", fmt.Sprintf("%s:%d", "0.0.0.0", peerListenPort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_CHAINCODEADDRESS", fmt.Sprintf("%s:%d", firstInterfaceAddress, peerChainCodePort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_CHAINCODELISTENADDRESS", fmt.Sprintf("0.0.0.0:%d", peerChainCodePort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_GOSSIP_EXTERNALENDPOINT", fmt.Sprintf("%s:%d", firstInterfaceAddress, peerListenPort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_GOSSIP_BOOTSTRAP", fmt.Sprintf("%s:%d", firstInterfaceAddress, peerListenPort)),
		fmt.Sprintf("%s=%s", "CORE_PEER_LOCALMSPID", fmt.Sprintf("Org%dMSP", fabricPeerNode.Id)),
		fmt.Sprintf("%s=%s", "CORE_PEER_MSPCONFIGPATH", "/etc/hyperledger/fabric/msp"),
		fmt.Sprintf("%s=%s", "CORE_OPERATIONS_LISTENADDRESS", fmt.Sprintf("%s:%d", firstInterfaceAddress, peerOperationPort)),
		fmt.Sprintf("%s=%s", "CORE_METRICS_PROVIDER", "prometheus"),
		fmt.Sprintf("%s=%s", "CHAINCODE_AS_A_SERVICE_BUILDER_CONFIG", fmt.Sprintf("{\"peername\":\"peer0org%d\"}", fabricPeerNode.Id)),
		fmt.Sprintf("%s=%s", "CORE_CHAINCODE_EXECUTETIMEOUT", "300s"),
		fmt.Sprintf("%s=%s", "CORE_VM_ENDPOINT", "unix:///host/var/run/docker.sock"),
		// fmt.Sprintf("%s=%s", "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE", "fabric_test"),
	}

	// 6. 资源限制
	resourcesLimit := container.Resources{
		NanoCPUs: int64(cpuLimit * 1e9),
		Memory:   int64(memoryLimit * 1024 * 1024), // memoryLimit 的单位是 MB
	}

	// 7. 创建端口映射
	listenPort := nat.Port(fmt.Sprintf("%d/tcp", peerListenPort))
	operationPort := nat.Port(fmt.Sprintf("%d/tcp", peerOperationPort))

	exposedPorts := nat.PortSet{
		listenPort:    {},
		operationPort: {},
	}

	portBindings := nat.PortMap{
		listenPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(listenPort),
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
		Image:        configs.TopConfiguration.ImagesConfig.FabricPeerImageName,
		Hostname:     fmt.Sprintf("peer0"),
		Domainname:   fmt.Sprintf("org%d.example.com", fabricPeerNode.Id),
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
		// Cmd: []string{
		//     "peer", "node", "start",
		// },
	}
	// 9. hostConfig
	hostConfig := &container.HostConfig{
		// 容器数据卷映射
		Binds:      volumes,
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: true,
		Sysctls:    sysctls,
		// ExtraHosts:   []string{fmt.Sprintf("peer0.org%d.example.com:%s", fabricPeerNode.Id, ipv4)},
		PortBindings: portBindings,
		Resources:    resourcesLimit,
		//指定宿主机作为DNS服务器
		DNS: []string{"10.134.86.192"},
	}
	// 10. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		fabricPeerNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create fabric peer node failed %v", err)
	}

	fabricPeerNode.ContainerId = response.ID

	// 9. 状态转换
	fabricPeerNode.Status = types.NetworkNodeStatus_Created

	return nil

}
