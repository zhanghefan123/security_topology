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
	"zhanghefan123/security_topology/modules/chain_prepares/chainmaker_prepare"
	"zhanghefan123/security_topology/modules/chain_prepares/fabric_prepare"
	"zhanghefan123/security_topology/modules/chain_prepares/fisco_bcos_prepare"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/services/http/params"
	"zhanghefan123/security_topology/utils/dir"
	"zhanghefan123/security_topology/utils/extract"
	"zhanghefan123/security_topology/utils/file"
	"zhanghefan123/security_topology/utils/judge"
	"zhanghefan123/security_topology/utils/network"
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

const (
	GenerateChainMakerConfig                            = "GenerateChainMakerConfig"                            // 生成长安链配置
	GenerateFabricConfig                                = "GenerateFabricConfig"                                // 生成 fabric 配置
	GenerateFiscoBcosConfig                             = "GenerateFiscoBcosConfig"                             // 生成 fisco bcos 配置
	GenerateNodes                                       = "GenerateNodes"                                       // 生成节点
	GenerateSubnets                                     = "GenerateSubnets"                                     // 创建子网
	GenerateLinks                                       = "GenerateISLs"                                        // 生成链路
	GenerateFrrConfigurationFiles                       = "GenerateFrrConfigurationFiles"                       // 生成 frr 配置
	GenerateAddressMapping                              = "GenerateAddressMapping"                              // 生成容器名 -> 地址的映射
	GeneratePortMapping                                 = "GeneratePortMapping"                                 // 生成容器名 -> 端口的映射
	CalculateAndWriteSegmentRoutes                      = "CalculateAndWriteSegmentRoutes"                      // 生成 srv6 路由文件
	CalculateAndWriteLiRRoutes                          = "CalculateAndWriteLiRRoutes"                          // 生成 path_validation 路由文件
	GenerateIfnameToLinkIdentifierMapping               = "GenerateIfnameToLinkIdentifierMapping"               // 生成从接口名称到 link identifier 的映射文件
	GenerateFabricNodeIDtoAddressMapping                = "GenerateFabricNodeIDtoAddressMapping"                // 生成从 fabric 节点 id 到对应的 ip 地址的映射文件
	GenerateChainMakerIDtoNameMapping                   = "GenerateChainMakerIDtoNameMapping"                   // 生成从 id 到 name 的映射
	GenerateAtlasSegmentsAndOutputLinkIdentifiers       = "GenerateAtlasSegmentsAndOutputLinkIdentifiers"       // 生成 atlas 多路径的分段
	GenerateMultipathSelirPathsAndOutputLinkIdentifiers = "GenerateMultipathSelirPathsAndOutputLinkIdentifiers" // 生成 multipath_selir 多路径
	GenerateTopologyForSecPathMab                       = "GenerateTopologyForSecPathMab"                       // 生成 sec_path_mab 的拓扑配置
)

// Init 进行初始化
func (t *Topology) Init() error {

	// 1. GenerateNodes 需要单独执行是因为其会更新 t.ChainMakerEnabled / t.FabricEnabled / t.FiscoBcosEnabled
	err := t.GenerateNodes()
	if err != nil {
		return fmt.Errorf("generate nodes failed")
	}

	var enableUnicast = false
	var enableAtlas = false
	var enableMultipathSelir = false
	if configs.TopConfiguration.PathValidationConfig.TransmissionType == (int)(types.TransmissionType_MULTIPATH) {
		if configs.TopConfiguration.PathValidationConfig.MultipathConfig.MultipathRoutingType == (int)(types.RoutingType_Atlas) {
			enableAtlas = true
		} else if configs.TopConfiguration.PathValidationConfig.MultipathConfig.MultipathRoutingType == (int)(types.RoutingType_MultipathSelir) {
			enableMultipathSelir = true
		} else {
			fmt.Printf("unsupported specific multipath protocol\n")
		}
	} else {
		enableUnicast = true
	}

	//var enableSecPathMab = false
	//if configs.TopConfiguration.PathValidationConfig.TransmissionType == (int)(types.TransmissionType_MAB) {
	//	enableSecPathMab = true
	//}

	initSteps := []map[string]InitModule{
		{GenerateSubnets: InitModule{true, t.GenerateSubnets}},
		{GenerateLinks: InitModule{true, t.GenerateLinks}},
		{GenerateFrrConfigurationFiles: InitModule{true, t.GenerateFrrConfigurationFiles}},
		{GenerateChainMakerConfig: InitModule{t.ChainMakerEnabled, t.GenerateChainMakerConfig}},
		{GenerateFabricConfig: InitModule{t.FabricEnabled, t.GenerateFabricConfig}},
		{GenerateFabricNodeIDtoAddressMapping: InitModule{t.FabricEnabled, t.GenerateFabricNodeIDtoAddressMapping}},
		{GenerateFiscoBcosConfig: InitModule{t.FiscoBcosEnabled, t.GenerateFiscoBcosConfig}},
		{GenerateAddressMapping: InitModule{true, t.GenerateAddressMapping}},
		{GeneratePortMapping: InitModule{true, t.GeneratePortMapping}},
		{CalculateAndWriteSegmentRoutes: InitModule{true, t.CalculateAndWriteSegmentRoutes}},
		{GenerateIfnameToLinkIdentifierMapping: InitModule{true, t.GenerateIfnameToLinkIdentifierMapping}},
		{GenerateChainMakerIDtoNameMapping: InitModule{t.ChainMakerEnabled, t.GenerateChainMakerIDtoNameMapping}},
		{CalculateAndWriteLiRRoutes: InitModule{enableUnicast, t.CalculateAndWriteLiRRoutes}},
		{GenerateAtlasSegmentsAndOutputLinkIdentifiers: InitModule{enableAtlas, t.GenerateAtlasSegmentsAndOutputLinkIdentifiers}},
		{GenerateMultipathSelirPathsAndOutputLinkIdentifiers: InitModule{enableMultipathSelir, t.GenerateMultipathSelirPathsAndOutputLinkIdentifiers}},
		//{GenerateTopologyForSecPathMab: InitModule{enableSecPathMab, t.GenerateTopologyForSecPathMab}},
	}

	err = t.initializeSteps(initSteps)
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
		var abstractNode *node.AbstractNode
		nodeType, err := types.ResolveNodeType(nodeParam.Type)
		if err != nil {
			return fmt.Errorf("resolve node type failed, %s", err)
		}
		switch *nodeType {
		case types.NetworkNodeType_Router: // 进行普通路由节点的添加
			routerTmp := nodes.NewRouter(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.Routers = append(t.Routers, routerTmp)
			// 注意只能唯一创建一次
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_Router, routerTmp, Instance.TopologyGraph)
			t.RouterAbstractNodes = append(t.RouterAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[routerTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_ChainMakerNode:
			chainmakerTmp := nodes.NewChainmakerNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.ChainmakerNodes = append(t.ChainmakerNodes, chainmakerTmp)
			// 注意只能唯一创建一次
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_ChainMakerNode, chainmakerTmp, Instance.TopologyGraph)
			t.ChainMakerAbstractNodes = append(t.ChainMakerAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[chainmakerTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_LirNode:
			lirNodeTmp := nodes.NewLiRNode(nodeParam.Index, nodeParam.X, nodeParam.Y, nodeParam.SpecialParams)
			t.LirNodes = append(t.LirNodes, lirNodeTmp)
			// 注意只能唯一创建一次
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_LirNode, lirNodeTmp, Instance.TopologyGraph)
			t.LirAbstractNodes = append(t.LirAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[lirNodeTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_Entrance:
			entranceTmp := nodes.NewEntrance(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.Entrances = append(t.Entrances, entranceTmp)
			// 注意只能通过实际节点创建抽象节点一次
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_Entrance, entranceTmp, Instance.TopologyGraph)
			t.EntranceAbstractNodes = append(t.EntranceAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[entranceTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_FabricPeerNode:
			fmt.Println("create fabric peer node")
			fabricPeerTmp := nodes.NewFabricPeerNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.FabricPeerNodes = append(t.FabricPeerNodes, fabricPeerTmp)
			// 注意只能进行一次抽象节点的创建
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_FabricPeerNode, fabricPeerTmp, Instance.TopologyGraph)
			t.FabricPeerAbstractNodes = append(t.FabricPeerAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[fabricPeerTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_FabricOrderNode:
			fmt.Println("create fabric order node")
			fabricOrderTmp := nodes.NewFabricOrderNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.FabricOrdererNodes = append(t.FabricOrdererNodes, fabricOrderTmp)
			// 注意只能进行一次抽象节点的创建
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_FabricOrderNode, fabricOrderTmp, Instance.TopologyGraph)
			t.FabricOrderAbstractNodes = append(t.FabricOrderAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[fabricOrderTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_MaliciousNode:
			fmt.Println("create malicious node")
			maliciousTmp := nodes.NewMaliciousNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.MaliciousNodes = append(t.MaliciousNodes, maliciousTmp)
			// 注意只能进行一次抽象节点的创建
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_MaliciousNode, maliciousTmp, Instance.TopologyGraph)
			t.MaliciousAbstractNodes = append(t.MaliciousAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[maliciousTmp.ContainerName] = abstractNode
		case types.NetworkNodeType_FiscoBcosNode:
			fmt.Println("create fisco bcos node")
			fiscoBcosNodeTmp := nodes.NewFiscoBcosNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.FiscoBcosNodes = append(t.FiscoBcosNodes, fiscoBcosNodeTmp)
			// 注意只能进行一次抽象节点的创建
			abstractNode = node.NewAbstractNode(types.NetworkNodeType_FiscoBcosNode, fiscoBcosNodeTmp, Instance.TopologyGraph)
			t.FiscoBcosAbstractNodes = append(t.FiscoBcosAbstractNodes, abstractNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNode)
			t.AbstractNodesMap[fiscoBcosNodeTmp.ContainerName] = abstractNode
		}
		if judge.IsBlockChainType(*nodeType) {
			t.AllChainAbstractNodes = append(t.AllChainAbstractNodes, abstractNode)
		}
	}

	// 判断哪些链的配置应该被激活
	t.ChainMakerEnabled = len(t.ChainmakerNodes) > 0
	t.FabricEnabled = (len(t.FabricOrdererNodes) > 0) && (len(t.FabricPeerNodes) > 0)
	t.FiscoBcosEnabled = len(t.FiscoBcosNodes) > 0

	t.topologyInitSteps[GenerateNodes] = struct{}{}
	topologyLogger.Infof("generate nodes")

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
			//linkType = types.NetworkLinkType_AccessLink
			////bandWidth = tt.TopologyParams.AccessLinkBandwidth * 1e6
			//bandWidth = 50 * 1e6
			linkType = types.NetworkLinkType_AccessLink
			bandWidth = 20 * 1e6 // 没有限制
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
			Instance.TopologyGraph,
			ipv4SubNet,
			ipv6SubNet)
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
	case types.NetworkNodeType_FiscoBcosNode:
		sourceNode = t.FiscoBcosAbstractNodes[sourceNodeParam.Index-1]
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
	case types.NetworkNodeType_FiscoBcosNode:
		targetNode = t.FiscoBcosAbstractNodes[targetNodeParam.Index-1]
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
	t.chainMakerPrepare = chainmaker_prepare.NewChainMakerPrepare(chainMakerNodeCount, ipv4Addresses, t.TopologyParams.ConsensusType)
	err := t.chainMakerPrepare.Generate()
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

	// 进行对等节点数量的获取
	fabricPeerNodesCount := len(t.FabricPeerNodes)
	// 进行排序节点数量的获取
	fabricOrdererNodesCount := len(t.FabricOrdererNodes)

	// 如果都为0 直接返回
	if (fabricPeerNodesCount == 0) && (fabricOrdererNodesCount == 0) {
		topologyLogger.Infof("no fabric nodes -> not generate")
		return nil
	}

	// 创建 fabric prepare 准备对象
	fabricPrepare := fabric_prepare.NewFabricPrepare(fabricPeerNodesCount, fabricOrdererNodesCount)

	// 进行配置文件的生成
	err := fabricPrepare.Generate()
	if err != nil {
		return fmt.Errorf("generate fabric config files failed, %s", err)
	}

	t.topologyInitSteps[GenerateFabricConfig] = struct{}{}
	topologyLogger.Infof("generate fabric config")
	return nil
}

func (t *Topology) GenerateFiscoBcosConfig() error {
	if _, ok := t.topologyInitSteps[GenerateFiscoBcosConfig]; ok {
		topologyLogger.Infof("already generate fisco bcos config")
		return nil
	}

	// 进行总的 fisco bcos 节点数量的获取
	fiscoBcosNodesCount := len(t.FiscoBcosNodes)

	// 如果为空, 直接进行返回
	if fiscoBcosNodesCount == 0 {
		topologyLogger.Infof("no fisco bcos nodes -> not generate")
		return nil
	}

	// 获取 fisco bcos 的 ip address
	ipAddresses := make([]string, 0)
	for _, fiscoBcosNode := range t.FiscoBcosNodes {
		ipAddresses = append(ipAddresses, fiscoBcosNode.Interfaces[0].SourceIpv4Addr)
	}

	// 获取 p2p start port
	p2pStartPort := configs.TopConfiguration.FiscoBcosConfig.P2pStartPort

	// 创建 fisco bcos prepare
	fiscoBcosPrepare := fisco_bcos_prepare.NewFiscoBcosPrepare(fiscoBcosNodesCount, ipAddresses, p2pStartPort)
	err := fiscoBcosPrepare.Generate()
	if err != nil {
		return fmt.Errorf("generate fisco bcos config files failed, %s", err)
	}

	t.topologyInitSteps[GenerateFiscoBcosConfig] = struct{}{}
	topologyLogger.Infof("generate fisco bcos config")
	return nil
}

func (t *Topology) GenerateFabricNodeIDtoAddressMapping() error {
	if _, ok := t.topologyInitSteps[GenerateFabricNodeIDtoAddressMapping]; ok {
		topologyLogger.Infof("already generate fabricNodeIDtoAddressMapping")
		return nil
	}

	finalString := ""
	for index, fabricOrdererNode := range t.FabricOrdererNodes {
		ordererDomainName := fmt.Sprintf("orderer%d.example.com", fabricOrdererNode.Id)
		ordererGeneralListenPort := configs.TopConfiguration.FabricConfig.OrderGeneralListenStartPort + fabricOrdererNode.Id
		if index != (len(t.FabricOrdererNodes) - 1) {
			finalString += fmt.Sprintf("%d,%s:%d\n", fabricOrdererNode.Id, ordererDomainName, ordererGeneralListenPort)
		} else {
			finalString += fmt.Sprintf("%d,%s:%d", fabricOrdererNode.Id, ordererDomainName, ordererGeneralListenPort)
		}
	}

	for _, fabricOrdererNode := range t.FabricOrdererNodes {
		var outputFilePath string
		containerName := fabricOrdererNode.ContainerName
		containerDir := filepath.Join(configs.TopConfiguration.PathConfig.ConfigGeneratePath, containerName)
		fabricDir := filepath.Join(containerDir, "fabric")

		err := dir.Generate(fabricDir)
		if err != nil {
			return fmt.Errorf("generate fabric dir failed")
		}
		outputFilePath = filepath.Join(fabricDir, "fabric_id_to_address.conf")
		// 创建一个文件
		// /simulation/containerName/fabric_id_to_address.conf
		f, err := os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("error create file %v", err)
		}
		_, err = f.WriteString(finalString)
		if err != nil {
			return fmt.Errorf("error write file %w", err)
		}
	}

	t.topologyInitSteps[GenerateFabricNodeIDtoAddressMapping] = struct{}{}
	topologyLogger.Infof("generate fabricNodeID to address mapping")
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
		return fmt.Errorf("get container name to address mapping failed, %w", err)
	}

	idMapping, err := t.GetContainerNameToGraphIdMapping()
	if err != nil {
		return fmt.Errorf("generate id mapping files failed, %w", err)
	}

	finalString := ""
	for containerName, ipv4andipv6 := range addressMapping {
		graphId := idMapping[containerName]
		finalString += fmt.Sprintf("%s->%d->%s->%s\n", containerName, graphId, ipv4andipv6[0], ipv4andipv6[1])
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
		outputFilePath = filepath.Join(outputDir, "route/address_mapping.conf")
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

	chainMakerPortMapping, err := t.GetContainerNameToChainMakerPortMapping()
	if err != nil {
		return fmt.Errorf("generate port mapping files failed: %w", err)
	}

	fabricPortMapping, err := t.GetContainerNameToFabricPortMapping()
	if err != nil {
		return fmt.Errorf("generate fabric port mapping files failed: %w", err)
	}

	fiscoBcosPortMapping, err := t.GetContainerNameToFiscoBcosPortMapping()
	if err != nil {
		return fmt.Errorf("generate fisco bcos port mapping files failed: %w", err)
	}

	finalString := ""
	for key, value := range chainMakerPortMapping {
		finalString += fmt.Sprintf("%s->%d->%d\n", key, value.p2pPort, value.rpcPort)
	}
	for key, value := range fabricPortMapping {
		finalString += fmt.Sprintf("%s->%d->%d\n", key, value.OrdererAdminListenPort, value.OrdererGeneralListenPort)
	}
	for key, value := range fiscoBcosPortMapping {
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
		err := route.CalculateAndWriteSegmentRoute(abstractNode, &(t.AllLinksMap), Instance.TopologyGraph)
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
		lirRoute, err := route.GenerateLiRRoute(abstractNode, &(t.AllLinksMap), Instance.TopologyGraph)
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
		// 写入每个节点的 lir.txt 文件
		if configs.TopConfiguration.PathValidationConfig.TransmissionType == (int)(types.TransmissionType_UNICAST) {
			currentNodeString := allLiRRoutes[index]
			err = file.WriteStringIntoFile(lirRouteFilePath, currentNodeString)
			if err != nil {
				return fmt.Errorf("error writing lir.txt file, %w", err)
			}
		}
		// 多播的时候只在源节点进行写入
		if (normalNode.Id == 1) && (configs.TopConfiguration.PathValidationConfig.TransmissionType == (int)(types.TransmissionType_MULTICAST)) {
			err = file.WriteStringIntoFile(allLiRRouteFilePath, allLirRoutesString)
			if err != nil {
				return fmt.Errorf("error writing all_lir.txt")
			}
		}
	}

	t.topologyInitSteps[CalculateAndWriteLiRRoutes] = struct{}{}
	topologyLogger.Infof("calculate path_validation routes")
	return nil
}

// GenerateIfnameToLinkIdentifierMapping 生成从接口名到链路标识的映射
func (t *Topology) GenerateIfnameToLinkIdentifierMapping() (err error) {
	if _, ok := t.topologyInitSteps[GenerateIfnameToLinkIdentifierMapping]; ok {
		topologyLogger.Infof("already generate ifname to link identifier")
		return nil
	}

	finalMapping := map[string]string{}
	allInterfaceList := ""

	// 进行所有的链路的遍历
	for _, abstractLink := range t.Links {
		// 提取源节点信息
		sourceNode := abstractLink.SourceNode
		sourceIntf := abstractLink.SourceInterface
		var sourceNormalNode *normal_node.NormalNode
		sourceNormalNode, err = sourceNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node failed: %v", err)
		}
		sourceNodeIdx := sourceNormalNode.Id

		// 提取目的节点信息
		targetNode := abstractLink.TargetNode
		targetIntf := abstractLink.TargetInterface
		var targetNormalNode *normal_node.NormalNode
		targetNormalNode, err = targetNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node failed: %v", err)
		}
		targetNodeIdx := targetNormalNode.Id

		// 更新源接口
		sourceDesc := fmt.Sprintf("%s->%d->%d->%d->%s\n", sourceIntf.IfName, sourceIntf.LinkIdentifier, sourceNodeIdx, targetNodeIdx, sourceIntf.TargetIpv4Addr)
		if result, ok := finalMapping[sourceNormalNode.ContainerName]; !ok {
			result += sourceDesc
			finalMapping[sourceNormalNode.ContainerName] = result
		} else {
			result = sourceDesc
			finalMapping[sourceNormalNode.ContainerName] += result
		}
		allInterfaceList = allInterfaceList + sourceDesc

		// 更新目的接口
		targetDesc := fmt.Sprintf("%s->%d->%d->%d->%s\n", targetIntf.IfName, targetIntf.LinkIdentifier, targetNodeIdx, sourceNodeIdx, targetIntf.TargetIpv4Addr)
		if result, ok := finalMapping[targetNormalNode.ContainerName]; !ok {
			result += targetDesc
			finalMapping[targetNormalNode.ContainerName] = result
		} else {
			result = targetDesc
			finalMapping[targetNormalNode.ContainerName] += result
		}
		allInterfaceList = allInterfaceList + targetDesc
	}

	// 进行所有的节点的遍历
	for _, abstractNode := range t.AllAbstractNodes {
		var normalNode *normal_node.NormalNode
		normalNode, err = abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("generate ifname to link identifier mapping files failed, %w", err)
		}
		// simulationDir 文件夹的位置
		simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
		// interface dir 文件架的位置
		outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "interface")
		// 进行文件夹的创建
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir for interface error: %w", err)
		}
		filePath := filepath.Join(outputDir, "interface.txt")
		// 进行写入
		var writeFile *os.File
		writeFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("Error opening file %s: %s\n", filePath, err)
		}
		_, err = writeFile.WriteString(finalMapping[normalNode.ContainerName])
		if err != nil {
			return fmt.Errorf("Error writing to file %s: %s\n", filePath, err)
		}
		// 如果是源节点的话则进行写入的操作
		if normalNode.Id == 1 {
			err = file.WriteStringIntoFile(filepath.Join(outputDir, "all_interfaces.txt"), allInterfaceList)
			if err != nil {
				return fmt.Errorf("create all_interfaces.txt file failed: %w", err)
			}
		}
	}

	t.topologyInitSteps[GenerateIfnameToLinkIdentifierMapping] = struct{}{}
	topologyLogger.Infof("generate ifname to link identifier")
	return nil
}

func (t *Topology) GenerateChainMakerIDtoNameMapping() error {
	if _, ok := t.topologyInitSteps[GenerateChainMakerIDtoNameMapping]; ok {
		topologyLogger.Infof("already generate chainmaker id to name mapping")
		return nil
	}

	// 将 mapping 写入文件之中
	finalString := ""
	count := 0
	for peerId, index := range t.chainMakerPrepare.PeerIdToIndexMapping {
		if count != (len(t.chainMakerPrepare.PeerIdToIndexMapping) - 1) {
			finalString += fmt.Sprintf("%s,%s\n", peerId, t.ChainmakerNodes[index].ContainerName)
		} else {
			finalString += fmt.Sprintf("%s,%s", peerId, t.ChainmakerNodes[index].ContainerName)
		}
	}
	// 需要进行所有的长安链节点的遍历
	for _, chainMakerNode := range t.ChainmakerNodes {
		mappingFilePath := fmt.Sprintf("%s/%s/peerIdToContainerName.txt",
			configs.TopConfiguration.PathConfig.ConfigGeneratePath,
			chainMakerNode.ContainerName)
		err := file.WriteStringIntoFile(mappingFilePath, finalString)
		fmt.Println(mappingFilePath)
		if err != nil {
			return fmt.Errorf("write string into file error: %v", err)
		}
	}

	t.topologyInitSteps[GenerateChainMakerIDtoNameMapping] = struct{}{}
	topologyLogger.Infof("generate chainmaker id to name mapping")
	return nil
}

// GenerateAtlasSegmentsAndOutputLinkIdentifiers 生成 Segments 和 output link identifiers
func (t *Topology) GenerateAtlasSegmentsAndOutputLinkIdentifiers() error {
	if _, ok := t.topologyInitSteps[GenerateAtlasSegmentsAndOutputLinkIdentifiers]; ok {
		topologyLogger.Infof("already generate segments and output link identifiers")
		return nil
	}

	// 图配置
	resourcesFilePath := configs.TopConfiguration.PathConfig.ResourcesPath
	multipathFileName := configs.TopConfiguration.PathValidationConfig.MultipathConfig.MultipathFileName
	multipathFilePath := filepath.Join(resourcesFilePath, fmt.Sprintf("multipath/complex/%s", multipathFileName))
	paths, segments, nodeSegmentsMapping, graphParams := graph.GenerateAtlasPathsAndSegmentsViaPathsFile(multipathFilePath)

	// 源节点的处理流程
	// ----------------------------------------------------------------------------------------------------
	var allUniqueOutputLinkIndentifiers = make(map[int]struct{})

	// 根据 paths 找到所有的 Link identifiers
	for _, path := range paths {
		sourceNodeName := path.NodeList[0].NodeName
		targetNodeName := path.NodeList[1].NodeName
		if abstractLink, ok := t.AllLinksMap[sourceNodeName][targetNodeName]; ok {
			if _, ok = allUniqueOutputLinkIndentifiers[abstractLink.SourceInterface.LinkIdentifier]; !ok {
				allUniqueOutputLinkIndentifiers[abstractLink.SourceInterface.LinkIdentifier] = struct{}{}
			}
		} else if abstractLink, ok = t.AllLinksMap[targetNodeName][sourceNodeName]; ok {
			if _, ok = allUniqueOutputLinkIndentifiers[abstractLink.TargetInterface.LinkIdentifier]; !ok {
				allUniqueOutputLinkIndentifiers[abstractLink.TargetInterface.LinkIdentifier] = struct{}{}
			}
		} else {
			fmt.Printf("cannot find link identifier\n")
		}
	}
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	nodeDir := filepath.Join(simulationDir, graphParams.Source)
	multipathOutputLinkIdentifiersFilePath := filepath.Join(nodeDir, "multipath_output_link_identifiers.txt")
	linkIdentifiersString := fmt.Sprintf("%d,%d,", graphParams.DestinationIndex, len(allUniqueOutputLinkIndentifiers))
	count := 0
	for linkIdentifier, _ := range allUniqueOutputLinkIndentifiers {
		if count != (len(allUniqueOutputLinkIndentifiers) - 1) {
			linkIdentifiersString += fmt.Sprintf("%d,", linkIdentifier)
		} else {
			linkIdentifiersString += fmt.Sprintf("%d", linkIdentifier)
		}
		count += 1
	}
	err := file.WriteStringIntoFile(multipathOutputLinkIdentifiersFilePath, linkIdentifiersString)
	if err != nil {
		return fmt.Errorf("write string into file error: %v", err)
	}

	// 打印所有的 segments
	multipathSegmentsFilePath := filepath.Join(nodeDir, "multipath_segments.txt")
	multipathSegmentsString := ""
	for index, segment := range segments {
		if index != (len(segments) - 1) {
			segmentString := entities.SegmentToIndexString(segment)
			multipathSegmentsString += fmt.Sprintf("%s\n", segmentString)
		} else {
			segmentString := entities.SegmentToIndexString(segment)
			multipathSegmentsString += fmt.Sprintf("%s", segmentString)
		}
	}
	err = file.WriteStringIntoFile(multipathSegmentsFilePath, multipathSegmentsString)
	if err != nil {
		return fmt.Errorf("write string into file error: %v", err)
	}

	// ----------------------------------------------------------------------------------------------------

	// 中间节点的处理流程
	// ----------------------------------------------------------------------------------------------------

	// 每个节点进行 nodeSegment 的存储
	for nodeName, nodeSegments := range nodeSegmentsMapping {
		if nodeName == graphParams.Source {
			continue
		}

		nodeDir = filepath.Join(simulationDir, nodeName)
		//fmt.Printf("intermediate node segment: %s", nodeDir)

		// 根据 segment 找到所有的可能的出接口
		// ------------------------------------------------------------------------------------------
		intermediateMultipathOutputLinkIdentifiers := filepath.Join(nodeDir, "intermediate_multipath_output_link_identifiers.txt")
		intermediateAllUniqueLinkIdentifiers := map[int]struct{}{}
		for _, path := range paths {
			for index, singleNode := range path.NodeList {
				if index != (len(path.NodeList) - 1) {
					if singleNode.NodeName == nodeName {
						sourceNodeName := nodeName
						targetNodeName := path.NodeList[index+1].NodeName
						if abstractLink, ok := t.AllLinksMap[sourceNodeName][targetNodeName]; ok {
							if _, ok = intermediateAllUniqueLinkIdentifiers[abstractLink.SourceInterface.LinkIdentifier]; !ok {
								intermediateAllUniqueLinkIdentifiers[abstractLink.SourceInterface.LinkIdentifier] = struct{}{}
							}
						} else if abstractLink, ok = t.AllLinksMap[targetNodeName][sourceNodeName]; ok {
							if _, ok = intermediateAllUniqueLinkIdentifiers[abstractLink.TargetInterface.LinkIdentifier]; !ok {
								intermediateAllUniqueLinkIdentifiers[abstractLink.TargetInterface.LinkIdentifier] = struct{}{}
							}
						} else {
							fmt.Printf("cannot find link identifier\n")
						}
					}
				}
			}
		}

		if len(intermediateAllUniqueLinkIdentifiers) > 0 {
			linkIdentifiersString = fmt.Sprintf("%d,%d,", graphParams.DestinationIndex, len(intermediateAllUniqueLinkIdentifiers))
			count = 0
			for linkIdentifier, _ := range intermediateAllUniqueLinkIdentifiers {
				if count != (len(intermediateAllUniqueLinkIdentifiers) - 1) {
					linkIdentifiersString += fmt.Sprintf("%d,", linkIdentifier)
				} else {
					linkIdentifiersString += fmt.Sprintf("%d", linkIdentifier)
				}
				count += 1
			}
			err = file.WriteStringIntoFile(intermediateMultipathOutputLinkIdentifiers, linkIdentifiersString)
			if err != nil {
				return fmt.Errorf("write string into file error: %v", err)
			}
		}
		// ------------------------------------------------------------------------------------------
		// 为每个 segment 找到对应的 link identifier

		// 进行每个中间节点的 segment list 的构造
		// ------------------------------------------------------------------------------------------
		intermediateMultipathSegmentsFilePath := filepath.Join(nodeDir, "intermediate_multipath_segments.txt")
		intermediateMultipathSegmentsString := ""
		for index, nodeSegment := range nodeSegments {
			if index != (len(nodeSegments) - 1) {
				segmentString := entities.SegmentToIndexString(nodeSegment)
				intermediateMultipathSegmentsString += fmt.Sprintf("%s\n", segmentString)
			} else {
				segmentString := entities.SegmentToIndexString(nodeSegment)
				intermediateMultipathSegmentsString += fmt.Sprintf("%s", segmentString)
			}
		}
		err = file.WriteStringIntoFile(intermediateMultipathSegmentsFilePath, intermediateMultipathSegmentsString)
		if err != nil {
			return fmt.Errorf("write string into file error: %v", err)
		}
		// ------------------------------------------------------------------------------------------
	}
	// ----------------------------------------------------------------------------------------------------

	t.topologyInitSteps[GenerateAtlasSegmentsAndOutputLinkIdentifiers] = struct{}{}
	topologyLogger.Infof("generate segments and output link identifiers")
	return nil
}

// GenerateMultipathSelirPathsAndOutputLinkIdentifiers (1) 源和目的节点插入全部的路由 (2) 中间节点只需要存储 output_link_identifiers. 确实不需要 array_based and hash_based routing table
func (t *Topology) GenerateMultipathSelirPathsAndOutputLinkIdentifiers() error {
	if _, ok := t.topologyInitSteps[GenerateMultipathSelirPathsAndOutputLinkIdentifiers]; ok {
		topologyLogger.Infof("already generate multipath selir mulitpaths and output link identifiers")
		return nil
	}

	// 最终结果
	var finalResult []string

	// 图配置
	// ----------------------------------------------------------------------------------------------------
	resourcesFilePath := configs.TopConfiguration.PathConfig.ResourcesPath
	multipathFileName := configs.TopConfiguration.PathValidationConfig.MultipathConfig.MultipathFileName
	multipathFilePath := filepath.Join(resourcesFilePath, fmt.Sprintf("multipath/complex/%s", multipathFileName))
	multipathSelirComplexGraph, paths, nameToNodeMapping, sourceStr, destinationStr := graph.GenerateMultipathSelirMultipathsViaPathsFile(multipathFilePath)

	// 进行源和目的节点的提取
	sourceIndex, _ := extract.NumberFromString(sourceStr)
	destinationIndex, _ := extract.NumberFromString(destinationStr)

	// 进行 relationship 的生成
	// ----------------------------------------------------------------------------------------------------
	for splitNode, mappings := range multipathSelirComplexGraph.RelationshipMapping {
		simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
		outputDir := filepath.Join(simulationDir, splitNode)
		relationshipFilePath := filepath.Join(outputDir, "relationship.txt")
		finalString := ""
		for index, singleMapping := range mappings {
			if index != (len(singleMapping) - 1) {
				finalString += singleMapping + "\n"
			} else {
				finalString += singleMapping
			}
		}
		err := file.WriteStringIntoFile(relationshipFilePath, finalString)
		if err != nil {
			return fmt.Errorf("write string into file error: %v", err)
		}
	}

	// ----------------------------------------------------------------------------------------------------

	// 根据 paths 进行路由条目的生成
	// ----------------------------------------------------------------------------------------------------
	for _, path := range paths {
		var linkIdentifiers []int
		var nodeIds []int
		for index := 0; index < len(path.NodeList); index++ {
			if index != (len(path.NodeList) - 1) {
				nextNodeIndex := index + 1
				sourceNode := path.NodeList[index]
				targetNode := path.NodeList[nextNodeIndex]
				sourceNodeName := sourceNode.NodeName
				targetNodeName := targetNode.NodeName
				var abstractLink *link.AbstractLink
				var ok bool
				if abstractLink, ok = t.AllLinksMap[sourceNodeName][targetNodeName]; ok {
					abstractLink = t.AllLinksMap[sourceNodeName][targetNodeName]
				} else if abstractLink, ok = t.AllLinksMap[targetNodeName][sourceNodeName]; ok {
					abstractLink = t.AllLinksMap[targetNodeName][sourceNodeName]
				} else {
					abstractLink = nil
					return fmt.Errorf("cannot find link")
				}
				linkIdentifier := abstractLink.SourceInterface.LinkIdentifier
				nodeId := targetNode.Index
				linkIdentifiers = append(linkIdentifiers, linkIdentifier)
				nodeIds = append(nodeIds, nodeId)
			}
		}
		generateLiRRouteString := route.GenerateLiRRoutingString((int64)(sourceIndex), (int64)(destinationIndex), linkIdentifiers, nodeIds)
		finalResult = append(finalResult, generateLiRRouteString)
	}
	// ----------------------------------------------------------------------------------------------------

	// 源和目的路由配置
	// ----------------------------------------------------------------------------------------------------

	// simulation dir 的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath

	// containerNames
	containerNames := []string{sourceStr}
	// 遍历 containerName 进行写入 (只向源和目的节点进行写入)
	for _, containerName := range containerNames {
		// route dir 文件的位置
		outputDir := filepath.Join(simulationDir, containerName, "route")
		// 进行文件夹的生成
		err := dir.Generate(outputDir)
		if err != nil {
			return fmt.Errorf("write route error: %w", err)
		}
		// 文件的路径
		var filePath string
		if containerName == destinationStr {
			fmt.Printf("write dest_multipath.txt\n")
			filePath = filepath.Join(outputDir, "dest_multipath.txt")
		} else {
			fmt.Printf("write multipath.txt\n")
			filePath = filepath.Join(outputDir, "multipath.txt")
		}
		// 创建写入文件
		var lirRouteFile *os.File
		lirRouteFile, err = os.Create(filePath)
		defer func() {
			closeErr := lirRouteFile.Close()
			if err == nil {
				err = closeErr
			}
		}()
		if err != nil {
			return fmt.Errorf("write route error: %w", err)
		}
		// 进行实际的写入
		_, err = lirRouteFile.WriteString(strings.Join(finalResult, "\n"))
		if err != nil {
			return fmt.Errorf("write route failed: %v", err)
		}
	}

	// 将存在多少的路径进行写入
	for _, abstractNode := range t.AllAbstractNodes {
		normalNode, _ := abstractNode.GetNormalNodeFromAbstractNode()
		filePath := filepath.Join(simulationDir, normalNode.ContainerName, "num_of_paths.txt")
		_ = file.WriteStringIntoFile(filePath, fmt.Sprintf("%d", len(paths)))
	}

	// ----------------------------------------------------------------------------------------------------

	// 中间节点出接口配置
	// ----------------------------------------------------------------------------------------------------
	for _, singleNode := range nameToNodeMapping {
		intermediateAllUniqueLinkIdentifiers := map[int]struct{}{}
		// 进行所有的 path 的遍历
		for _, path := range paths {
			for index := 0; index < len(path.NodeList); index++ {
				if index != (len(path.NodeList) - 1) {
					nextNodeIndex := index + 1
					sourceNode := path.NodeList[index]
					targetNode := path.NodeList[nextNodeIndex]
					sourceNodeName := sourceNode.NodeName
					targetNodeName := targetNode.NodeName
					if sourceNodeName == singleNode.NodeName {
						if abstractLink, ok := t.AllLinksMap[sourceNodeName][targetNodeName]; ok {
							if _, ok = intermediateAllUniqueLinkIdentifiers[abstractLink.SourceInterface.LinkIdentifier]; !ok {
								intermediateAllUniqueLinkIdentifiers[abstractLink.SourceInterface.LinkIdentifier] = struct{}{}
							}
						} else if abstractLink, ok = t.AllLinksMap[targetNodeName][sourceNodeName]; ok {
							if _, ok = intermediateAllUniqueLinkIdentifiers[abstractLink.TargetInterface.LinkIdentifier]; !ok {
								intermediateAllUniqueLinkIdentifiers[abstractLink.TargetInterface.LinkIdentifier] = struct{}{}
							}
						} else {
							fmt.Printf("cannot find link identifier\n")
						}
					}
				}
			}
		}

		if len(intermediateAllUniqueLinkIdentifiers) > 0 {
			// 构建 link identifier string
			linkIdentifiersString := fmt.Sprintf("%d,%d,", destinationIndex, len(intermediateAllUniqueLinkIdentifiers))
			count := 0
			for linkIdentifier, _ := range intermediateAllUniqueLinkIdentifiers {
				if count != (len(intermediateAllUniqueLinkIdentifiers) - 1) {
					linkIdentifiersString += fmt.Sprintf("%d,", linkIdentifier)
				} else {
					linkIdentifiersString += fmt.Sprintf("%d", linkIdentifier)
				}
				count += 1
			}

			// 将 uniqueLinkIdentifiersMapping 写入
			nodeDir := filepath.Join(simulationDir, singleNode.NodeName)
			intermediateMultipathOutputLinkIdentifiers := filepath.Join(nodeDir, "intermediate_multipath_output_link_identifiers.txt")
			fmt.Println("write string into file", intermediateMultipathOutputLinkIdentifiers)
			err := file.WriteStringIntoFile(intermediateMultipathOutputLinkIdentifiers, linkIdentifiersString)
			if err != nil {
				return fmt.Errorf("cannot write string into file: %v", err)
			}
		}
	}
	// ----------------------------------------------------------------------------------------------------

	t.topologyInitSteps[GenerateMultipathSelirPathsAndOutputLinkIdentifiers] = struct{}{}
	return nil
}

//func (t *Topology) GenerateTopologyForSecPathMab() error {
//	if _, ok := t.topologyInitSteps[GenerateTopologyForSecPathMab]; ok {
//		topologyLogger.Infof("already generate topology for sec path mab")
//		return nil
//	}
//
//	simulatorInstance := steps.NewSimulator(online_securest_path.SimulatorParamsInstance,
//		online_securest_path.ExperimentResultsDir,
//		[]*online_entities.SimEvent{})
//
//	err := simulatorInstance.Init()
//	if err != nil {
//		return fmt.Errorf("init simulator failed: %v", err)
//	}
//
//	normalNode, err := t.AllAbstractNodes[0].GetNormalNodeFromAbstractNode()
//	if err != nil {
//		return fmt.Errorf("get normal node from abstract node failed: %v", err)
//	}
//
//	topologyDir := fmt.Sprintf("%s/%s/topology", configs.TopConfiguration.PathConfig.ConfigGeneratePath, normalNode.ContainerName)
//	err = os.MkdirAll(topologyDir, os.ModePerm)
//	if err != nil {
//		return fmt.Errorf("mkdir failed for %s", topologyDir)
//	}
//
//	// 将  sec_path_mab_topology.txt 进行写入, 主要是为了生成所有的 available paths
//	outputFilePath := fmt.Sprintf("%s/sec_path_mab_topology.txt", topologyDir)
//	err = simulatorInstance.GenerateFileForSecPathMabSourceNode(outputFilePath)
//	if err != nil {
//		return fmt.Errorf("generate file for sec path mab source node failed: %v", err)
//	}
//
//	t.topologyInitSteps[GenerateTopologyForSecPathMab] = struct{}{}
//	topologyLogger.Infof("generate topology for sec path mab")
//	return nil
//}
