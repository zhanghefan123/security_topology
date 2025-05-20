package constellation

import "C"
import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"zhanghefan123/security_topology/api/route"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/intf"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
	"zhanghefan123/security_topology/modules/utils/file"
	"zhanghefan123/security_topology/modules/utils/network"
	"zhanghefan123/security_topology/modules/utils/position"
)

const (
	SetHostReceiveMemory                  = "SetHostReceiveMemory"                  // 设置接收缓冲区
	GenerateSatellites                    = "GenerateSatellites"                    // 生成卫星
	GenerateGroundStations                = "GenerateGroundStations"                // 生成地面站
	GenerateSubnets                       = "GenerateIPv4Subnets"                   // 创建子网
	GenerateISLs                          = "GenerateISLs"                          // 生成星间链路
	GenerateGSLs                          = "GenerateGSLs"                          // 生成星地链路
	GenerateAddressMapping                = "GenerateAddressMapping"                // 生成容器名 -> 地址的映射
	GeneratePortMapping                   = "GeneratePortMapping"                   // 生成容器名 -> 端口的映射
	GenerateFrrConfigurationFiles         = "GenerateFrrConfigurationFiles"         // 生成 frr 配置
	CalculateAndWriteSegmentRoutes        = "CalculateAndWriteSegmentRoutes"        // 进行段路由的计算
	CalculateAndWriteLiRRoutes            = "CalculateAndWriteLiRRoutes"            // 进行 lir 路由的计算
	GenerateIfnameToLinkIdentifierMapping = "GenerateIfnameToLinkIdentifierMapping" // 生成从接口名称到 link identifier 的映射
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

// Init 进行初始化
func (c *Constellation) Init() error {

	// enableSRv6 := configs.TopConfiguration.NetworkConfig.EnableSRv6

	initSteps := []map[string]InitModule{
		{SetHostReceiveMemory: InitModule{false, c.SetHostReceiveMemory}},
		{GenerateSatellites: InitModule{true, c.GenerateSatellites}},
		{GenerateGroundStations: InitModule{true, c.GenerateGroundStations}},
		{GenerateSubnets: InitModule{true, c.GenerateSubnets}},
		{GenerateISLs: InitModule{true, c.GenerateISLs}},
		{GenerateGSLs: InitModule{true, c.GenerateGSLs}},
		{GenerateFrrConfigurationFiles: InitModule{true, c.GenerateFrrConfigurationFiles}},
		{GenerateAddressMapping: InitModule{true, c.GenerateAddressMapping}},
		{GeneratePortMapping: InitModule{true, c.GeneratePortMapping}},
		{CalculateAndWriteSegmentRoutes: InitModule{true, c.CalculateAndWriteSegmentRoutes}},
		{CalculateAndWriteLiRRoutes: InitModule{true, c.CalculateAndWriteLiRRoutes}},
		{GenerateIfnameToLinkIdentifierMapping: InitModule{true, c.GenerateIfnameToLinkIdentifierMapping}},
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

// SetHostReceiveMemory 进行主机的接收缓存的设置
func (c *Constellation) SetHostReceiveMemory() error {
	if _, ok := c.systemInitSteps[SetHostReceiveMemory]; ok {
		constellationLogger.Infof("already set host receive memory")
		return nil
	}

	err := execute.Command("sysctl", strings.Split("-w net.core.rmem_max=16777216", " "))
	if err != nil {
		return fmt.Errorf("set host receive memory failed %w", err)
	}

	err = execute.Command("sysctl", strings.Split("-w net.core.rmem_default=16777216", " "))
	if err != nil {
		return fmt.Errorf("set host receive memory failed %w", err)
	}

	c.systemInitSteps[GenerateSatellites] = struct{}{}
	constellationLogger.Infof("set host receive memory")
	return nil
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
	freq := 1 / 0.0695
	for orbitId := 0; orbitId < c.OrbitNumber; orbitId++ {
		orbitStartLatitude := startLatitude + delta
		orbitLongitude := startLongitude + 180*float32(orbitId)/float32(c.OrbitNumber)
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
				c.SatelliteAbstractNodes = append(c.SatelliteAbstractNodes, normalSatelliteAbstract)
			} else if c.SatelliteType == types.NetworkNodeType_LiRSatellite { // 2. 如果是生成 lir 卫星
				// 创建lir卫星
				lirSatellite := satellites.NewLiRSatellite(nodeId+1, orbitId, indexInOrbit, tle)
				// 添加卫星
				c.LiRSatellites = append(c.LiRSatellites, lirSatellite)
				// 创建抽象节点
				lirSatelliteAbstract := node.NewAbstractNode(lirSatellite.Type, lirSatellite, ConstellationInstance.ConstellationGraph)
				//  将 satellite 放到 allAbstractNodes 之中
				c.SatelliteAbstractNodes = append(c.SatelliteAbstractNodes, lirSatelliteAbstract)
			} else {
				return fmt.Errorf("not supported network node type")
			}
		}
	}

	c.systemInitSteps[GenerateSatellites] = struct{}{}
	constellationLogger.Infof("generate satellites")

	return nil
}

// GenerateGroundStations 进行地面站的生成
func (c *Constellation) GenerateGroundStations() error {
	if _, ok := c.systemInitSteps[GenerateGroundStations]; ok {
		constellationLogger.Infof("already generate groundstations")
		return nil
	}

	// 进行所有的地面站的遍历进行节点的添加
	// ------------------------------------------------------------------------------------------
	for index, groundStationParam := range c.Parameters.GroundStationsParams {
		// 进行地面站实例的创建
		groundStationInstance := ground_station.NewGroundStation(index+1,
			groundStationParam.Longitude,
			groundStationParam.Latitude,
			groundStationParam.Name)
		// 向地面站列表之中进行添加
		c.GroundStations = append(c.GroundStations, groundStationInstance)
		// 创建抽象节点
		groundStationAbstract := node.NewAbstractNode(groundStationInstance.Type, groundStationInstance, ConstellationInstance.ConstellationGraph)
		// 将节点存放到 abstractNodes 之中
		c.GroundStationAbstractNodes = append(c.GroundStationAbstractNodes, groundStationAbstract)
	}
	// ------------------------------------------------------------------------------------------

	c.systemInitSteps[GenerateGroundStations] = struct{}{}
	constellationLogger.Infof("generate ground stations")
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

// GenerateAddressMapping 生成地址映射
func (c *Constellation) GenerateAddressMapping() error {
	if _, ok := c.systemInitSteps[GenerateAddressMapping]; ok {
		constellationLogger.Infof("already generate address mapping")
		return nil
	}

	addressMapping, err := c.GetContainerNameToAddressMapping()
	if err != nil {
		fmt.Printf("generate address mapping: %v", err)
		return fmt.Errorf("generate address mapping: %w", err)
	}

	idMapping, err := c.GetContainerNameToGraphIdMapping()
	if err != nil {
		fmt.Printf("generate address mapping: %v", err)
		return fmt.Errorf("generate address mapping: %w", err)
	}

	finalString := ""
	for containerName, ipv4andipv6 := range addressMapping {
		graphId := idMapping[containerName]
		finalString += fmt.Sprintf("%s->%d->%s->%s\n", containerName, graphId, ipv4andipv6[0], ipv4andipv6[1])
	}

	// 进行所有节点的变量将 finalString 进行存储
	allAbstractNodes := append(c.SatelliteAbstractNodes, c.GroundStationAbstractNodes...)
	for _, abstractNode := range allAbstractNodes {
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
			fmt.Printf("error create file: %v", err)
			return fmt.Errorf("error create file %v", err)
		}
		_, err = f.WriteString(finalString)
		if err != nil {
			fmt.Printf("error write file: %v", err)
			return fmt.Errorf("error write file %w", err)
		}
	}

	c.systemInitSteps[GenerateAddressMapping] = struct{}{}
	constellationLogger.Infof("generate address mapping")
	return nil
}

// GeneratePortMapping 生成端口映射
func (c *Constellation) GeneratePortMapping() error {
	if _, ok := c.systemInitSteps[GeneratePortMapping]; ok {
		constellationLogger.Infof("already generate port mapping")
		return nil
	}

	c.systemInitSteps[GeneratePortMapping] = struct{}{}
	constellationLogger.Infof("generate port mapping")
	return nil
}

// GenerateGSLs 进行星地链路的生成
func (c *Constellation) GenerateGSLs() error {
	if _, ok := c.systemInitSteps[GenerateGSLs]; ok {
		constellationLogger.Infof("already generate GSLs")
		return nil
	}

	// 主要逻辑
	// ------------------------------------------------------------
	// 进行所有的地面站的遍历
	currentISLsNum := len(c.AllSatelliteLinks)
	for _, groundStation := range c.GroundStations {
		linkType := types.NetworkLinkType_GroundSatelliteLink
		sourceNodeType := types.NetworkNodeType_GroundStation
		targetNodeType := c.SatelliteType
		groundAbstractNode := c.GroundStationAbstractNodes[groundStation.Id-1]
		ipv4SubNet := c.Ipv4SubNets[currentISLsNum+len(c.AllGroundSatelliteLinks)]
		ipv6SubNet := c.Ipv6SubNets[currentISLsNum+len(c.AllGroundSatelliteLinks)]
		groundStation.ConnectedIpv4SubnetList = append(groundStation.ConnectedIpv4SubnetList, ipv4SubNet.String())
		groundStation.ConnectedIpv6SubnetList = append(groundStation.ConnectedIpv6SubnetList, ipv6SubNet.String())
		groundIpv4Addr, satelliteIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet) // 提取ipv4第一个和第二个地址
		groundIpv6Addr, satelliteIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)
		groundIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(groundStation.Type), groundStation.Id, groundStation.Ifidx)
		groundIntf := intf.NewNetworkInterface(groundStation.Ifidx, groundIfName, groundIpv4Addr, groundIpv6Addr, satelliteIpv4Addr, satelliteIpv6Addr, -1, nil) // 创建地面站接口
		groundStation.IfNameToInterfaceMap[groundIfName] = groundIntf
		groundStation.Interfaces = append(groundStation.Interfaces, groundIntf)
		// 注意下面的这个 link (targetNodeID, targetContainerName, sourceInterface, targetInterface, targetAbstractNode) 没有填充
		abstractGSL := link.NewAbstractLink(linkType,
			groundStation.Id, sourceNodeType, targetNodeType,
			groundStation.Id, -1,
			groundStation.ContainerName, "",
			groundIntf, &intf.NetworkInterface{},
			groundAbstractNode, nil,
			configs.TopConfiguration.ConstellationConfig.GSLBandwidth,
			ConstellationInstance.ConstellationGraph,
			ipv4SubNet,
			ipv6SubNet,
		)
		c.AllGroundSatelliteLinks = append(c.AllGroundSatelliteLinks, abstractGSL)
		c.AllGroundSatelliteLinksMap[groundStation.ContainerName] = abstractGSL
	}
	// ------------------------------------------------------------

	c.systemInitSteps[GenerateGSLs] = struct{}{}
	constellationLogger.Infof("generate GSLs")
	return nil
}

// GenerateISLs 进行星间链路的生成
func (c *Constellation) GenerateISLs() error {
	if _, ok := c.systemInitSteps[GenerateISLs]; ok {
		constellationLogger.Infof("already generate ISLs")
		return nil
	}

	if c.SatelliteType == types.NetworkNodeType_NormalSatellite {
		c.generateISLsForNormalSatellites()
	} else if c.SatelliteType == types.NetworkNodeType_LiRSatellite {
		c.generateISLsForLiRSatellite()
	} else {
		return fmt.Errorf("not supported network node type")
	}

	c.systemInitSteps[GenerateISLs] = struct{}{}
	constellationLogger.Infof("generate ISLs")
	return nil
}

// generateISLsForLiRSatellite 为 lir 卫星生成链路
func (c *Constellation) generateISLsForLiRSatellite() {
	satelliteNodeType := types.NetworkNodeType_LiRSatellite
	for index, sat := range c.LiRSatellites {
		satReal := sat
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceAbstract := c.SatelliteAbstractNodes[index]
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat := c.LiRSatellites[targetSatId] // 目的节点的抽象标识
		targetAbstract := c.SatelliteAbstractNodes[targetSatId]
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)               // 当前链路数量
			linkId := currentLinkNums + 1                             // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink // 链路类型
			nodeType := satelliteNodeType
			// link 相关的内容
			// ------------------------------------------------------------------------------------------------
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                                                                    // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                                                                    // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                                                            // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                                                            // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                                                             // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                                                             // 目的接口名
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces, nil) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces, nil) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                       // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                                                                 // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                       // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)
			// ------------------------------------------------------------------------------------------------
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph,
				ipv4SubNet,
				ipv6SubNet)
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
			targetSat = c.LiRSatellites[targetSatId]
			targetAbstract = c.SatelliteAbstractNodes[targetSatId]
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                                                     // 当前链路数量
			linkId := currentLinkNums + 1                                                                                                                                   // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                                                                                       // 链路类型
			nodeType := satelliteNodeType                                                                                                                                   // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                                                                    // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                                                                    // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                                                            // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                                                            // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                                                             // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                                                             // 目的接口名
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces, nil) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces, nil) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                       // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                                                                 // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                       // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                                                                 // 添加接口
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph,
				ipv4SubNet,
				ipv6SubNet)
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

// generateISLsForNormalSatellites 为普通卫星生成链路
func (c *Constellation) generateISLsForNormalSatellites() {
	satelliteNodeType := types.NetworkNodeType_NormalSatellite
	for index, sat := range c.NormalSatellites {
		satReal := sat
		// <---------------- 生成同轨道的星间链路 ---------------->
		sourceSat := satReal
		sourceAbstract := c.SatelliteAbstractNodes[index]
		sourceOrbitId := sourceSat.OrbitId
		targetOrbitId := sourceOrbitId
		targetIndexInOrbit := (sourceSat.IndexInOrbit + 1) % c.SatellitePerOrbit
		targetSatId := targetOrbitId*c.SatellitePerOrbit + targetIndexInOrbit
		targetSat := c.NormalSatellites[targetSatId] // 目的节点的抽象标识
		targetAbstract := c.SatelliteAbstractNodes[targetSatId]
		if reflect.DeepEqual(sourceSat, targetSat) {
			continue
		} else {
			currentLinkNums := len(c.AllSatelliteLinks)               // 当前链路数量
			linkId := currentLinkNums + 1                             // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_IntraOrbitSatelliteLink // 链路类型
			nodeType := satelliteNodeType
			// link 相关的内容
			// ------------------------------------------------------------------------------------------------
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                                                                    // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                                                                    // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                                                            // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                                                            // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                                                             // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                                                             // 目的接口名
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces, nil) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces, nil) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                       // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                                                                 // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                       // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)
			// ------------------------------------------------------------------------------------------------
			intraOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph,
				ipv4SubNet,
				ipv6SubNet)
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
			targetAbstract = c.SatelliteAbstractNodes[targetSatId]
			currentLinkNums := len(c.AllSatelliteLinks)                                                                                                                     // 当前链路数量
			linkId := currentLinkNums + 1                                                                                                                                   // 当前链路数量 + 1 -> 链路 id
			linkType := types.NetworkLinkType_InterOrbitSatelliteLink                                                                                                       // 链路类型
			nodeType := satelliteNodeType                                                                                                                                   // 节点类型
			ipv4SubNet := c.Ipv4SubNets[currentLinkNums]                                                                                                                    // 获取当前ipv4 子网
			ipv6SubNet := c.Ipv6SubNets[currentLinkNums]                                                                                                                    // 获取当前 ipv6 子网
			sourceSat.ConnectedIpv4SubnetList = append(sourceSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			targetSat.ConnectedIpv4SubnetList = append(targetSat.ConnectedIpv4SubnetList, ipv4SubNet.String())                                                              // 卫星添加ipv4子网
			sourceSat.ConnectedIpv6SubnetList = append(sourceSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			targetSat.ConnectedIpv6SubnetList = append(targetSat.ConnectedIpv6SubnetList, ipv6SubNet.String())                                                              // 卫星添加ipv6子网
			sourceIpv4Addr, targetIpv4Addr := network.GenerateTwoAddrsFromIpv4Subnet(ipv4SubNet)                                                                            // 提取ipv4第一个和第二个地址
			sourceIpv6Addr, targetIpv6Addr := network.GenerateTwoAddrsFromIpv6Subnet(ipv6SubNet)                                                                            // 提取ipv6第一个和第二个地址
			sourceIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), sourceSat.Id, sourceSat.Ifidx)                                                             // 源接口名
			targetIfName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(nodeType), targetSat.Id, targetSat.Ifidx)                                                             // 目的接口名
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			sourceIntf := intf.NewNetworkInterface(sourceSat.Ifidx, sourceIfName, sourceIpv4Addr, sourceIpv6Addr, targetIpv4Addr, targetIpv6Addr, c.NetworkInterfaces, nil) // 创建第一个接口
			c.NetworkInterfaces += 1                                                                                                                                        // 接口数量 ++
			targetIntf := intf.NewNetworkInterface(targetSat.Ifidx, targetIfName, targetIpv4Addr, targetIpv6Addr, sourceIpv4Addr, sourceIpv6Addr, c.NetworkInterfaces, nil) // 创建第二个接口
			sourceSat.IfNameToInterfaceMap[sourceIfName] = sourceIntf                                                                                                       // 设置源卫星地址
			sourceSat.Interfaces = append(sourceSat.Interfaces, sourceIntf)                                                                                                 // 添加接口
			targetSat.IfNameToInterfaceMap[targetIfName] = targetIntf                                                                                                       // 设置目的卫星地址
			targetSat.Interfaces = append(targetSat.Interfaces, targetIntf)                                                                                                 // 添加接口
			interOrbitLink := link.NewAbstractLink(linkType, linkId,
				nodeType, nodeType,
				sourceSat.Id, targetSat.Id,
				sourceSat.ContainerName, targetSat.ContainerName,
				sourceIntf, targetIntf,
				sourceAbstract, targetAbstract, configs.TopConfiguration.ConstellationConfig.ISLBandwidth,
				ConstellationInstance.ConstellationGraph,
				ipv4SubNet,
				ipv6SubNet)
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

	// 将两个切片进行拼接
	satelliteAndGroundStations := append(c.SatelliteAbstractNodes, c.GroundStationAbstractNodes...)

	// 进行所有的遍历
	for index, abstractNode := range satelliteAndGroundStations {
		routerId := index + 1
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("cannot convert abstract node to normal node")
		}
		ospfVersion := configs.TopConfiguration.NetworkConfig.OspfVersion
		if ospfVersion == "ospfv2" {
			// 生成 ospfv2 配置
			err = normalNode.GenerateOspfV2FrrConfig(routerId)
			if err != nil {
				return fmt.Errorf("generate ospfv2 frr configuration files failed: %w", err)
			}
		} else if ospfVersion == "ospfv3" {
			// 生成 ospfv3 配置
			err = normalNode.GenerateOspfV3FrrConfig(routerId)
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

// CalculateAndWriteSegmentRoutes 进行段路由的计算
func (c *Constellation) CalculateAndWriteSegmentRoutes() error {
	if _, ok := c.systemInitSteps[CalculateAndWriteSegmentRoutes]; ok {
		constellationLogger.Infof("already calculate segment routes")
		return nil
	}

	for _, abstractNode := range c.SatelliteAbstractNodes {
		err := route.CalculateAndWriteSegmentRoute(abstractNode, &(c.AllSatelliteLinksMap), ConstellationInstance.ConstellationGraph)
		if err != nil {
			return fmt.Errorf("calculate route failed: %w", err)
		}
	}

	c.systemInitSteps[CalculateAndWriteSegmentRoutes] = struct{}{}
	constellationLogger.Infof("calculate segment routes")
	return nil
}

// CalculateAndWriteLiRRoutes 计算 LiR 路由
func (c *Constellation) CalculateAndWriteLiRRoutes() error {
	if _, ok := c.systemInitSteps[CalculateAndWriteLiRRoutes]; ok {
		constellationLogger.Infof("already calculate lir routes")
		return nil
	}

	// simulation 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath

	// 所有的节点的 LiR 路由
	allLiRRoutes := make([]string, 0)

	// 遍历所有节点生成路由文件
	for _, abstractNode := range c.SatelliteAbstractNodes {
		// 获取单个节点的路由条目集合
		lirRoute, err := route.GenerateLiRRoute(abstractNode, &(c.AllSatelliteLinksMap), ConstellationInstance.ConstellationGraph)
		if err != nil {
			return fmt.Errorf("generate path_validation route failed: %w", err)
		}
		// 更新总路由条目
		allLiRRoutes = append(allLiRRoutes, lirRoute)
	}

	allLirRoutesString := strings.Join(allLiRRoutes, "\n")

	// 准备写入文件
	for index, abstractNode := range c.SatelliteAbstractNodes {
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

	c.systemInitSteps[CalculateAndWriteLiRRoutes] = struct{}{}
	constellationLogger.Infof("calculate lir routes")
	return nil
}

func (c *Constellation) GenerateIfnameToLinkIdentifierMapping() error {
	if _, ok := c.systemInitSteps[GenerateIfnameToLinkIdentifierMapping]; ok {
		constellationLogger.Infof("already generate ifname to link identifier")
		return nil
	}

	// 遍历所有的节点生成 mapping
	for _, abstractNode := range c.SatelliteAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("generate ifname to link identifier mapping files failed, %w", err)
		}
		err = normalNode.GenerateIfnameToLidMapping()
		if err != nil {
			return fmt.Errorf("generate ifname to link identifier mapping files failed, %w", err)
		}
	}

	c.systemInitSteps[GenerateIfnameToLinkIdentifierMapping] = struct{}{}
	constellationLogger.Infof("generate ifname to link identifier")
	return nil
}
