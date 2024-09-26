package constellation

import (
	"fmt"
	"reflect"
	"strconv"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/position"
	"zhanghefan123/security_topology/modules/utils/subnet"
)

const (
	GenerateSatellites            = "GenerateSatellites"            // 生成卫星
	GenerateSubnets               = "GenerateSubnets"               // 创建子网
	GenerateLinks                 = "GenerateLinks"                 // 生成链路
	GenerateFrrConfigurationFiles = "GenerateFrrConfigurationFiles" // 生成 frr 配置
)

type InitFunction func() error

// Init 进行初始化
func (c *Constellation) Init() {
	initSteps := []map[string]InitFunction{
		{GenerateSatellites: c.GenerateSatellites},
		{GenerateSubnets: c.GenerateSubnets},
		{GenerateLinks: c.GenerateLinks},
		{GenerateFrrConfigurationFiles: c.GenerateFrrConfigurationFiles},
	}
	err := c.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		constellationLogger.Errorf("constellation init failed: %v", err)
	}
}

// InitializeSteps 按步骤进行初始化
func (c *Constellation) initializeSteps(initSteps []map[string]InitFunction) (err error) {
	fmt.Println()
	moduleNum := len(initSteps)
	for idx, initStep := range initSteps {
		for name, initFunc := range initStep {
			if err := initFunc(); err != nil {
				return fmt.Errorf("init step [%s] failed, %s", name, err)
			}
			constellationLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
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
				sat := satellite.NewNormalSatellite(nodeId+1, orbitId,
					indexInOrbit, c.SatelliteImageName, tle)
				// 将普通卫星放在抽象节点之中
				abstractNode := node.NewAbstractNode(types.NetworkNodeType_NormalSatellite, sat)
				// 添加卫星
				c.Satellites = append(c.Satellites, abstractNode)
			} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite { // 2. 如果是生成共识卫星
				// 创建共识卫星
				sat := satellite.NewConsensusSatellite(nodeId+1, orbitId,
					indexInOrbit, c.SatelliteImageName,
					c.SatelliteRPCPort, c.SatelliteP2PPort, tle)
				abstractNode := node.NewAbstractNode(types.NetworkNodeType_ConsensusSatellite, sat)
				// 添加卫星
				c.Satellites = append(c.Satellites, abstractNode)
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
	}

	subNets, err := subnet.GenerateSubnets(configs.TopConfiguration.NetworkConfig.BaseNetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	c.SubNets = subNets

	c.systemInitSteps[GenerateSubnets] = struct{}{}
	constellationLogger.Infof("generate subnets")
	return nil
}

// GenerateLinks 进行链路的生成
func (c *Constellation) GenerateLinks() error {
	if _, ok := c.systemInitSteps[GenerateLinks]; ok {
		constellationLogger.Infof("already generate links")
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

// GenerateFrrConfigurationFiles 生成 frr 配置文件
func (c *Constellation) GenerateFrrConfigurationFiles() error {
	if _, ok := c.systemInitSteps[GenerateFrrConfigurationFiles]; ok {
		constellationLogger.Infof("already generate frr configuration files")
	}

	for _, sat := range c.Satellites {
		err := sat.GenerateFrrConfig()
		if err != nil {
			return fmt.Errorf("generate frr configuration files failed: %w", err)
		}
	}
	c.systemInitSteps[GenerateFrrConfigurationFiles] = struct{}{}
	constellationLogger.Infof("generate frr configuration files")
	return nil
}

// GenerateLinksForNormalSatellite 为共识卫星生成链路
func (c *Constellation) generateLinksForConsensusSatellites() {
	for _, sat := range c.Satellites {
		satReal, _ := sat.ActualNode.(*satellite.ConsensusSatellite)
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat, _ := c.Satellites[targetSatId].ActualNode.(*satellite.ConsensusSatellite)
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)                                                         // 当前链路数量
			linkId := currentLinkNums + 1                                                                       // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink                                           // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                // 节点类型
			subNet := c.SubNets[currentLinkNums]                                                                // 获取当前子网
			sourceSat.ConnectedSubnetList = append(sourceSat.ConnectedSubnetList, subNet.String())              // 卫星添加子网
			targetSat.ConnectedSubnetList = append(targetSat.ConnectedSubnetList, subNet.String())              // 卫星添加子网
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceIntf, targetIntf,
				sourceSat, targetSat)
			sourceSat.Ifidx++                                                               // 接口索引变化
			targetSat.Ifidx++                                                               // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, intraOrbitLink)               // 添加到所有链路集合
			c.IntraOrbitSatelliteLinks = append(c.IntraOrbitSatelliteLinks, intraOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成同轨道的星间链路 ---------------->
		// <---------------- 生成异轨道的星间链路 ---------------->
		targetOrbitId = sourceOrbitId + 1
		if targetOrbitId < c.OrbitNumber {
			targetIndexInOrbit = sourceSat.IndexInOrbit
			targetSatId = targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
			targetSat, _ = c.Satellites[targetSatId].ActualNode.(*satellite.ConsensusSatellite)
			currentLinkNums := len(c.AllSatelliteLinks)                                                         // 当前链路数量
			linkId := currentLinkNums + 1                                                                       // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                           // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                // 节点类型
			subNet := c.SubNets[currentLinkNums]                                                                // 获取当前子网
			sourceSat.ConnectedSubnetList = append(sourceSat.ConnectedSubnetList, subNet.String())              // 源卫星添加子网
			targetSat.ConnectedSubnetList = append(targetSat.ConnectedSubnetList, subNet.String())              // 目的卫星添加子网
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceIntf, targetIntf,
				sourceSat, targetSat)
			sourceSat.Ifidx++                                                               // 接口索引变化
			targetSat.Ifidx++                                                               // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, interOrbitLink)               // 添加到所有链路集合
			c.InterOrbitSatelliteLinks = append(c.InterOrbitSatelliteLinks, interOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成异轨道的星间链路 ---------------->
	}
}

// GenerateLinksForNormalSatellite 为普通卫星生成链路
func (c *Constellation) generateLinksForNormalSatellite() {
	for _, sat := range c.Satellites {
		satReal, _ := sat.ActualNode.(*satellite.NormalSatellite)
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat, _ := c.Satellites[targetSatId].ActualNode.(*satellite.NormalSatellite)
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)                                                         // 当前链路数量
			linkId := currentLinkNums + 1                                                                       // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink                                           // 链路类型
			nodeType := types.NetworkNodeType_NormalSatellite                                                   // 节点类型
			subNet := c.SubNets[currentLinkNums]                                                                // 获取当前子网
			sourceSat.ConnectedSubnetList = append(sourceSat.ConnectedSubnetList, subNet.String())              // 卫星添加子网
			targetSat.ConnectedSubnetList = append(targetSat.ConnectedSubnetList, subNet.String())              // 卫星添加子网
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceIntf, targetIntf,
				sourceSat, targetSat)
			sourceSat.Ifidx++                                                               // 接口索引变化
			targetSat.Ifidx++                                                               // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, intraOrbitLink)               // 添加到所有链路集合
			c.IntraOrbitSatelliteLinks = append(c.IntraOrbitSatelliteLinks, intraOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成同轨道的星间链路 ---------------->
		// <---------------- 生成异轨道的星间链路 ---------------->
		targetOrbitId = sourceOrbitId + 1
		if targetOrbitId < c.OrbitNumber {
			targetIndexInOrbit = sourceSat.IndexInOrbit
			targetSatId = targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
			targetSat, _ = c.Satellites[targetSatId].ActualNode.(*satellite.NormalSatellite)
			currentLinkNums := len(c.AllSatelliteLinks)                                                         // 当前链路数量
			linkId := currentLinkNums + 1                                                                       // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                           // 链路类型
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                // 节点类型
			subNet := c.SubNets[currentLinkNums]                                                                // 获取当前子网
			sourceSat.ConnectedSubnetList = append(sourceSat.ConnectedSubnetList, subNet.String())              // 源卫星添加子网
			targetSat.ConnectedSubnetList = append(targetSat.ConnectedSubnetList, subNet.String())              // 目的卫星添加子网
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceIntf, targetIntf,
				sourceSat, targetSat)
			sourceSat.Ifidx++                                                               // 接口索引变化
			targetSat.Ifidx++                                                               // 接口索引变化
			c.AllSatelliteLinks = append(c.AllSatelliteLinks, interOrbitLink)               // 添加到所有链路集合
			c.InterOrbitSatelliteLinks = append(c.InterOrbitSatelliteLinks, interOrbitLink) // 添加到轨内链路集合
		}
		// <---------------- 生成异轨道的星间链路 ---------------->
	}
}
