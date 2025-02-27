package topology

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"os"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/api/route"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/chainmaker_prepare"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/fabric_prepare"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/file"
	"zhanghefan123/security_topology/modules/utils/network"
	"zhanghefan123/security_topology/services/http/params"
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

const (
	GenerateChainMakerConfig              = "GenerateChainMakerConfig"              // 生成长安链配置
	GenerateFabricConfig                  = "GenerateFabricConfig"                  // 生成 fabric 配置
	GenerateNodes                         = "GenerateNodes"                         // 生成节点
	GenerateSubnets                       = "GenerateSubnets"                       // 创建子网
	GenerateLinks                         = "GenerateISLs"                          // 生成链路
	GenerateFrrConfigurationFiles         = "GenerateFrrConfigurationFiles"         // 生成 frr 配置
	GenerateAddressMapping                = "GenerateAddressMapping"                // 生成容器名 -> 地址的映射
	GeneratePortMapping                   = "GeneratePortMapping"                   // 生成容器名 -> 端口的映射
	CalculateAndWriteSegmentRoutes        = "CalculateAndWriteSegmentRoutes"        // 生成 srv6 路由文件
	CalculateAndWriteLiRRoutes            = "CalculateAndWriteLiRRoutes"            // 生成 path_validation 路由文件
	GenerateIfnameToLinkIdentifierMapping = "GenerateIfnameToLinkIdentifierMapping" // 生成从接口名称到 link identifier 的映射文件
)

// Init 进行初始化
func (t *Topology) Init() error {
	enabledChainMaker := configs.TopConfiguration.ChainMakerConfig.Enabled
	enabledFabric := configs.TopConfiguration.FabricConfig.Enabled

	initSteps := []map[string]InitModule{
		{GenerateNodes: InitModule{true, t.GenerateNodes}},
		{GenerateSubnets: InitModule{true, t.GenerateSubnets}},
		{GenerateLinks: InitModule{true, t.GenerateLinks}},
		{GenerateFrrConfigurationFiles: InitModule{true, t.GenerateFrrConfigurationFiles}},
		{GenerateChainMakerConfig: InitModule{enabledChainMaker, t.GenerateChainMakerConfig}},
		{GenerateFabricConfig: InitModule{enabledFabric, t.GenerateFabricConfig}},
		{GenerateAddressMapping: InitModule{true, t.GenerateAddressMapping}},
		{GeneratePortMapping: InitModule{true, t.GeneratePortMapping}},
		{CalculateAndWriteSegmentRoutes: InitModule{true, t.CalculateAndWriteSegmentRoutes}},
		{CalculateAndWriteLiRRoutes: InitModule{true, t.CalculateAndWriteLiRRoutes}},
		{GenerateIfnameToLinkIdentifierMapping: InitModule{true, t.GenerateIfnameToLinkIdentifierMapping}},
	}
	err := t.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("constellation init failed: %w", err)
	}
	return nil
}

func (t *Topology) initStepsNum(initSteps []map[string]InitModule) int {
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
func (t *Topology) initializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := t.initStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.init {
				if err = initModule.initFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				topologyLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

// GenerateNodes 进行节点的生成
func (t *Topology) GenerateNodes() error {
	if _, ok := t.topologyInitSteps[GenerateNodes]; ok {
		topologyLogger.Infof("already generate satellites")
		return nil // 重复生成没有必要，但是实际上只要返回就行，不是错误
	}

	// 进行所有的节点的遍历
	for _, nodeParam := range t.TopologyParams.Nodes {
		nodeType, err := types.ResolveNodeType(nodeParam.Type)
		if err != nil {
			return fmt.Errorf("resolve node type failed, %s", err)
		}
		switch *nodeType {
		case types.NetworkNodeType_Router: // 进行普通路由节点的添加
			routerTmp := nodes.NewRouter(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.Routers = append(t.Routers, routerTmp)
			// 注意只能唯一创建一次
			abstractRouter := node.NewAbstractNode(types.NetworkNodeType_Router, routerTmp, TopologyInstance.TopologyGraph)
			t.RouterAbstractNodes = append(t.RouterAbstractNodes, abstractRouter)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractRouter)
			t.AbstractNodesMap[routerTmp.ContainerName] = abstractRouter
		case types.NetworkNodeType_NormalNode:
			normalNodeTmp := normal_node.NewNormalNode(types.NetworkNodeType_NormalNode, nodeParam.Index, fmt.Sprintf("%s-%d", nodeType.String(), nodeParam.Index))
			t.NormalNodes = append(t.NormalNodes, normalNodeTmp)
			// 注意只能唯一创建一次
			abstractNormalNode := node.NewAbstractNode(types.NetworkNodeType_NormalNode, normalNodeTmp, TopologyInstance.TopologyGraph)
			t.NormalAbstractNodes = append(t.NormalAbstractNodes, abstractNormalNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNormalNode)
			t.AbstractNodesMap[normalNodeTmp.ContainerName] = abstractNormalNode
		case types.NetworkNodeType_ConsensusNode:
			consensusNodeTmp := nodes.NewConsensusNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.ConsensusNodes = append(t.ConsensusNodes, consensusNodeTmp)
			// 注意只能唯一创建一次
			abstractConsensusNode := node.NewAbstractNode(types.NetworkNodeType_ConsensusNode, consensusNodeTmp, TopologyInstance.TopologyGraph)
			t.ConsensusAbstractNodes = append(t.ConsensusAbstractNodes, abstractConsensusNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractConsensusNode)
			t.AbstractNodesMap[consensusNodeTmp.ContainerName] = abstractConsensusNode
		case types.NetworkNodeType_ChainMakerNode:
			chainmakerNodeTmp := nodes.NewChainmakerNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.ChainmakerNodes = append(t.ChainmakerNodes, chainmakerNodeTmp)
			// 注意只能唯一创建一次
			abstractChainmakerNode := node.NewAbstractNode(types.NetworkNodeType_ChainMakerNode, chainmakerNodeTmp, TopologyInstance.TopologyGraph)
			t.ChainMakerAbstractNodes = append(t.ChainMakerAbstractNodes, abstractChainmakerNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractChainmakerNode)
			t.AbstractNodesMap[chainmakerNodeTmp.ContainerName] = abstractChainmakerNode
		case types.NetworkNodeType_MaliciousNode:
			maliciousNodeTmp := nodes.NewMaliciousNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.MaliciousNodes = append(t.MaliciousNodes, maliciousNodeTmp)
			// 注意只能唯一创建一次
			abstractMaliciousNode := node.NewAbstractNode(types.NetworkNodeType_MaliciousNode, maliciousNodeTmp, TopologyInstance.TopologyGraph)
			t.MaliciousAbstractNodes = append(t.MaliciousAbstractNodes, abstractMaliciousNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractMaliciousNode)
			t.AbstractNodesMap[maliciousNodeTmp.ContainerName] = abstractMaliciousNode
		case types.NetworkNodeType_LirNode:
			lirNodeTmp := nodes.NewLiRNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.LirNodes = append(t.LirNodes, lirNodeTmp)
			// 注意只能唯一创建一次
			abstractLirNode := node.NewAbstractNode(types.NetworkNodeType_LirNode, lirNodeTmp, TopologyInstance.TopologyGraph)
			t.LirAbstractNodes = append(t.LirAbstractNodes, abstractLirNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractLirNode)
			t.AbstractNodesMap[lirNodeTmp.ContainerName] = abstractLirNode
		case types.NetworkNodeType_Entrance:
			entranceTmp := nodes.NewEntrance(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.Entrances = append(t.Entrances, entranceTmp)
			// 注意只能通过实际节点创建抽象节点一次
			abstractEntrance := node.NewAbstractNode(types.NetworkNodeType_Entrance, entranceTmp, TopologyInstance.TopologyGraph)
			t.EntranceAbstractNodes = append(t.EntranceAbstractNodes, abstractEntrance)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractEntrance)
			t.AbstractNodesMap[entranceTmp.ContainerName] = abstractEntrance
		case types.NetworkNodeType_FabricPeerNode:
			fmt.Println("create fabric peer node")
			fabricPeerTmp := nodes.NewFabricPeerNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.FabricPeerNodes = append(t.FabricPeerNodes, fabricPeerTmp)
			// 注意只能进行一次抽象节点的创建
			abstractFabricPeer := node.NewAbstractNode(types.NetworkNodeType_FabricPeerNode, fabricPeerTmp, TopologyInstance.TopologyGraph)
			t.FabricPeerAbstractNodes = append(t.FabricPeerAbstractNodes, abstractFabricPeer)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractFabricPeer)
			t.AbstractNodesMap[fabricPeerTmp.ContainerName] = abstractFabricPeer
		case types.NetworkNodeType_FabricOrderNode:
			fmt.Println("create fabric order node")
			fabricOrderTmp := nodes.NewFabricOrderNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.FabricOrderNodes = append(t.FabricOrderNodes, fabricOrderTmp)
			// 注意只能进行一次抽象节点的创建
			abstractFabricOrder := node.NewAbstractNode(types.NetworkNodeType_FabricOrderNode, fabricOrderTmp, TopologyInstance.TopologyGraph)
			t.FabricOrderAbstractNodes = append(t.FabricOrderAbstractNodes, abstractFabricOrder)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractFabricOrder)
			t.AbstractNodesMap[fabricOrderTmp.ContainerName] = abstractFabricOrder
		}
	}

	t.topologyInitSteps[GenerateNodes] = struct{}{}
	topologyLogger.Infof("generate satellites")

	return nil
}

func (t *Topology) GenerateSubnets() error {
	if _, ok := t.topologyInitSteps[GenerateSubnets]; ok {
		topologyLogger.Infof("already generate subnets")
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
	t.Ipv4SubNets = ipv4Subnets

	// 进行 ipv6 的子网的生成
	ipv6Subnets, err = network.GenerateIpv6Subnets(configs.TopConfiguration.NetworkConfig.BaseV6NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	t.Ipv6SubNets = ipv6Subnets

	t.topologyInitSteps[GenerateSubnets] = struct{}{}
	topologyLogger.Infof("generate subnets")
	return nil
}

func (t *Topology) GenerateLinks() error {
	if _, ok := t.topologyInitSteps[GenerateLinks]; ok {
		topologyLogger.Infof("already generate links")
		return nil
	}

	// ----------------实际逻辑--------------------
	for _, linkTmp := range t.TopologyParams.Links {
		// 拿到从前端传递过来的 (源节点 目的节点的参数)
		sourceNodeParam := linkTmp.SourceNode
		targetNodeParam := linkTmp.TargetNode
		// 找到节点对应的类型
		sourceAbstractNode, targetAbstractNode, err := t.getSourceNodeAndTargetNode(sourceNodeParam, targetNodeParam)
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
		currentLinkNums := len(t.Links)
		linkId := currentLinkNums + 1
		var linkType types.NetworkLinkType
		var bandWidth int
		if linkTmp.LinkType == "access" {
			linkType = types.NetworkLinkType_AccessLink
			bandWidth = t.TopologyParams.AccessLinkBandwidth * 1e6
		} else {
			linkType = types.NetworkLinkType_BackboneLink
			bandWidth = linux_tc_api.LargeBandwidth // 没有限制
		}
		ipv4SubNet := t.Ipv4SubNets[currentLinkNums]                                                                                                                           // 获取当前ipv4 子网
		ipv6SubNet := t.Ipv6SubNets[currentLinkNums]                                                                                                                           // 获取当前 ipv6 子网
		sourceNormalNode.ConnectedIpv4SubnetList = append(sourceNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                       // 卫星添加ipv4子网
		targetNormalNode.ConnectedIpv4SubnetList = append(targetNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                       // 卫星添加ipv4子网
		sourceNormalNode.ConnectedIpv6SubnetList = append(sourceNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                       // 卫星添加ipv6子网
		targetNormalNode.ConnectedIpv6SubnetList = append(targetNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                       // 卫星添加ipv6子网
		sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                                                                   // 提取ipv4第一个和第二个地址
		sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                                                                   // 提取ipv6第一个和第二个地址
		sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(sourceNodeType), sourceNormalNode.Id, sourceNormalNode.Ifidx)                                                // 源接口名
		targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(targetNodeType), targetNormalNode.Id, targetNormalNode.Ifidx)                                                // 目的接口名
		t.NetworkInterfaces += 1                                                                                                                                               // 接口数量 ++
		sourceIntf := intf.NewNetworkInterface(sourceNormalNode.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, t.NetworkInterfaces, nil) // 创建第一个接口
		t.NetworkInterfaces += 1                                                                                                                                               // 接口数量 ++
		targetIntf := intf.NewNetworkInterface(targetNormalNode.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, t.NetworkInterfaces, nil) // 创建第二个接口
		sourceNormalNode.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                       // 设置源节点地址
		targetNormalNode.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                       // 设置目的节点地址
		sourceNormalNode.Interfaces = append(sourceNormalNode.Interfaces, sourceIntf)                                                                                          // 源接口
		targetNormalNode.Interfaces = append(targetNormalNode.Interfaces, targetIntf)                                                                                          // 目的接口
		abstractLink := link.NewAbstractLink(linkType,
			linkId,
			sourceNodeType, targetNodeType,
			sourceNormalNode.Id, targetNormalNode.Id,
			sourceNormalNode.ContainerName, targetNormalNode.ContainerName,
			sourceIntf, targetIntf,
			sourceAbstractNode, targetAbstractNode,
			bandWidth,
			TopologyInstance.TopologyGraph)
		sourceNormalNode.Ifidx++
		targetNormalNode.Ifidx++
		t.Links = append(t.Links, abstractLink)
		if _, ok := t.AllLinksMap[sourceNormalNode.ContainerName]; !ok {
			t.AllLinksMap[sourceNormalNode.ContainerName] = make(map[string]*link.AbstractLink)
		}
		t.AllLinksMap[sourceNormalNode.ContainerName][targetNormalNode.ContainerName] = abstractLink
	}
	// -------------------------------------------

	t.topologyInitSteps[GenerateLinks] = struct{}{}
	topologyLogger.Infof("generate links")
	return nil
}

// getSourceNodeAndTargetNode 获取源和目的抽象节点
func (t *Topology) getSourceNodeAndTargetNode(sourceNodeParam, targetNodeParam params.NodeParam) (*node.AbstractNode, *node.AbstractNode, error) {
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
		sourceNode = t.RouterAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_NormalNode:
		sourceNode = t.NormalAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_ConsensusNode:
		sourceNode = t.ConsensusAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_ChainMakerNode:
		sourceNode = t.ChainMakerAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_MaliciousNode:
		sourceNode = t.MaliciousAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_LirNode:
		sourceNode = t.LirAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_Entrance:
		sourceNode = t.EntranceAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_FabricPeerNode:
		sourceNode = t.FabricPeerAbstractNodes[sourceNodeParam.Index-1]
	case types.NetworkNodeType_FabricOrderNode:
		sourceNode = t.FabricOrderAbstractNodes[sourceNodeParam.Index-1]
	default:
		return nil, nil, fmt.Errorf("unsupported source node type: %s", *sourceNodeType)
	}

	// 拿到目的节点
	switch *targetNodeType {
	case types.NetworkNodeType_Router:
		targetNode = t.RouterAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_NormalNode:
		targetNode = t.NormalAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_ConsensusNode:
		targetNode = t.ConsensusAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_ChainMakerNode:
		targetNode = t.ChainMakerAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_MaliciousNode:
		targetNode = t.MaliciousAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_LirNode:
		targetNode = t.LirAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_Entrance:
		targetNode = t.EntranceAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_FabricPeerNode:
		targetNode = t.FabricPeerAbstractNodes[targetNodeParam.Index-1]
	case types.NetworkNodeType_FabricOrderNode:
		targetNode = t.FabricOrderAbstractNodes[targetNodeParam.Index-1]
	default:
		return nil, nil, fmt.Errorf("unsupported target node type: %s", *sourceNodeType)
	}

	return sourceNode, targetNode, nil
}

// GenerateFrrConfigurationFiles 进行 frr 配置文件的生成
func (t *Topology) GenerateFrrConfigurationFiles() error {
	if _, ok := t.topologyInitSteps[GenerateFrrConfigurationFiles]; ok {
		topologyLogger.Infof("already generate frr configuration files")
		return nil
	}

	selectedOspfVersion := configs.TopConfiguration.NetworkConfig.OspfVersion

	for index, abstractNode := range t.AllAbstractNodes {
		routerId := index + 1
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("unsupported ")
		}
		if selectedOspfVersion == "ospfv2" {
			err = normalNode.GenerateOspfV2FrrConfig(routerId)
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed, %s", err)
			}
		} else if selectedOspfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err = normalNode.GenerateOspfV3FrrConfig(routerId)
			if err != nil {
				return fmt.Errorf("generate ospfv3 frr configuration files failed, %s", err)
			}
		} else {
			return fmt.Errorf("unsupported ospf version: %s", selectedOspfVersion)
		}
	}

	t.topologyInitSteps[GenerateFrrConfigurationFiles] = struct{}{}
	topologyLogger.Infof("generate frr configuration files")
	return nil
}

// GenerateChainMakerConfig 进行 ChainMaker 配置文件的生成
func (t *Topology) GenerateChainMakerConfig() error {
	if _, ok := t.topologyInitSteps[GenerateChainMakerConfig]; ok {
		topologyLogger.Infof("already generate chain maker config")
		return nil
	}

	chainMakerNodeCount := len(t.ChainmakerNodes)
	if chainMakerNodeCount == 0 {
		topologyLogger.Infof("no chainmaker nodes -> not generate")
		return nil
	}

	ipv4Addresses := t.GetChainMakerNodeListenAddresses()
	chainMakerPrepare := chainmaker_prepare.NewChainMakerPrepare(chainMakerNodeCount, ipv4Addresses)
	err := chainMakerPrepare.Generate()
	if err != nil {
		return fmt.Errorf("generate chain maker config files failed, %s", err)
	}

	t.topologyInitSteps[GenerateChainMakerConfig] = struct{}{}
	topologyLogger.Infof("generate chain maker config")
	return nil
}

// GenerateFabricConfig 进行 fabric 配置文件的生成
func (t *Topology) GenerateFabricConfig() error {
	if _, ok := t.topologyInitSteps[GenerateFabricConfig]; ok {
		topologyLogger.Infof("already generate fabric config")
		return nil
	}

	fabricPeerNodesCount := len(t.FabricPeerNodes)
	fabricOrderNodesCount := len(t.FabricOrderNodes)
	if (fabricPeerNodesCount == 0) && (fabricOrderNodesCount == 0) {
		topologyLogger.Infof("no fabric nodes -> not generate")
		return nil
	}

	fabricPrepare := fabric_prepare.NewFabricPrepare(fabricPeerNodesCount, fabricOrderNodesCount)
	err := fabricPrepare.Generate()
	if err != nil {
		return fmt.Errorf("generate fabric config files failed, %s", err)
	}

	t.topologyInitSteps[GenerateFabricConfig] = struct{}{}
	topologyLogger.Infof("generate fabric config")
	return nil
}

// GenerateAddressMapping 生成地址映射
func (t *Topology) GenerateAddressMapping() error {
	if _, ok := t.topologyInitSteps[GenerateAddressMapping]; ok {
		topologyLogger.Infof("already generate address mapping")
		return nil
	}

	addressMapping, err := t.GetContainerNameToAddressMapping()
	if err != nil {
		return fmt.Errorf("generate address mapping files failed, %w", err)
	}

	idMapping, err := t.GetContainerNameToGraphIdMapping()
	if err != nil {
		return fmt.Errorf("generate id mapping files failed, %w", err)
	}

	finalString := ""
	for containerName, ip := range addressMapping {
		graphId := idMapping[containerName]
		finalString += fmt.Sprintf("%s->%d->%s\n", containerName, graphId, ip)
	}

	// 进行所有节点的遍历
	for _, abstractNode := range t.AllAbstractNodes {
		var f *os.File
		var normalNode *normal_node.NormalNode
		var outputDir string
		var outputFilePath string
		normalNode, err = abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("generate address mapping files failed, %s", err)
		}
		containerName := normalNode.ContainerName
		outputDir = filepath.Join(configs.TopConfiguration.PathConfig.ConfigGeneratePath, containerName)
		outputFilePath = filepath.Join(outputDir, "address_mapping.conf")
		// 创建一个文件
		// /simulation/containerName/address_mapping.conf
		f, err = os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("error create file %v", err)
		}
		_, err = f.WriteString(finalString)
		if err != nil {
			return fmt.Errorf("error write file %w", err)
		}
	}

	t.topologyInitSteps[GenerateAddressMapping] = struct{}{}
	topologyLogger.Infof("generate address mapping")
	return nil
}

func (t *Topology) GeneratePortMapping() error {
	if _, ok := t.topologyInitSteps[GeneratePortMapping]; ok {
		topologyLogger.Infof("already generate port mapping")
		return nil
	}

	portMapping, err := t.GetContainerNameToPortMapping()
	if err != nil {
		return fmt.Errorf("generate port mapping files failed, %w", err)
	}

	finalString := ""
	for key, value := range portMapping {
		finalString += fmt.Sprintf("%s->%d->%d\n", key, value.p2pPort, value.rpcPort)
	}

	// 进行所有节点的遍历
	for _, abstractNode := range t.AllAbstractNodes {
		var f *os.File
		var normalNode *normal_node.NormalNode
		var outputDir string
		var outputFilePath string
		normalNode, err = abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("generate address mapping files failed, %s", err)
		}
		containerName := normalNode.ContainerName
		outputDir = filepath.Join(configs.TopConfiguration.PathConfig.ConfigGeneratePath, containerName)
		outputFilePath = filepath.Join(outputDir, "port_mapping.conf")
		// 创建一个文件
		// /simulation/containerName/address_mapping.conf
		f, err = os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("error create file %v", err)
		}
		_, err = f.WriteString(finalString)
		if err != nil {
			return fmt.Errorf("error write file %w", err)
		}
	}

	t.topologyInitSteps[GeneratePortMapping] = struct{}{}
	topologyLogger.Infof("generate port mapping")
	return nil
}

// CalculateAndWriteSegmentRoutes 进行段路由的计算
func (t *Topology) CalculateAndWriteSegmentRoutes() error {
	if _, ok := t.topologyInitSteps[CalculateAndWriteSegmentRoutes]; ok {
		topologyLogger.Infof("already calculate segment routes")
		return nil
	}

	for _, abstractNode := range t.AllAbstractNodes {
		err := route.CalculateAndWriteSegmentRoute(abstractNode, &(t.AllLinksMap), TopologyInstance.TopologyGraph)
		if err != nil {
			return fmt.Errorf("calculate route failed: %w", err)
		}
	}

	t.topologyInitSteps[CalculateAndWriteSegmentRoutes] = struct{}{}
	topologyLogger.Infof("calculate segment routes")
	return nil
}

// CalculateAndWriteLiRRoutes 进行 LiR 路由的计算
func (t *Topology) CalculateAndWriteLiRRoutes() error {
	if _, ok := t.topologyInitSteps[CalculateAndWriteLiRRoutes]; ok {
		topologyLogger.Infof("already calculate path_validation routes")
		return nil
	}

	// simulation 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath

	// 所有的节点的 LiR 路由
	allLiRRoutes := make([]string, 0)

	// 遍历所有节点生成路由文件
	for _, abstractNode := range t.AllAbstractNodes {
		// 获取单个节点的路由条目集合
		lirRoute, err := route.GenerateLiRRoute(abstractNode, &(t.AllLinksMap), TopologyInstance.TopologyGraph)
		if err != nil {
			return fmt.Errorf("generate path_validation route failed: %w", err)
		}
		// 更新总路由条目
		allLiRRoutes = append(allLiRRoutes, lirRoute)
	}

	allLirRoutesString := strings.Join(allLiRRoutes, "\n")

	// 准备写入文件
	for index, abstractNode := range t.AllAbstractNodes {
		// 提取普通节点
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("cannot get normal node from abstract node, %w", err)
		}
		// route 文件夹的位置
		routeDir := filepath.Join(simulationDir, normalNode.ContainerName, "route")
		// 进行文件夹的生成
		err = dir.Generate(routeDir)
		if err != nil {
			return fmt.Errorf("cannot generate route directory, %w", err)
		}
		// 单节点路由文件的路径
		lirRouteFilePath := filepath.Join(routeDir, "lir.txt")
		// 所有节点路由文件的路径
		allLiRRouteFilePath := filepath.Join(routeDir, "all_lir.txt")
		// 写入文件
		err = file.WriteStringIntoFile(lirRouteFilePath, allLiRRoutes[index])
		if err != nil {
			return fmt.Errorf("error writing path_validation route file, %w", err)
		}
		err = file.WriteStringIntoFile(allLiRRouteFilePath, allLirRoutesString)
	}

	t.topologyInitSteps[CalculateAndWriteLiRRoutes] = struct{}{}
	topologyLogger.Infof("calculate path_validation routes")
	return nil
}

// GenerateIfnameToLinkIdentifierMapping 生成从接口名到链路标识的映射
func (t *Topology) GenerateIfnameToLinkIdentifierMapping() error {
	if _, ok := t.topologyInitSteps[GenerateIfnameToLinkIdentifierMapping]; ok {
		topologyLogger.Infof("already generate ifname to link identifier")
		return nil
	}

	// 遍历所有的节点生成 mapping
	for _, abstractNode := range t.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("generate ifname to link identifier mapping files failed, %w", err)
		}
		err = normalNode.GenerateIfnameToLidMapping()
		if err != nil {
			return fmt.Errorf("generate ifname to link identifier mapping files failed, %w", err)
		}
	}

	t.topologyInitSteps[GenerateIfnameToLinkIdentifierMapping] = struct{}{}
	topologyLogger.Infof("generate ifname to link identifier")
	return nil
}
