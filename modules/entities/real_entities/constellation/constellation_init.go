package constellation

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"reflect"
	"strconv"
	"zhanghefan123/security_topology/api/route"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/network"
	"zhanghefan123/security_topology/modules/utils/position"
)

const (
	GenerateSatellites             = "GenerateSatellites"             // 生成卫星
	GenerateSubnets                = "GenerateIPv4Subnets"            // 创建子网
	GenerateLinks                  = "GenerateLinks"                  // 生成链路
	GenerateFrrConfigurationFiles  = "GenerateFrrConfigurationFiles"  // 生成 frr 配置
	GeneratePeerIdAndPrivateKey    = "GeneratePeerIdAndPrivateKey"    // 生成 peerId 以及私钥
	CalculateAndWriteSegmentRoutes = "CalculateAndWriteSegmentRoutes" // 进行段路由的计算
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

// Init 进行初始化
func (c *Constellation) Init() error {

	enableSRv6 := configs.TopConfiguration.NetworkConfig.EnableSRv6

	initSteps := []map[string]InitModule{
		{GenerateSatellites: InitModule{true, c.GenerateSatellites}},
		{GenerateSubnets: InitModule{true, c.GenerateSubnets}},
		{GenerateLinks: InitModule{true, c.GenerateLinks}},
		{GenerateFrrConfigurationFiles: InitModule{true, c.GenerateFrrConfigurationFiles}},
		{GeneratePeerIdAndPrivateKey: InitModule{true, c.GeneratePeerIdAndPrivateKey}},
		{CalculateAndWriteSegmentRoutes: InitModule{enableSRv6, c.CalculateAndWriteSegmentRoutes}},
	}
	err := c.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("constellation init failed %w", err)
	}
	return nil
}

// initStepsNum 初始化模块的数量
func (c *Constellation) initStepsNum(initSteps []map[string]InitModule) int {
	result := 0
	for _, initStep := range initSteps {
		for _, startModule := range initStep {
			if startModule.init {
				result += 1
			}
		}
	}
	return result
}

// InitializeSteps 按步骤进行初始化
func (c *Constellation) initializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := c.initStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.init {
				if err = initModule.initFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				constellationLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

// GenerateSatellites 生成卫星
func (c *Constellation) GenerateSatellites() error {
	if _, ok := c.systemInitSteps[GenerateSatellites]; ok {
		constellationLogger.Infof("already generate satellites")
		return nil // 重复生成没有必要，但是实际上只要返回就行，不是错误
	}

	startYear, startDay := position.GetYearAndDay(c.startTime)
	firstLineTleTemplate := "1 00000U 23666A   %02d%012.8f  .00000000  00000-0 00000000 0 0000"
	secondLineTleTemplate := "2 00000  90.0000 %08.4f 0000011   0.0000 %8.4f %11.8f00000"
	var startLatitude float32 = 0
	var startLongitude float32 = 0
	var delta float32 = 5
	freq := 1 / 0.06965
	for orbitId := 0; orbitId < c.OrbitNumber; orbitId++ {
		orbitStartLatitude := startLatitude + delta
		orbitLongitude := startLongitude + 180*float32(orbitId)/float32(c.SatellitePerOrbit)
		for nodeId := c.SatellitePerOrbit * orbitId; nodeId < c.SatellitePerOrbit*(orbitId+1); nodeId++ {
			// 轨道内编号
			indexInOrbit := nodeId % c.SatellitePerOrbit
			// 获取卫星的纬度
			satelliteLatitude := orbitStartLatitude + 360*float32(indexInOrbit)/float32(c.SatellitePerOrbit)
			// 卫星的经度和轨道的经度一致
			satelliteLongitude := orbitLongitude
			// 获取 TLE
			firstLineTle := fmt.Sprintf(firstLineTleTemplate, startYear, startDay)
			secondLineTle := fmt.Sprintf(secondLineTleTemplate, satelliteLongitude, satelliteLatitude, freq)
			// 存储 TLE
			tle := make([]string, 2)
			tle[0] = firstLineTle + strconv.Itoa(position.TleCheckSum(firstLineTle))
			tle[1] = secondLineTle + strconv.Itoa(position.TleCheckSum(secondLineTle))
			// 判断该进行什么卫星的生成
			if c.SatelliteType == types.NetworkNodeType_NormalSatellite { // 1. 如果是生成普通卫星
				// 创建普通卫星
				normalSatellite := satellites.NewNormalSatellite(nodeId+1, orbitId, indexInOrbit, tle)
				// 添加卫星
				c.NormalSatellites = append(c.NormalSatellites, normalSatellite)
				// 创建抽象节点
				normalSatelliteAbstract := node.NewAbstractNode(normalSatellite.Type, normalSatellite, ConstellationInstance.ConstellationGraph)
				// 将 satellite 放到 allAbstractNodes 之中
				c.AllAbstractNodes = append(c.AllAbstractNodes, normalSatelliteAbstract)
			} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite { // 2. 如果是生成共识卫星
				// 创建共识卫星
				consensusSatellite := satellites.NewConsensusSatellite(nodeId+1, orbitId, indexInOrbit, c.SatelliteRPCPort, c.SatelliteP2PPort, tle)
				// 添加卫星
				c.ConsensusSatellites = append(c.ConsensusSatellites, consensusSatellite)
				// 创建抽象节点
				consensusSatelliteAbstract := node.NewAbstractNode(consensusSatellite.Type, consensusSatellite, ConstellationInstance.ConstellationGraph)
				// 将 satellite 放到 allAbstractNodes 之中
				c.AllAbstractNodes = append(c.AllAbstractNodes, consensusSatelliteAbstract)
			} else {
				return fmt.Errorf("not supported network node type")
			}

		}
	}

	c.systemInitSteps[GenerateSatellites] = struct{}{}
	constellationLogger.Infof("generate satellites")

	return nil
}

// GenerateSubnets 进行子网的生成
func (c *Constellation) GenerateSubnets() error {
	if _, ok := c.systemInitSteps[GenerateSubnets]; ok {
		constellationLogger.Infof("already generate subnets")
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
	c.Ipv4SubNets = ipv4Subnets

	// 进行 ipv6 的子网的生成
	ipv6Subnets, err = network.GenerateIpv6Subnets(configs.TopConfiguration.NetworkConfig.BaseV6NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	c.Ipv6SubNets = ipv6Subnets

	c.systemInitSteps[GenerateSubnets] = struct{}{}
	constellationLogger.Infof("generate subnets")
	return nil
}

// GenerateLinks 进行链路的生成
func (c *Constellation) GenerateLinks() error {
	if _, ok := c.systemInitSteps[GenerateLinks]; ok {
		constellationLogger.Infof("already generate links")
		return nil
	}

	if c.SatelliteType == types.NetworkNodeType_NormalSatellite {
		c.generateLinksForNormalSatellite()
	} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite {
		c.generateLinksForConsensusSatellites()
	} else {
		return fmt.Errorf("not supported network node type")
	}

	c.systemInitSteps[GenerateLinks] = struct{}{}
	constellationLogger.Infof("generate links")
	return nil
}

// GenerateLinksForNormalSatellite 为共识卫星生成链路
func (c *Constellation) generateLinksForConsensusSatellites() {
	for index, sat := range c.ConsensusSatellites {
		satReal := sat
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceAbstract := c.AllAbstractNodes[index]
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat := c.ConsensusSatellites[targetSatId] // 目的节点的抽象标识
		targetAbstract := c.AllAbstractNodes[targetSatId]
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                // 当前链路数量
			linkId := currentLinkNums + 1                                                                                              // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink                                                                  // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                                       // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                               // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                               // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                       // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                       // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                        // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                        // 目的接口名
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                  // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                            // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                  // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                            // 添加接口
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth, ConstellationInstance.ConstellationGraph)
			sourceSat.Ifidx++                                                 // 接口索引变化
			targetSat.Ifidx++                                                 // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, intraOrbitLink) // 添加到所有链路集合
			if _, ok := c.AllSatelliteLinksMap[sourceSat.ContainerName]; !ok {
				c.AllSatelliteLinksMap[sourceSat.ContainerName] = make(map[string]*link.AbstractLink)
			}
			c.AllSatelliteLinksMap[sourceSat.ContainerName][targetSat.ContainerName] = intraOrbitLink
			c.IntraOrbitSatelliteLinks = append(c.IntraOrbitSatelliteLinks, intraOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成同轨道的星间链路 ---------------->
		// <---------------- 生成异轨道的星间链路 ---------------->
		targetOrbitId = sourceOrbitId + 1
		if targetOrbitId < c.OrbitNumber {
			targetIndexInOrbit = sourceSat.IndexInOrbit
			targetSatId = targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
			targetSat = c.ConsensusSatellites[targetSatId]
			targetAbstract = c.AllAbstractNodes[targetSatId]
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                // 当前链路数量
			linkId := currentLinkNums + 1                                                                                              // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                                                  // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                                       // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                               // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                               // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                       // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                       // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                        // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                        // 目的接口名
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                  // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                            // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                  // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                            // 添加接口
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth, ConstellationInstance.ConstellationGraph)
			sourceSat.Ifidx++                                                 // 接口索引变化
			targetSat.Ifidx++                                                 // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, interOrbitLink) // 添加到所有链路集合
			if _, ok := c.AllSatelliteLinksMap[sourceSat.ContainerName]; !ok {
				c.AllSatelliteLinksMap[sourceSat.ContainerName] = make(map[string]*link.AbstractLink)
			}
			c.AllSatelliteLinksMap[sourceSat.ContainerName][targetSat.ContainerName] = interOrbitLink
			c.InterOrbitSatelliteLinks = append(c.InterOrbitSatelliteLinks, interOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成异轨道的星间链路 ---------------->
	}
}

// GenerateLinksForNormalSatellite 为普通卫星生成链路
func (c *Constellation) generateLinksForNormalSatellite() {
	for index, sat := range c.NormalSatellites {
		satReal := sat
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceAbstract := c.AllAbstractNodes[index]
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat := c.NormalSatellites[targetSatId] // 目的节点的抽象标识
		targetAbstract := c.AllAbstractNodes[targetSatId]
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                // 当前链路数量
			linkId := currentLinkNums + 1                                                                                              // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink                                                                  // 链路类型
			nodeType := types.NetworkNodeType_NormalSatellite                                                                          // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                               // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                               // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                       // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                       // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                        // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                        // 目的接口名
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                  // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                            // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                  // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                            // 添加接口
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph)
			sourceSat.Ifidx++                                                 // 接口索引变化
			targetSat.Ifidx++                                                 // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, intraOrbitLink) // 添加到所有链路集合
			if _, ok := c.AllSatelliteLinksMap[sourceSat.ContainerName]; !ok {
				c.AllSatelliteLinksMap[sourceSat.ContainerName] = make(map[string]*link.AbstractLink)
			}
			c.AllSatelliteLinksMap[sourceSat.ContainerName][targetSat.ContainerName] = intraOrbitLink
			c.IntraOrbitSatelliteLinks = append(c.IntraOrbitSatelliteLinks, intraOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成同轨道的星间链路 ---------------->
		// <---------------- 生成异轨道的星间链路 ---------------->
		targetOrbitId = sourceOrbitId + 1
		if targetOrbitId < c.OrbitNumber {
			targetIndexInOrbit = sourceSat.IndexInOrbit
			targetSatId = targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
			targetSat = c.NormalSatellites[targetSatId]
			targetAbstract = c.AllAbstractNodes[targetSatId]
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                // 当前链路数量
			linkId := currentLinkNums + 1                                                                                              // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                                                  // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                                       // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                               // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                               // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                         // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                         // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                       // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                       // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                        // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                        // 目的接口名
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                   // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                  // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                            // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                  // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                            // 添加接口
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph)
			sourceSat.Ifidx++                                                 // 接口索引变化
			targetSat.Ifidx++                                                 // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, interOrbitLink) // 添加到所有链路集合
			if _, ok := c.AllSatelliteLinksMap[sourceSat.ContainerName]; !ok {
				c.AllSatelliteLinksMap[sourceSat.ContainerName] = make(map[string]*link.AbstractLink)
			}
			c.AllSatelliteLinksMap[sourceSat.ContainerName][targetSat.ContainerName] = interOrbitLink
			c.InterOrbitSatelliteLinks = append(c.InterOrbitSatelliteLinks, interOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成异轨道的星间链路 ---------------->
	}
}

// GenerateFrrConfigurationFiles 生成 frr 配置文件
func (c *Constellation) GenerateFrrConfigurationFiles() error {
	if _, ok := c.systemInitSteps[GenerateFrrConfigurationFiles]; ok {
		constellationLogger.Infof("already generate frr configuration files")
		return nil
	}

	for _, abstractNode := range c.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("cannot convert abstract node to normal node")
		}
		ospfVersion := configs.TopConfiguration.NetworkConfig.OspfVersion
		if ospfVersion == "ospfv2" {
			// 生成 ospfv6 配置
			err = normalNode.GenerateOspfV2FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed: %w", err)
			}
		} else if ospfVersion == "ospfv3" {
			// 生成 ospfv6 配置
			err = normalNode.GenerateOspfV3FrrConfig()
			if err != nil {
				return fmt.Errorf("generate ospfv3 frr configuration files failed: %w", err)
			}
		} else {
			return fmt.Errorf("unsupported ospf version %s", ospfVersion)
		}
	}

	c.systemInitSteps[GenerateFrrConfigurationFiles] = struct{}{}
	constellationLogger.Infof("generate frr configuration files")
	return nil
}

func (c *Constellation) GeneratePeerIdAndPrivateKey() error {
	if _, ok := c.systemInitSteps[GeneratePeerIdAndPrivateKey]; ok {
		constellationLogger.Infof("already generate peer id and private key")
		return nil
	}

	for _, abstractNode := range c.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("cannot convert abstract node to normal node")
		}
		err = normalNode.GeneratePeerIdAndPrivateKey()
		if err != nil {
			return fmt.Errorf("generate peer id and private key failed: %w", err)
		}
	}

	c.systemInitSteps[GeneratePeerIdAndPrivateKey] = struct{}{}
	constellationLogger.Infof("generate peer id and private key")
	return nil
}

// CalculateAndWriteSegmentRoutes 进行段路由的计算
func (c *Constellation) CalculateAndWriteSegmentRoutes() error {
	if _, ok := c.systemInitSteps[CalculateAndWriteSegmentRoutes]; ok {
		constellationLogger.Infof("already calculate segment routes")
		return nil
	}

	for _, abstractNode := range c.AllAbstractNodes {
		err := route.CalculateAndWriteSegmentRoute(abstractNode, &(c.AllSatelliteLinksMap), ConstellationInstance.ConstellationGraph)
		if err != nil {
			return fmt.Errorf("calculate route failed: %w", err)
		}
	}

	c.systemInitSteps[CalculateAndWriteSegmentRoutes] = struct{}{}
	constellationLogger.Infof("calculate segment routes")
	return nil
}
