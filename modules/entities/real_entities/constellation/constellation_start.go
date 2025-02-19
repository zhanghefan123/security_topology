package constellation

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"strconv"
	"strings"
	"sync"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
	"zhanghefan123/security_topology/modules/entities/real_entities/position_info"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	posPbLink "zhanghefan123/security_topology/services/position/protobuf/link"
	posPbNode "zhanghefan123/security_topology/services/position/protobuf/node"
)

const (
	GenerateSatelliteLinks       = "GenerateSatelliteLinks"
	StartSatelliteContainers     = "StartSatelliteContainers"
	StartGroundStationContainers = "StartGroundStationContainers"
	SetVethNameSpaces            = "SetVethNameSpaces"
	StartEtcdService             = "StartEtcdService"
	StoreToEtcd                  = "StoreToEtcd"
	CreateServiceContext         = "CreateServiceContext"
	StartPositionService         = "StartPositionService"
	StartUpdateDelayService      = "StartUpdateDelayService"
)

type StartFunction func() error

type StartModule struct {
	start         bool          // 是否启动
	startFunction StartFunction // 相应的启动函数
}

// Start 启动
func (c *Constellation) Start() error {
	enablePositionService := configs.TopConfiguration.ServicesConfig.PositionUpdateConfig.Enabled
	enableUpdatedDelayService := configs.TopConfiguration.ServicesConfig.DelayUpdateConfig.Enabled

	startSteps := []map[string]StartModule{
		{GenerateSatelliteLinks: StartModule{true, c.GenerateSatelliteVethPairs}},                    // step1 先创建 veth pair 然后改变链路的命名空间
		{StartSatelliteContainers: StartModule{true, c.StartSatelliteContainers}},                    // step2 一定要在 step1 之后，因为创建了容器后才有命名空间
		{StartGroundStationContainers: StartModule{true, c.StartGroundStationContainers}},            // step3 进行地面站容器的创建
		{SetVethNameSpaces: StartModule{true, c.SetVethNamespaces}},                                  // step4 一定要在 step2 之后，因为创建了容器才能设置 veth 的 namespace
		{StartEtcdService: StartModule{true, c.StartEtcdService}},                                    // step5 进行 etcd 服务的启动
		{StoreToEtcd: StartModule{true, c.StoreToEtcd}},                                              // step6 一定要在 step4 之后，因为创建了 etcd 服务才能进行存
		{CreateServiceContext: StartModule{true, c.CreateServiceContext}},                            // step7 创建服务上下文
		{StartPositionService: StartModule{enablePositionService, c.StartPositionService}},           // step8 启动位置 position
		{StartUpdateDelayService: StartModule{enableUpdatedDelayService, c.StartUpdateDelayService}}, // step9 开启监听的服务
	}
	err := c.startSteps(startSteps)
	if err != nil {
		return fmt.Errorf("constellation start error: %w", err)
	}
	return nil
}

// startStepsNum 获取启动的模块的数量
func (c *Constellation) startStepsNum(startSteps []map[string]StartModule) int {
	result := 0
	for _, startStep := range startSteps {
		for _, startModule := range startStep {
			if startModule.start {
				result += 1
			}
		}
	}
	return result
}

// startSteps 调用所有的启动方法
func (c *Constellation) startSteps(startSteps []map[string]StartModule) (err error) {
	moduleNum := c.startStepsNum(startSteps)
	for idx, startStep := range startSteps {
		for name, startModule := range startStep {
			// 判断是否需要进行启动, 如果要进行启动，再调用
			if startModule.start {
				if err = startModule.startFunction(); err != nil {
					constellationLogger.Errorf("start step [%s] failed, %s", name, err)
					return err
				}
				constellationLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	return
}

// GenerateSatelliteVethPairs 进行卫星之间的 veth pairs 的生成
func (c *Constellation) GenerateSatelliteVethPairs() error {
	if _, ok := c.systemStartSteps[GenerateSatelliteLinks]; ok {
		constellationLogger.Infof("GenerateSatelliteVethPairs is already running")
		return nil
	}
	description := fmt.Sprintf("%20s", "generate veth pairs")
	var taskFunc multithread.TaskFunc[*link.AbstractLink] = func(link *link.AbstractLink) error {
		err := link.GenerateVethPair()
		if err != nil {
			return err
		}
		return nil
	}

	c.systemStartSteps[GenerateSatelliteLinks] = struct{}{}
	constellationLogger.Infof("generate satellite veth pairs")

	return multithread.RunInMultiThread[*link.AbstractLink](description, taskFunc, c.AllSatelliteLinks)
}

// StartSatelliteContainers 启动卫星容器
func (c *Constellation) StartSatelliteContainers() error {
	if _, ok := c.systemStartSteps[StartSatelliteContainers]; ok {
		constellationLogger.Infof("StartSatelliteContainers is already running")
		return nil
	}
	description := fmt.Sprintf("%20s", "start containers")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.CreateContainer(c.client, node)
		if err != nil {
			return err
		}
		err = container_api.StartContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	c.systemStartSteps[StartSatelliteContainers] = struct{}{}
	constellationLogger.Infof("execute start containers")

	return multithread.RunInMultiThread(description, taskFunc, c.SatelliteAbstractNodes)
}

// StartGroundStationContainers 启动地面站容器
func (c *Constellation) StartGroundStationContainers() error {
	if _, ok := c.systemInitSteps[StartGroundStationContainers]; ok {
		constellationLogger.Infof("StartGroundStationContainers is already running")
		return nil
	}
	description := fmt.Sprintf("%20s", "start ground station containers")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.CreateContainer(c.client, node)
		if err != nil {
			return err
		}
		err = container_api.StartContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}
	c.systemInitSteps[StartGroundStationContainers] = struct{}{}
	constellationLogger.Infof("execute start ground station containers")
	return multithread.RunInMultiThread(description, taskFunc, c.GroundStationAbstractNodes)
}

// SetVethNamespaces 设置 veth 命名空间
func (c *Constellation) SetVethNamespaces() error {
	if _, ok := c.systemStartSteps[SetVethNameSpaces]; ok {
		constellationLogger.Infof("SetVethNameSpaces is already running")
		return nil
	}
	description := fmt.Sprintf("%20s", "set veth namespaces")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		normalNode, err := node.GetNormalNodeFromAbstractNode()
		if err != nil {
			return err
		}
		// 进行 veth 命名空间的设置
		err = normalNode.SetVethNamespace()
		if err != nil {
			return err
		}
		return nil
	}

	c.systemStartSteps[SetVethNameSpaces] = struct{}{}
	constellationLogger.Infof("execute set veth namespaces")

	return multithread.RunInMultiThread(description, taskFunc, c.SatelliteAbstractNodes)
}

// StartEtcdService 开启 etcd 服务
func (c *Constellation) StartEtcdService() error {
	if _, ok := c.systemStartSteps[StartEtcdService]; ok {
		constellationLogger.Infof("StartEtcdService is already running")
		return nil
	}

	// 1. 解析配置
	etcdConfig := configs.TopConfiguration.ServicesConfig.EtcdConfig
	clientPort := etcdConfig.ClientPort
	peerPort := etcdConfig.PeerPort
	dataDir := etcdConfig.DataDir
	etcdName := etcdConfig.EtcdName

	// 2. 根据配置创建节点
	etcdService := etcd.NewEtcdNode(types.NetworkNodeStatus_Logic, clientPort, peerPort, dataDir, etcdName)

	// 3. 配置
	c.etcdService = etcdService

	// 4. 创建抽象节点
	c.abstractEtcdService = node.NewAbstractNode(types.NetworkNodeType_EtcdService, c.etcdService, nil)

	// 5. 进行容器的创建和启动
	err := container_api.CreateContainer(c.client, c.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("create etcd container failed, %s", err.Error())
	}
	err = container_api.StartContainer(c.client, c.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("start etcd container failed, %s", err.Error())
	}

	c.systemStartSteps[StartEtcdService] = struct{}{}
	constellationLogger.Infof("execute start etcd service")

	return nil
}

// StoreToEtcd 将卫星/链路信息放到 Etcd 之中
func (c *Constellation) StoreToEtcd() (err error) {
	if _, ok := c.systemStartSteps[StoreToEtcd]; ok {
		constellationLogger.Infof("StoreToEtcd is already running")
		return nil
	}

	waitGroup := sync.WaitGroup{}
	// 存在两个子任务
	waitGroup.Add(3)
	// 第一个子任务: 进行星座链路的存储
	go func() {
		defer waitGroup.Done() // 最终记录一下任务做完了
		for _, satelliteLink := range c.AllSatelliteLinks {
			err = satelliteLink.StoreToEtcd(c.etcdClient)
			if err != nil {
				err = fmt.Errorf("store ISL to etcd error %w", err)
				return
			}
		}
	}()
	// 第二个子任务: 进行卫星节点的存储
	go func() {
		// 最终记录一下任务做完了
		defer waitGroup.Done()
		for _, absNode := range c.SatelliteAbstractNodes {
			if absNode.Type == types.NetworkNodeType_NormalSatellite {
				// 如果节点为普通卫星
				normalSat, _ := absNode.ActualNode.(*satellites.NormalSatellite)
				err = normalSat.StoreToEtcd(c.etcdClient)
			} else if absNode.Type == types.NetworkNodeType_ConsensusSatellite {
				// 如果节点为共识卫星
				consensusSat, _ := absNode.ActualNode.(*satellites.ConsensusSatellite)
				err = consensusSat.StoreToEtcd(c.etcdClient)
			} else {
				err = fmt.Errorf("unsupported node type")
			}
			if err != nil {
				err = fmt.Errorf("store ISL to etcd error %w", err)
				return
			}
		}
	}()
	// 第三个子任务: 进行地面站节点的存储
	go func() {
		defer waitGroup.Done()
		for _, absNode := range c.GroundStationAbstractNodes {
			if absNode.Type == types.NetworkNodeType_GroundStation {
				groundStation, _ := absNode.ActualNode.(*ground_station.GroundStation)
				err = groundStation.StoreToEtcd(c.etcdClient)
			} else {
				err = fmt.Errorf("unsupported node type")
			}
			if err != nil {
				err = fmt.Errorf("store GroundStation to etcd error %w", err)
				return
			}
		}
	}()
	waitGroup.Wait()

	c.systemStartSteps[StoreToEtcd] = struct{}{}
	constellationLogger.Infof("execute store to etcd")

	return nil
}

// CreateServiceContext 创建程序上下文
func (c *Constellation) CreateServiceContext() (err error) {
	if _, ok := c.systemStartSteps[CreateServiceContext]; ok {
		constellationLogger.Infof("CreateServiceContext is already running")
		return nil
	}
	c.serviceContext, c.serviceContextCancelFunc = context.WithCancel(context.Background())

	c.systemStartSteps[CreateServiceContext] = struct{}{}
	constellationLogger.Infof("execute create context")
	return nil
}

// StartPositionService 启动位置服务
func (c *Constellation) StartPositionService() error {
	if _, ok := c.systemStartSteps[StartPositionService]; ok {
		constellationLogger.Infof("StartPositionService is already running")
		return nil
	}

	// 1. 解析配置
	etcdConfig := configs.TopConfiguration.ServicesConfig.EtcdConfig                        // etcd 配置
	etcdListenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress            // etcd 监听地址
	etcdClientPort := etcdConfig.ClientPort                                                 // etcd 客户端口
	etcdISLsPrefix := etcdConfig.EtcdPrefix.ISLsPrefix                                      // etcd isl 的前缀
	etcdGSLsPrefix := etcdConfig.EtcdPrefix.GSLsPrefix                                      // etcd gsl 的前缀
	etcdSatellitesPrefix := etcdConfig.EtcdPrefix.SatellitesPrefix                          // etcd satellite 的前缀
	etcdGroundStationsPrefix := etcdConfig.EtcdPrefix.GroundStationsPrefix                  // etcd GroundStations 的前缀
	constellationStartTime := configs.TopConfiguration.ConstellationConfig.StartTime        // 星座启动时间
	updateInterval := configs.TopConfiguration.ServicesConfig.PositionUpdateConfig.Interval // 更新时间间隔

	// 2. 根据配置创建节点
	positionService := position.NewPositionService(types.NetworkNodeStatus_Logic,
		etcdListenAddr, etcdClientPort,
		etcdISLsPrefix, etcdGSLsPrefix, etcdSatellitesPrefix, etcdGroundStationsPrefix,
		constellationStartTime, updateInterval)

	// 3. 创建抽象节点
	c.positionService = positionService

	c.abstractPositionService = node.NewAbstractNode(types.NetworkNodeType_PositionService, positionService, nil)

	// 4. 进行容器的创建
	err := container_api.CreateContainer(c.client, c.abstractPositionService)
	if err != nil {
		return fmt.Errorf("create position service failed, %s", err)
	}

	// 5. 进行容器的启动
	err = container_api.StartContainer(c.client, c.abstractPositionService)
	if err != nil {
		return fmt.Errorf("start position service failed, %s", err)
	}

	// 6. 进行启动
	c.systemStartSteps[StartPositionService] = struct{}{}
	constellationLogger.Infof("execute start position service")

	return nil
}

// StartUpdateDelayService 开启更新服务
func (c *Constellation) StartUpdateDelayService() error {
	if _, ok := c.systemStartSteps[StartUpdateDelayService]; ok {
		constellationLogger.Infof("StartUpdateDelayService is already running")
		return nil
	}
	// 由于地面站位置始终不变, 所以不用进行周期性的更新, 直接设置就可以了
	for _, groundStation := range c.GroundStations {
		c.ContainerNameToPosition[groundStation.ContainerName] = &position_info.Position{
			NodeType:  types.NetworkNodeType_GroundStation.String(), // 节点类型
			Latitude:  float64(groundStation.Latitude),              // 纬度
			Longitude: float64(groundStation.Longitude),             // 经度
			Altitude:  0,                                            // 高度
		}
	}

	// 开启一个线程, 不断进行更新事件的获取 (gsls 更新事件)
	go func() {
		// 创建一个监听键值对更新事件的 channel
		watchChan := c.etcdClient.Watch(
			c.serviceContext,
			configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.GSLsPrefix,
			clientv3.WithPrefix(),
		)
		for response := range watchChan {
			for _, event := range response.Events {
				go func() {
					// 创建 protobuf Node
					pbGSL := &posPbLink.Link{}
					// 将 etcd 的值进行反序列化
					protobuf.MustUnmarshal(event.Kv.Value, pbGSL)
					// 获取源容器的 containerName
					groundStation := c.GroundStations[pbGSL.SourceNodeId-1]
					satellite := c.NormalSatellites[pbGSL.TargetNodeId-1]
					abstractSatellite := c.SatelliteAbstractNodes[pbGSL.TargetNodeId-1]
					// 拿到相应的 abstractLink
					abstractGSL := c.AllGroundSatelliteLinksMap[groundStation.ContainerName]
					// 根据 pbGSL 进行更新
					abstractGSL.TargetNodeId = satellite.Id
					abstractGSL.TargetContainerName = satellite.ContainerName
					abstractGSL.TargetNode = abstractSatellite
				}()
			}
		}
	}()

	// 开启一个线程，准备不断的进行更新事件的获取 (卫星更新事件)
	go func() {
		// 创建一个监听键值对更新事件的 channel
		watchChan := c.etcdClient.Watch(
			c.serviceContext,
			configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.SatellitesPrefix,
			clientv3.WithPrefix(),
		)
		for response := range watchChan {
			for _, event := range response.Events {
				go func() {
					// 创建 protobuf Node
					sat := &posPbNode.Node{}
					// 将 etcd 的值进行反序列化
					protobuf.MustUnmarshal(event.Kv.Value, sat)
					// 获取卫星容器的 pid
					satPid := sat.Pid
					// 获取卫星容器名
					containerName := sat.ContainerName
					// 进行位置的设置
					c.ContainerNameToPosition[containerName] = &position_info.Position{
						NodeType:  types.NetworkNodeType_NormalSatellite.String(), // 节点类型
						Latitude:  float64(sat.Latitude),                          // 纬度
						Longitude: float64(sat.Longitude),                         // 经度
						Altitude:  float64(sat.Altitude),                          // 高度
					}
					// 创建接口数组
					interfaces := make([]string, len(sat.InterfaceDelay))
					// 创建延迟数组
					interfaceDelays := make([]float64, len(sat.InterfaceDelay))
					for index, interfaceAndDelayStr := range sat.InterfaceDelay {
						// InterfaceDelay 存储的是 Interface (str) -> Delay (float) 的映射
						interfaceAndDelay := strings.Split(interfaceAndDelayStr, ":")
						// 获取接口名称
						interfaceName := interfaceAndDelay[0]
						// 获取延迟
						interfaceDelay, _ := strconv.ParseFloat(interfaceAndDelay[1], 64)
						// 存放到切片之中
						interfaces[index] = interfaceName
						interfaceDelays[index] = interfaceDelay
					}
					// 忽略错误 -> 这里会进行标红的原因就是 linux 下才有这个 api
					_ = linux_tc_api.SetInterfacesDelay(int(satPid), interfaces, interfaceDelays)
				}()
			}
		}
		constellationLogger.Infof("local update delay service exit")
	}()

	c.systemStartSteps[StartUpdateDelayService] = struct{}{}
	constellationLogger.Infof("execute update delay service")
	return nil
}
