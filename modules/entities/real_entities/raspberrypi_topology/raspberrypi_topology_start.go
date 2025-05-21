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
	SetEnvs                               = "SetEnvs"
	LoadKernelInfo                        = "LoadKernelInfo"
)

func (rpt *RaspberrypiTopology) Start() error {
	startSteps := []map[string]StartModule{
		{SetInterfaceAddr: StartModule{true, rpt.SetInterfaceAddr}},
		{CalculateAndInstallStaticRoutes: StartModule{true, rpt.CalculateStaticRoutes}},
		{GenerateIfnameToLinkIdentifierMapping: StartModule{true, rpt.GenerateIfnameToLinkIdentifierMapping}},
		{CalculateAndWriteLiRRoutes: StartModule{true, rpt.CalculateAndWriteLiRRoutes}},
		{SetEnvs: StartModule{true, rpt.SetEnv}},
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
			var response *protobuf.NormalResponse
			response, err = raspberrypiClient.SetAddr(context.Background(), &protobuf.SetAddrRequest{InterfaceName: networkInterface.IfName, InterfaceAddr: networkInterface.SourceIpv4Addr})
			if err != nil {
				return fmt.Errorf("error setting addr %v", err)
			}
			fmt.Println(response.Reply)
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

	for index, _ := range configs.TopConfiguration.RaspberryPiConfig.IpAddresses {
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
			DestinationPath: "/home/zeusnet/Projects/configuration/route/lir.txt",
			Content:         allLiRRoutes[index],
		})
		if err != nil {
			return fmt.Errorf("error setting addr %v", err)
		}
		fmt.Println(response.Reply)

		response, err = raspberrypiClient.TransmitFile(context.Background(), &protobuf.TransmitFileRequest{
			DestinationPath: "/home/zeusnet/Projects/configuration/route/all_lir.txt",
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
			DestinationPath: "/home/zeusnet/Projects/configuration/interface/interface.txt",
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

// SetEnv 进行环境变量的设置
func (rpt *RaspberrypiTopology) SetEnv() error {
	if _, ok := rpt.topologyStartSteps["SetEnv"]; ok {
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

	rpt.topologyStartSteps["SetEnv"] = struct{}{}
	raspberrypiTopologyLogger.Infof("set env")
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
