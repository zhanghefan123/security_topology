package constellation

import (
	"errors"
	"fmt"
	"reflect"
	"zhanghefan123/security_topology/modules/config/system"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/subnet"
)

const (
	GenerateSatellites = "GenerateSatellites"
	GenerateSubnets    = "GenerateSubnets"
	GenerateLinks      = "GenerateLinks"
)

type InitFunction func() error

var (
	ErrNotSupportedNetworkNodeType = errors.New("ErrNotSupportedNetworkNodeType")
)

// Init 进行初始化
func (c *Constellation) Init() {
	initSteps := []map[string]InitFunction{
		{GenerateSatellites: c.GenerateSatellites},
		{GenerateSubnets: c.GenerateSubnets},
		{GenerateLinks: c.GenerateLinks},
	}
	err := c.initializeSteps(initSteps)
	if err != nil {
		moduleConstellationLogger.Error("constellation init failed: %v", err)
	}
}

// InitializeSteps 按步骤进行初始化
func (c *Constellation) initializeSteps(initSteps []map[string]InitFunction) (err error) {
	moduleNum := len(initSteps)
	for idx, initStep := range initSteps {
		for name, initFunc := range initStep {
			if err := initFunc(); err != nil {
				moduleConstellationLogger.Errorf("init step [%s] failed, %s", name, err)
				return err
			}
			moduleConstellationLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

// GenerateSatellites 生成卫星
func (c *Constellation) GenerateSatellites() error {
	if _, ok := c.initModules[GenerateSatellites]; ok {
		moduleConstellationLogger.Infof("already generate satellites")
		return nil
	}

	for orbitId := 0; orbitId < c.OrbitNumber; orbitId++ {
		for nodeId := c.SatellitePerOrbit * orbitId; nodeId < c.SatellitePerOrbit*(orbitId+1); nodeId++ {
			indexInOrbit := nodeId % c.SatellitePerOrbit
			// 判断该进行什么卫星的生成
			if c.SatelliteType == types.NetworkNodeType_NormalSatellite {
				sat := satellite.NewNormalSatellite(nodeId+1, orbitId,
					indexInOrbit, c.SatelliteImageName)
				c.Satellites = append(c.Satellites, sat)
			} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite {
				sat := satellite.NewConsensusSatellite(nodeId+1, orbitId,
					indexInOrbit, c.SatelliteImageName,
					c.SatelliteRPCPort, c.SatelliteP2PPort)
				c.Satellites = append(c.Satellites, sat)
			} else {
				moduleConstellationLogger.Errorf("unsupported network node type")
				return ErrNotSupportedNetworkNodeType
			}
		}
	}

	c.initModules[GenerateSatellites] = struct{}{}
	moduleConstellationLogger.Infof("generate satellites")

	return nil
}

// GenerateSubnets 进行子网的生成
func (c *Constellation) GenerateSubnets() error {
	if _, ok := c.initModules[GenerateSubnets]; ok {
		moduleConstellationLogger.Infof("already generate subnets")
	}

	subNets := subnet.GenerateSubnets(system.TopConfiguration.NetworkConfig.BaseNetworkAddress)
	c.SubNets = subNets

	c.initModules[GenerateSubnets] = struct{}{}
	moduleConstellationLogger.Infof("generate subnets")
	return nil
}

// GenerateLinks 进行链路的生成
func (c *Constellation) GenerateLinks() error {
	if _, ok := c.initModules[GenerateLinks]; ok {
		moduleConstellationLogger.Infof("already generate links")
	}

	if c.SatelliteType == types.NetworkNodeType_NormalSatellite {
		c.generateLinksForNormalSatellite()
	} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite {
		c.generateLinksForConsensusSatellites()
	} else {
		return ErrNotSupportedNetworkNodeType
	}

	c.initModules[GenerateLinks] = struct{}{}
	moduleConstellationLogger.Infof("generate links")
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
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
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
			nodeType := types.NetworkNodeType_ConsensusSatellite                                                // 节点类型
			subNet := c.SubNets[currentLinkNums]                                                                // 获取当前子网
			sourceSat.ConnectedSubnetList = append(sourceSat.ConnectedSubnetList, subNet.String())              // 卫星添加子网
			sourceAddr, targetAddr := subnet.GenerateTwoAddrsFrom30MaskSubnet(subNet)                           // 提取第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx) // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx) // 目的接口名
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceAddr)                   // 创建第一个接口
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetAddr)                   // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                           // 设置源卫星地址
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                           // 设置目的卫星地址
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
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
