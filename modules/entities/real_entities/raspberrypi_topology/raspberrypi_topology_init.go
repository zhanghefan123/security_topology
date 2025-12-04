package raspberrypi_topology

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"strconv"
	"strings"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/services/http/params"
	"zhanghefan123/security_topology/utils/network"
)

const (
	GenerateNodes   = "GenerateNodes"
	GenerateSubnets = "GenerateSubnets"
	GenerateLinks   = "GenerateLinks"
	PrintTopology   = "PrintTopology"
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

// Init 进行初始化
func (rpt *RaspberrypiTopology) Init() error {
	initSteps := []map[string]InitModule{
		{GenerateNodes: InitModule{true, rpt.GenerateNodes}},
		{GenerateSubnets: InitModule{true, rpt.GenerateSubnets}},
		{GenerateLinks: InitModule{true, rpt.GenerateLinks}},
		{PrintTopology: InitModule{true, rpt.PrintTopology}},
	}
	err := rpt.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("constellation init failed: %v", err)
	}
	return nil
}

func (rpt *RaspberrypiTopology) initStepsNum(initSteps []map[string]InitModule) int {
	result := 0
	for _, initStep := range initSteps {
		for _, initModule := range initStep {
			if initModule.init {
				result += 1
			}
		}
	}
	return result
}

// InitializeSteps 按步骤进行初始化
func (rpt *RaspberrypiTopology) initializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := rpt.initStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.init {
				if err = initModule.initFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				raspberrypiTopologyLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

func (rpt *RaspberrypiTopology) GenerateNodes() error {
	if _, ok := rpt.topologyInitSteps[GenerateNodes]; ok {
		raspberrypiTopologyLogger.Infof("already generate nodes")
		return nil
	}

	for index, nodeTypeString := range configs.TopConfiguration.RaspberryPiConfig.NodeTypes {
		nodeId := configs.TopConfiguration.RaspberryPiConfig.NodeIDs[index]
		nodeType, err := types.ResolveNodeType(nodeTypeString)
		if err != nil {
			return fmt.Errorf("cannot resolve node type")
		}
		switch *nodeType {
		case types.NetworkNodeType_Router: // 进行普通路由节点的添加
			routerTmp := nodes.NewRouter(nodeId, 0, 0)
			rpt.Routers = append(rpt.Routers, routerTmp)
			// 注意只能唯一创建一次
			abstractRouter := node.NewAbstractNode(types.NetworkNodeType_Router, routerTmp, rpt.TopologyGraph)
			rpt.RouterAbstractNodes = append(rpt.RouterAbstractNodes, abstractRouter)
			rpt.AllAbstractNodes = append(rpt.AllAbstractNodes, abstractRouter)
			rpt.AbstractNodesMap[routerTmp.ContainerName] = abstractRouter
		case types.NetworkNodeType_LirNode:
			lirNodeTmp := nodes.NewLiRNode(nodeId, 0, 0)
			rpt.LirNodes = append(rpt.LirNodes, lirNodeTmp)
			// 注意只能唯一创建一次
			abstractLirNode := node.NewAbstractNode(types.NetworkNodeType_LirNode, lirNodeTmp, rpt.TopologyGraph)
			rpt.LirAbstractNodes = append(rpt.LirAbstractNodes, abstractLirNode)
			rpt.AllAbstractNodes = append(rpt.AllAbstractNodes, abstractLirNode)
			rpt.AbstractNodesMap[lirNodeTmp.ContainerName] = abstractLirNode
		}
	}

	rpt.topologyInitSteps[GenerateNodes] = struct{}{}
	raspberrypiTopologyLogger.Infof("generate nodes")
	return nil
}

func (rpt *RaspberrypiTopology) GenerateSubnets() error {
	if _, ok := rpt.topologyInitSteps[GenerateSubnets]; ok {
		raspberrypiTopologyLogger.Infof("already generate subnets")
		return nil
	}
	var err error
	var ipv4Subnets []iplib.Net4
	var ipv6Subnets []iplib.Net6

	// 进行 ipv4 的子网的生成
	ipv4Subnets, err = network.GenerateIPv4Subnets(configs.TopConfiguration.NetworkConfig.BaseV4NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	rpt.Ipv4SubNets = ipv4Subnets

	// 进行 ipv6 的子网的生成
	ipv6Subnets, err = network.GenerateIpv6Subnets(configs.TopConfiguration.NetworkConfig.BaseV6NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	rpt.Ipv6SubNets = ipv6Subnets

	rpt.topologyInitSteps[GenerateSubnets] = struct{}{}
	raspberrypiTopologyLogger.Infof("generate subnets")
	return nil
}

func (rpt *RaspberrypiTopology) ResolveLinkInformation(connection string) (*params.LinkParam, error) {
	endPoints := strings.Split(connection, "-")
	sourceIndexAndIntf := strings.Split(endPoints[0], ":")
	targetIndexAndIntf := strings.Split(endPoints[1], ":")
	sourceIndex, err := strconv.ParseInt(sourceIndexAndIntf[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot parse index: %v", err)
	}
	sourceInterfaceName := sourceIndexAndIntf[1]
	targetIndex, err := strconv.ParseInt(targetIndexAndIntf[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot parse index: %v", err)
	}
	targetInterfaceName := targetIndexAndIntf[1]
	linkParam := &params.LinkParam{
		LinkType: "access",
		SourceNode: params.NodeParam{
			Index: int(sourceIndex),
			Type:  configs.TopConfiguration.RaspberryPiConfig.NodeTypes[sourceIndex-1],
			X:     0,
			Y:     0,
		},
		TargetNode: params.NodeParam{
			Index: int(targetIndex),
			Type:  configs.TopConfiguration.RaspberryPiConfig.NodeTypes[targetIndex-1],
			X:     0,
			Y:     0,
		},
		SourceInterfaceName: sourceInterfaceName,
		TargetInterfaceName: targetInterfaceName,
	}
	return linkParam, nil
}

// getSourceNodeAndTargetNode 获取源和目的抽象节点
func (rpt *RaspberrypiTopology) getSourceNodeAndTargetNode(sourceNodeParam, targetNodeParam params.NodeParam) (*node.AbstractNode, *node.AbstractNode, error) {
	var sourceNode, targetNode *node.AbstractNode
	sourceNodeType, err := types.ResolveNodeType(sourceNodeParam.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve source node type failed, %s", err)
	}
	targetNodeType, err := types.ResolveNodeType(targetNodeParam.Type)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve target node type failed, %s", err)
	}
	// 拿到源节点
	switch *sourceNodeType {
	case types.NetworkNodeType_Router:
		sourceNode = rpt.RouterAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_LirNode:
		sourceNode = rpt.LirAbstractNodes[sourceNodeParam.Index-1]
	default:
		return nil, nil, fmt.Errorf("unsupported source node type: %s", *sourceNodeType)
	}

	// 拿到目的节点
	switch *targetNodeType {
	case types.NetworkNodeType_Router:
		targetNode = rpt.RouterAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_LirNode:
		targetNode = rpt.LirAbstractNodes[targetNodeParam.Index-1]
	default:
		return nil, nil, fmt.Errorf("unsupported target node type: %s", *sourceNodeType)
	}

	return sourceNode, targetNode, nil
}

func (rpt *RaspberrypiTopology) GenerateLinks() error {
	if _, ok := rpt.topologyInitSteps[GenerateLinks]; ok {
		raspberrypiTopologyLogger.Infof("already generate links")
		return nil
	}

	// ----------------实际逻辑--------------------
	allLinkParams := make([]*params.LinkParam, 0)
	for _, linkInStr := range configs.TopConfiguration.RaspberryPiConfig.Connections {
		linkParam, err := rpt.ResolveLinkInformation(linkInStr)
		if err != nil {
			return fmt.Errorf("resolve link information failed: %v", err)
		}
		allLinkParams = append(allLinkParams, linkParam)
	}

	for _, linkTmp := range allLinkParams {
		// 拿到从前端传递过来的 (源节点 目的节点的参数)
		sourceNodeParam := linkTmp.SourceNode
		targetNodeParam := linkTmp.TargetNode
		// 找到节点对应的类型
		sourceAbstractNode, targetAbstractNode, err := rpt.getSourceNodeAndTargetNode(sourceNodeParam, targetNodeParam)
		if err != nil {
			return fmt.Errorf("get source node and target node failed, %s", err)
		}
		// 进行相应的转换
		sourceNormalNode, err := sourceAbstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		targetNormalNode, err := targetAbstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node failed, %s", err)
		}
		// 获取类型
		sourceNodeType := sourceAbstractNode.Type
		targetNodeType := targetAbstractNode.Type
		currentLinkNums := len(rpt.Links)
		linkId := currentLinkNums + 1
		var linkType types.NetworkLinkType
		var bandWidth int
		if linkTmp.LinkType == "access" {
			linkType = types.NetworkLinkType_AccessLink
			bandWidth = 10 * 1e6
		} else {
			linkType = types.NetworkLinkType_BackboneLink
			bandWidth = linux_tc_api.LargeBandwidth // 没有限制
		}
		ipv4SubNet := rpt.Ipv4SubNets[currentLinkNums]                                                                   // 获取当前ipv4 子网
		ipv6SubNet := rpt.Ipv6SubNets[currentLinkNums]                                                                   // 获取当前 ipv6 子网
		sourceNormalNode.ConnectedIpv4SubnetList = append(sourceNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String()) // 卫星添加ipv4子网
		targetNormalNode.ConnectedIpv4SubnetList = append(targetNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String()) // 卫星添加ipv4子网
		sourceNormalNode.ConnectedIpv6SubnetList = append(sourceNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String()) // 卫星添加ipv6子网
		targetNormalNode.ConnectedIpv6SubnetList = append(targetNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String()) // 卫星添加ipv6子网
		sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                             // 提取ipv4第一个和第二个地址
		sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                             // 提取ipv6第一个和第二个地址
		//sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(sourceNodeType), sourceNormalNode.Id, sourceNormalNode.Ifidx)                                                  // 源接口名
		//targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(targetNodeType), targetNormalNode.Id, targetNormalNode.Ifidx)                                                  // 目的接口名
		sourceIfName := linkTmp.SourceInterfaceName
		targetIfName := linkTmp.TargetInterfaceName
		rpt.NetworkInterfaces += 1                                                                                                                                               // 接口数量 ++
		sourceIntf := intf.NewNetworkInterface(sourceNormalNode.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, rpt.NetworkInterfaces, nil) // 创建第一个接口
		rpt.NetworkInterfaces += 1                                                                                                                                               // 接口数量 ++
		targetIntf := intf.NewNetworkInterface(targetNormalNode.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, rpt.NetworkInterfaces, nil) // 创建第二个接口
		sourceNormalNode.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                         // 设置源节点地址
		targetNormalNode.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                         // 设置目的节点地址
		sourceNormalNode.Interfaces = append(sourceNormalNode.Interfaces, sourceIntf)                                                                                            // 源接口
		targetNormalNode.Interfaces = append(targetNormalNode.Interfaces, targetIntf)                                                                                            // 目的接口
		abstractLink := link.NewAbstractLink(linkType,
			linkId,
			sourceNodeType, targetNodeType,
			sourceNormalNode.Id, targetNormalNode.Id,
			sourceNormalNode.ContainerName, targetNormalNode.ContainerName,
			sourceIntf, targetIntf,
			sourceAbstractNode, targetAbstractNode,
			bandWidth,
			rpt.TopologyGraph,
			ipv4SubNet,
			ipv6SubNet)
		sourceNormalNode.Ifidx++
		targetNormalNode.Ifidx++
		rpt.Links = append(rpt.Links, abstractLink)
		if _, ok := rpt.AllLinksMap[sourceNormalNode.ContainerName]; !ok {
			rpt.AllLinksMap[sourceNormalNode.ContainerName] = make(map[string]*link.AbstractLink)
		}
		rpt.AllLinksMap[sourceNormalNode.ContainerName][targetNormalNode.ContainerName] = abstractLink
	}
	// -------------------------------------------

	rpt.topologyInitSteps[GenerateLinks] = struct{}{}
	raspberrypiTopologyLogger.Infof("generate links")
	return nil
}

func (rpt *RaspberrypiTopology) PrintTopology() error {
	if _, ok := rpt.topologyInitSteps[PrintTopology]; ok {
		raspberrypiTopologyLogger.Infof("already generate links")
		return nil
	}

	for _, router := range rpt.Routers {
		fmt.Println(router)
		fmt.Println("------------------------")
		for _, networkInterface := range router.Interfaces {
			fmt.Printf("interface address: %s\n", networkInterface.SourceIpv4Addr)
		}
		fmt.Println("------------------------")
	}

	for index, linkConnection := range rpt.Links {
		fmt.Println(linkConnection)
		fmt.Println("subnet:", rpt.Ipv4SubNets[index].String())
	}

	rpt.topologyInitSteps[PrintTopology] = struct{}{}
	raspberrypiTopologyLogger.Infof("already print topology")
	return nil
}
