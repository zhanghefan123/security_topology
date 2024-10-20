package topology

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/network"
	"zhanghefan123/security_topology/services/http/params"
)

type InitFunction func() error

const (
	GenerateNodes                 = "GenerateNodes"                 // 生成卫星
	GenerateSubnets               = "GenerateSubnets"               // 创建子网
	GenerateLinks                 = "GenerateLinks"                 // 生成链路
	GenerateFrrConfigurationFiles = "GenerateFrrConfigurationFiles" // 生成 frr 配置
)

// Init 进行初始化
func (t *Topology) Init() error {
	initSteps := []map[string]InitFunction{
		{GenerateNodes: t.GenerateNodes},
		{GenerateSubnets: t.GenerateSubnets},
		{GenerateLinks: t.GenerateLinks},
		{GenerateFrrConfigurationFiles: t.GenerateFrrConfigurationFiles},
	}
	err := t.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("constellation init failed: %w", err)
	}
	return nil
}

// InitializeSteps 按步骤进行初始化
func (t *Topology) initializeSteps(initSteps []map[string]InitFunction) (err error) {
	fmt.Println()
	moduleNum := len(initSteps)
	for idx, initStep := range initSteps {
		for name, initFunc := range initStep {
			if err = initFunc(); err != nil {
				return fmt.Errorf("init step [%s] failed, %s", name, err)
			}
			topologyLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
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
	for _, nodeParam := range t.topologyParams.Nodes {
		nodeType, err := types.ResolveNodeType(nodeParam.Type)
		if err != nil {
			return fmt.Errorf("resolve node type failed, %s", err)
		}
		switch *nodeType {
		case types.NetworkNodeType_Router: // 进行普通路由节点的添加
			routerTmp := nodes.NewRouter(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.Routers = append(t.Routers, routerTmp)
			// 注意只能唯一创建一次
			abstractRouter := node.NewAbstractNode(types.NetworkNodeType_Router, routerTmp)
			t.RouterAbstractNodes = append(t.RouterAbstractNodes, abstractRouter)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractRouter)
		case types.NetworkNodeType_NormalNode:
			normalNodeTmp := normal_node.NewNormalNode(types.NetworkNodeType_NormalNode, nodeParam.Index, fmt.Sprintf("%s-%d", nodeType.String(), nodeParam.Index))
			t.NormalNodes = append(t.NormalNodes, normalNodeTmp)
			// 注意只能唯一创建一次
			abstractNormalNode := node.NewAbstractNode(types.NetworkNodeType_NormalNode, normalNodeTmp)
			t.NormalAbstractNodes = append(t.NormalAbstractNodes, abstractNormalNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractNormalNode)
		case types.NetworkNodeType_ConsensusNode:
			consensusNodeTmp := nodes.NewConsensusNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.ConsensusNodes = append(t.ConsensusNodes, consensusNodeTmp)
			// 注意只能唯一创建一次
			abstractConsensusNode := node.NewAbstractNode(types.NetworkNodeType_ConsensusNode, consensusNodeTmp)
			t.ConsensusAbstractNodes = append(t.ConsensusAbstractNodes, abstractConsensusNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractConsensusNode)
		case types.NetworkNodeType_MaliciousNode:
			maliciousNodeTmp := nodes.NewMaliciousNode(nodeParam.Index, nodeParam.X, nodeParam.Y)
			t.MaliciousNodes = append(t.MaliciousNodes, maliciousNodeTmp)
			// 注意只能唯一创建一次
			abstractMaliciousNode := node.NewAbstractNode(types.NetworkNodeType_MaliciousNode, maliciousNodeTmp)
			t.MaliciousAbstractNodes = append(t.MaliciousAbstractNodes, abstractMaliciousNode)
			t.AllAbstractNodes = append(t.AllAbstractNodes, abstractMaliciousNode)
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
	for _, linkTmp := range t.topologyParams.Links {
		// 拿到源节点和目的节点的参数
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
		linkType := types.NetworkLinkType_NormalLink
		ipv4SubNet := t.Ipv4SubNets[currentLinkNums]                                                                            // 获取当前ipv4 子网
		ipv6SubNet := t.Ipv6SubNets[currentLinkNums]                                                                            // 获取当前 ipv6 子网
		sourceNormalNode.ConnectedIpv4SubnetList = append(sourceNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String())        // 卫星添加ipv4子网
		targetNormalNode.ConnectedIpv4SubnetList = append(targetNormalNode.ConnectedIpv4SubnetList, ipv4SubNet.String())        // 卫星添加ipv4子网
		sourceNormalNode.ConnectedIpv6SubnetList = append(sourceNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String())        // 卫星添加ipv6子网
		targetNormalNode.ConnectedIpv6SubnetList = append(targetNormalNode.ConnectedIpv6SubnetList, ipv6SubNet.String())        // 卫星添加ipv6子网
		sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                    // 提取ipv4第一个和第二个地址
		sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                    // 提取ipv6第一个和第二个地址
		sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(sourceNodeType), sourceNormalNode.Id, sourceNormalNode.Ifidx) // 源接口名
		targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(targetNodeType), targetNormalNode.Id, targetNormalNode.Ifidx) // 目的接口名
		sourceIntf := intf.NewNetworkInterface(sourceNormalNode.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr)            // 创建第一个接口
		targetIntf := intf.NewNetworkInterface(targetNormalNode.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr)            // 创建第二个接口
		sourceNormalNode.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                        // 设置源卫星地址
		targetNormalNode.IfNameToInterfaceMap[targetIfName] = targetIntf                                                        // 设置目的卫星地址
		abstractLink := link.NewAbstractLink(linkType,
			linkId,
			sourceNodeType, targetNodeType,
			sourceNormalNode.Id, targetNormalNode.Id,
			sourceIntf, targetIntf,
			sourceAbstractNode, targetAbstractNode)
		sourceNormalNode.Ifidx++
		targetNormalNode.Ifidx++
		t.Links = append(t.Links, abstractLink)
		if _, ok := t.LinksMap[sourceNormalNode.ContainerName]; !ok {
			t.LinksMap[sourceNormalNode.ContainerName] = make(map[string]*link.AbstractLink)
		}
		t.LinksMap[sourceNormalNode.ContainerName][targetNormalNode.ContainerName] = abstractLink
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
	case types.NetworkNodeType_MaliciousNode:
		sourceNode = t.MaliciousAbstractNodes[sourceNodeParam.Index-1]
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
	case types.NetworkNodeType_MaliciousNode:
		targetNode = t.MaliciousAbstractNodes[targetNodeParam.Index-1]
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
	for _, router := range t.Routers {
		if selectedOspfVersion == "ospfv2" {
			// 生成 ospfv2 配置
			err := router.GenerateOspfV2FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed, %s", err)
			}
		} else if selectedOspfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err := router.GenerateOspfV3FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv3 frr configuration files failed, %s", err)
			}
		} else {
			return fmt.Errorf("unsupported ospf version: %s", selectedOspfVersion)
		}

	}

	for _, normalNode := range t.NormalNodes {
		if selectedOspfVersion == "ospfv2" {
			// 生成 ospfv2 配置
			err := normalNode.GenerateOspfV2FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed, %s", err)
			}
		} else if selectedOspfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err := normalNode.GenerateOspfV3FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv3 frr configuration files failed, %s", err)
			}
		} else {
			return fmt.Errorf("unsupported ospf version: %s", selectedOspfVersion)
		}
	}

	for _, consensusNode := range t.ConsensusNodes {
		if selectedOspfVersion == "ospfv2" {
			// 生成 ospfv2 配置
			err := consensusNode.GenerateOspfV2FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed, %s", err)
			}
		} else if selectedOspfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err := consensusNode.GenerateOspfV3FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv3 frr configuration files failed, %s", err)
			}
		} else {
			return fmt.Errorf("unsupported ospf version: %s", selectedOspfVersion)
		}
	}

	for _, maliciousNode := range t.MaliciousNodes {
		if selectedOspfVersion == "ospfv2" {
			// 生成 ospfv2 配置
			err := maliciousNode.GenerateOspfV2FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed, %s", err)
			}
		} else if selectedOspfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err := maliciousNode.GenerateOspfV3FrrConfig()
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
