package raspberrypi_topology

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"zhanghefan123/security_topology/api/route"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/raspberrypi_topology/protobuf"
)

type StartFunction func() error

type StartModule struct {
	start         bool
	startFunction StartFunction
}

const (
	SetInterfaceAddr                      = "SetInterfaceAddr"
	CalculateAndInstallStaticRoutes       = "CalculateAndInstallStaticRoutes"
	CalculateAndWriteLiRRoutes            = "CalculateAndWriteLiRRoutes"
	GenerateIfnameToLinkIdentifierMapping = "GenerateIfnameToLinkIdentifierMapping"
	GenerateAddressMapping                = "GenerateAddressMapping"
	GenerateSRv6Routes                    = "GenerateSRv6Routes"
	SetEnvs                               = "SetEnvs"
	SetSysctls                            = "SetSysctls"
	LoadKernelInfo                        = "LoadKernelInfo"
)

func (rpt *RaspberrypiTopology) Start() error {
	startSteps := []map[string]StartModule{
		{SetInterfaceAddr: StartModule{true, rpt.SetInterfaceAddr}},
		{CalculateAndInstallStaticRoutes: StartModule{true, rpt.CalculateStaticRoutes}},
		{GenerateIfnameToLinkIdentifierMapping: StartModule{true, rpt.GenerateIfnameToLinkIdentifierMapping}},
		{CalculateAndWriteLiRRoutes: StartModule{true, rpt.CalculateAndWriteLiRRoutes}},
		{GenerateAddressMapping: StartModule{true, rpt.GenerateAddressMapping}},
		{GenerateSRv6Routes: StartModule{true, rpt.GenerateSRv6Routes}},
		{SetEnvs: StartModule{true, rpt.SetEnv}},
		{SetSysctls: StartModule{true, rpt.SetSysctls}},
		{LoadKernelInfo: StartModule{true, rpt.LoadKernelInfo}},
	}
	err := rpt.startSteps(startSteps)
	if err != nil {
		return fmt.Errorf("start steps failed, %v", err)
	}
	return nil
}

func (rpt *RaspberrypiTopology) startStepsNum(startSteps []map[string]StartModule) int {
	result := 0
	for _, initStep := range startSteps {
		for _, initModule := range initStep {
			if initModule.start {
				result += 1
			}
		}
	}
	return result
}

func (rpt *RaspberrypiTopology) startSteps(startSteps []map[string]StartModule) (err error) {
	fmt.Println()
	moduleNum := rpt.startStepsNum(startSteps)
	for idx, initStep := range startSteps {
		for name, initModule := range initStep {
			if initModule.start {
				if err = initModule.startFunction(); err != nil {
					return fmt.Errorf("start step [%s] failed, %s", name, err)
				}
				raspberrypiTopologyLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

// SetInterfaceAddr 进行 IP 地址的配置
func (rpt *RaspberrypiTopology) SetInterfaceAddr() error {
	if _, ok := rpt.topologyStartSteps[SetInterfaceAddr]; ok {
		raspberrypiTopologyLogger.Infof("already set interface addr")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	// 创建 connection
	for index, abstractNode := range rpt.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}

		for _, networkInterface := range normalNode.Interfaces {
			fmt.Printf("set network interface: %s, addr: %s\n", networkInterface.IfName, networkInterface.SourceIpv4Addr)
			_, err = raspberrypiClient.SetAddr(context.Background(),
				&protobuf.SetAddrRequest{InterfaceName: networkInterface.IfName,
					InterfaceAddr: networkInterface.SourceIpv4Addr,
					AddrType:      "ipv4"})
			_, err = raspberrypiClient.SetAddr(context.Background(),
				&protobuf.SetAddrRequest{InterfaceName: networkInterface.IfName,
					InterfaceAddr: networkInterface.SourceIpv6Addr,
					AddrType:      "ipv6"})
		}

		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[SetInterfaceAddr] = struct{}{}
	raspberrypiTopologyLogger.Infof("set interface addr")
	return nil
}

func (rpt *RaspberrypiTopology) CalculateStaticRoutes() error {
	if _, ok := rpt.topologyStartSteps[CalculateAndInstallStaticRoutes]; ok {
		raspberrypiTopologyLogger.Infof("already calculate static routes")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}
		allStaticRoutes, err := route.GenerateStaticRoutes(abstractNode, &rpt.AllLinksMap, rpt.TopologyGraph)
		if err != nil {
			return fmt.Errorf("generate static routes failed: %w", err)
		}
		for _, staticRoute := range allStaticRoutes {
			var response *protobuf.NormalResponse
			response, err = raspberrypiClient.AddRoute(context.Background(), &protobuf.AddRouteRequest{
				DestinationNetworkSegment: staticRoute.DestinationNetworkSegment,
				Gateway:                   staticRoute.Gateway,
			})
			if err != nil {
				return fmt.Errorf("cannot add route: %v", err)
			}
			fmt.Println(response.Reply)
		}
	}

	rpt.topologyStartSteps[CalculateAndInstallStaticRoutes] = struct{}{}
	return nil
}

func (rpt *RaspberrypiTopology) CalculateAndWriteLiRRoutes() error {
	if _, ok := rpt.topologyStartSteps[CalculateAndWriteLiRRoutes]; ok {
		raspberrypiTopologyLogger.Infof("already calculate and write routes")
		return nil
	}

	// 所有的节点的 LiR 路由的列表形式
	allLiRRoutes := make([]string, 0)

	// 遍历所有节点生成路由文件
	for _, abstractNode := range rpt.AllAbstractNodes {
		// 获取单个节点的路由条目集合
		lirRoute, err := route.GenerateLiRRoute(abstractNode, &(rpt.AllLinksMap), rpt.TopologyGraph)
		if err != nil {
			return fmt.Errorf("generate path_validation route failed: %w", err)
		}
		// 更新总路由条目
		allLiRRoutes = append(allLiRRoutes, lirRoute)
	}

	// 所有的路由的字符串形式
	allLirRoutesString := strings.Join(allLiRRoutes, "\n")

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}
		var response *protobuf.NormalResponse
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/lir.txt", normalNode.ContainerName),
			Content:         allLiRRoutes[index],
		})
		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)

		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/all_lir.txt", normalNode.ContainerName),
			Content:         allLirRoutesString,
		})
		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)
		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[CalculateAndWriteLiRRoutes] = struct{}{}
	return nil
}

func (rpt *RaspberrypiTopology) GenerateAddressMapping() error {
	if _, ok := rpt.topologyStartSteps[GenerateAddressMapping]; ok {
		raspberrypiTopologyLogger.Infof("already generate address mapping")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}

		finalString := ""

		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}

		addressMapping, err := rpt.GetContainerNameToAddressMapping()
		if err != nil {
			return fmt.Errorf("get container name to address mapping failed, %v", err)
		}

		idMapping, err := rpt.GetContainerNameToGraphIdMapping()
		if err != nil {
			return fmt.Errorf("get container name to graph id mapping failed, %v", err)
		}

		for containerName, ipv4andipv6 := range addressMapping {
			graphId := idMapping[containerName]
			finalString += fmt.Sprintf("%s->%d->%s->%s\n", containerName, graphId, ipv4andipv6[0], ipv4andipv6[1])
		}

		var response *protobuf.NormalResponse
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/address_mapping.conf", normalNode.ContainerName),
			Content:         finalString,
		})
		if err != nil {
			return fmt.Errorf("error generating address mapping file %v", err)
		}
		fmt.Println(response.Reply)
		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[GenerateAddressMapping] = struct{}{}
	return nil
}

func (rpt *RaspberrypiTopology) GenerateIfnameToLinkIdentifierMapping() error {
	if _, ok := rpt.topologyStartSteps[GenerateIfnameToLinkIdentifierMapping]; ok {
		raspberrypiTopologyLogger.Infof("already generate ifname to link identifier mapping")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		finalString := ""
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %v", err)
		}
		for interfaceName, networkIntf := range normalNode.IfNameToInterfaceMap {
			finalString += fmt.Sprintf("%s->%d->%s\n", interfaceName, networkIntf.LinkIdentifier, networkIntf.TargetIpv4Addr)
		}
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}
		var response *protobuf.NormalResponse
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/interface/interface.txt", normalNode.ContainerName),
			Content:         finalString,
		})
		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)
		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[GenerateIfnameToLinkIdentifierMapping] = struct{}{}
	return nil
}

func (rpt *RaspberrypiTopology) GenerateSRv6Routes() error {
	if _, ok := rpt.topologyStartSteps[GenerateSRv6Routes]; ok {
		raspberrypiTopologyLogger.Infof("already generate srv6 routes")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		ipRouteStrings, destinationIpv6AddressMapping, destinationPathLengthMapping, err := route.GenerateSegmentRoutingStrings(abstractNode, &(rpt.AllLinksMap), rpt.TopologyGraph)
		if err != nil {
			return fmt.Errorf("generate segment routing strings failed: %w", err)
		}
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node failed: %w", err)
		}

		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}

		ipv6DestinationMappingString := ""
		var response *protobuf.NormalResponse
		for key, value := range destinationIpv6AddressMapping {
			ipv6DestinationMappingString += fmt.Sprintf("%s->%s->%s\n", key, value[0], value[1])
		}
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/ipv6_destination.txt", normalNode.ContainerName),
			Content:         ipv6DestinationMappingString,
		})
		if err != nil {
			return fmt.Errorf("error generating ipv6_destination.txt")
		}
		fmt.Println(response.Reply)

		srv6RouteString := strings.Join(ipRouteStrings, "\n")
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/srv6.txt", normalNode.ContainerName),
			Content:         srv6RouteString,
		})
		if err != nil {
			return fmt.Errorf("error generating srv6.txt")
		}
		fmt.Println(response.Reply)

		pathLengthMappingString := ""
		for key, value := range destinationPathLengthMapping {
			pathLengthMappingString += fmt.Sprintf("%s->%d\n", key, value)
		}
		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: fmt.Sprintf("/configuration/%s/route/destination_path_length.txt", normalNode.ContainerName),
			Content:         pathLengthMappingString,
		})
		if err != nil {
			return fmt.Errorf("error generating destination_path_length.txt")
		}
		fmt.Println(response.Reply)

		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[GenerateSRv6Routes] = struct{}{}
	return nil
}

// SetEnv 进行环境变量的设置
func (rpt *RaspberrypiTopology) SetEnv() error {
	if _, ok := rpt.topologyStartSteps[SetEnvs]; ok {
		raspberrypiTopologyLogger.Infof("already set env")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		// 进行 normalNode 的获取
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}

		// 创建环境变量列表
		envFields := []string{
			"CONTAINER_NAME",
			"BF_EFFECTIVE_BITS",
			"PVF_EFFECTIVE_BITS",
			"HASH_SEED",
			"NUMBER_OF_HASH_FUNCTIONS",
			"ROUTING_TABLE_TYPE",
			"NODE_ID", // 暂时没有使用到
			"GRAPH_NODE_ID",
			"LIR_SINGLE_TIME_ENCODING_COUNT",
			"ENABLE_SRV6",
		}
		envValues := []string{
			normalNode.ContainerName,
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.BfEffectiveBits),
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.PVFEffectiveBits),
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.HashSeed),
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.NumberOfHashFunctions),
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.RoutingTableType),
			strconv.Itoa(normalNode.Id),
			strconv.Itoa(int(abstractNode.Node.ID())),
			strconv.Itoa(configs.TopConfiguration.PathValidationConfig.LiRSingleTimeEncodingCount),
			fmt.Sprintf("%t", configs.TopConfiguration.NetworkConfig.EnableSRv6),
		}

		var response *protobuf.NormalResponse
		response, err = raspberrypiClient.SetEnv(context.Background(), &protobuf.SetEnvRequest{
			EnvFields: envFields,
			EnvValues: envValues,
		})
		if err != nil {
			return fmt.Errorf("error setting env %v", err)
		}
		fmt.Println(response.Reply)
	}

	rpt.topologyStartSteps[SetEnvs] = struct{}{}
	raspberrypiTopologyLogger.Infof("set env")
	return nil
}

func (rpt *RaspberrypiTopology) SetSysctls() error {
	if _, ok := rpt.topologyStartSteps[SetSysctls]; ok {
		raspberrypiTopologyLogger.Infof("already set sysctls")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, abstractNode := range rpt.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}

		var response *protobuf.NormalResponse
		sysctlFields := []string{
			// ipv4 相关配置
			"net.ipv4.ip_forward",
			"net.ipv4.conf.all.forwarding",
			// ipv6 相关配置
			"net.ipv6.conf.default.disable_ipv6",
			"net.ipv6.conf.all.disable_ipv6",
			"net.ipv6.conf.all.forwarding",
			"net.ipv6.conf.default.seg6_enabled",
			"net.ipv6.conf.all.seg6_enabled",
			"net.ipv6.conf.all.keep_addr_on_down",
			"net.ipv6.route.skip_notify_on_dev_down",
			"net.ipv4.conf.all.rp_filter",
			"net.ipv6.seg6_flowlabel",
		}
		sysctlValues := []int32{
			1, 1, 0, 0,
			1, 1, 1, 1,
			1, 0, 1,
		}
		for _, networkIntf := range normalNode.Interfaces {
			sysctlFields = append(sysctlFields, fmt.Sprintf("net.ipv6.conf.%s.seg6_enabled", networkIntf.IfName))
			sysctlValues = append(sysctlValues, 1)
		}

		response, err = raspberrypiClient.SetSysctls(context.Background(), &protobuf.SetSysctlsRequest{
			SysctlFields: sysctlFields,
			SysctlValues: sysctlValues,
		})

		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)

		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}

	rpt.topologyStartSteps[SetSysctls] = struct{}{}
	return nil
}

// LoadKernelInfo 加载内核信息
func (rpt *RaspberrypiTopology) LoadKernelInfo() error {
	if _, ok := rpt.topologyStartSteps[LoadKernelInfo]; ok {
		raspberrypiTopologyLogger.Infof("already load kernel info")
		return nil
	}

	// 监听端口
	grpcPort := configs.TopConfiguration.RaspberryPiConfig.GrpcPort

	for index, _ := range rpt.AllAbstractNodes {
		raspberrypiConn, err := CreateRaspberrypiConnection(fmt.Sprintf("%s:%d", configs.TopConfiguration.RaspberryPiConfig.IpAddresses[index], grpcPort))
		if err != nil {
			return fmt.Errorf("create raspberrypi connection failed")
		}
		raspberrypiClient, err := CreateRaspberrypiClient(raspberrypiConn)
		if err != nil {
			return fmt.Errorf("create raspberrypi client failed")
		}
		var response *protobuf.NormalResponse
		response, err = raspberrypiClient.LoadKernelInfo(context.Background(), &protobuf.LoadKernelInfoRequest{})
		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)
		err = raspberrypiConn.Close()
		if err != nil {
			return fmt.Errorf("close error")
		}
	}
	rpt.topologyStartSteps[LoadKernelInfo] = struct{}{}
	return nil
}
