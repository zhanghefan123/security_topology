package constellation

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"strconv"
	"strings"
	"sync"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/linux_tc_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/position"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellite"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/entities/utils"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	posPbNode "zhanghefan123/security_topology/services/position/protobuf/node"
)

const (
	StartSatelliteContainers = "StartSatelliteContainers"
	GenerateSatelliteLinks   = "GenerateSatelliteLinks"
	SetVethNameSpaces        = "SetVethNameSpaces"
	StartEtcdService         = "StartEtcdService"
	StartPositionService     = "StartPositionService"
	StoreToEtcd              = "StoreToEtcd"
	StartUpdateDelayService  = "StartUpdateDelayService"
)

type StartFunction func() error

// Start 启动
func (c *Constellation) Start() {
	startSteps := []map[string]StartFunction{
		{GenerateSatelliteLinks: c.GenerateSatelliteVethPairs}, // step1 先创建 veth pair 然后改变链路的命名空间
		{StartSatelliteContainers: c.StartSatelliteContainers}, // step2 一定要在 step1 之后，因为创建了容器后才有命名空间
		{SetVethNameSpaces: c.SetVethNamespaces},               // step3 一定要在 step2 之后，因为创建了容器才能设置 veth 的 namespace
		{StartEtcdService: c.StartEtcdService},                 // step4 进行 etcd 服务的启动
		{StartPositionService: c.StartPositionService},         // step5 启动位置 position 服务
		{StoreToEtcd: c.StoreToEtcd},                           // step5 一定要在 step4 之后，因为创建了 etcd 服务才能进行存储
		{StartUpdateDelayService: c.StartUpdateDelayService},   // step6 开启监听的服务
	}
	err := c.startSteps(startSteps)
	if err != nil {
		constellationLogger.Errorf("constellation start error")
	}
}

// startSteps 调用所有的启动方法
func (c *Constellation) startSteps(startSteps []map[string]StartFunction) (err error) {
	moduleNum := len(startSteps)
	for idx, startStep := range startSteps {
		for name, startFunc := range startStep {
			if err := startFunc(); err != nil {
				constellationLogger.Errorf("start step [%s] failed, %s", name, err)
				return err
			}
			constellationLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

// StartPositionService 启动位置服务
func (c *Constellation) StartPositionService() error {
	// 1. 解析配置
	etcdConfig := configs.TopConfiguration.ServicesConfig.EtcdConfig                            // etcd 配置
	etcdListenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress                // etcd 监听地址
	etcdClientPort := etcdConfig.ClientPort                                                     // etcd 客户端口
	etcdISLsPrefix := etcdConfig.EtcdPrefix.ISLsPrefix                                          // etcd isl 的前缀
	etcdSatellitesPrefix := etcdConfig.EtcdPrefix.SatellitesPrefix                              // etcd satellite 的前缀
	constellationStartTime := configs.TopConfiguration.ConstellationConfig.StartTime            // 星座启动时间
	positionImageName := configs.TopConfiguration.ServicesConfig.PositionUpdateConfig.ImageName // 镜像名称
	updateInterval := configs.TopConfiguration.ServicesConfig.PositionUpdateConfig.Interval     // 更新时间间隔

	// 2. 根据配置创建节点
	positionService := position.NewPositionService(types.NetworkNodeStatus_Logic,
		etcdListenAddr, etcdClientPort,
		etcdISLsPrefix, etcdSatellitesPrefix,
		constellationStartTime, updateInterval, positionImageName)

	// 3. 创建抽象节点
	c.positionService = node.NewAbstractNode(types.NetworkNodeType_PositionService,
		positionService)

	// 4. 进行容器的创建
	err := container_api.CreateContainer(c.client, c.positionService)
	if err != nil {
		return fmt.Errorf("create position service failed, %s", err)
	}

	// 5. 进行容器的启动
	err = container_api.StartContainer(c.client, c.positionService)
	if err != nil {
		return fmt.Errorf("start position service failed, %s", err)
	}
	return nil
}

// StartEtcdService 开启 etcd 服务
func (c *Constellation) StartEtcdService() error {
	// 1. 解析配置
	etcdConfig := configs.TopConfiguration.ServicesConfig.EtcdConfig
	clientPort := etcdConfig.ClientPort
	peerPort := etcdConfig.PeerPort
	dataDir := etcdConfig.DataDir
	etcdName := etcdConfig.EtcdName
	imageName := etcdConfig.ImageName

	// 2. 根据配置创建节点
	actualEtcdNode := etcd.NewEtcdNode(types.NetworkNodeStatus_Logic, clientPort,
		peerPort, dataDir, etcdName, imageName)

	// 3. 创建抽象节点
	c.etcdService = node.NewAbstractNode(types.NetworkNodeType_EtcdNode, actualEtcdNode)

	// 4. 进行容器的创建和启动
	err := container_api.CreateContainer(c.client, c.etcdService)
	if err != nil {
		return fmt.Errorf("create etcd container failed, %s", err.Error())
	}
	err = container_api.StartContainer(c.client, c.etcdService)
	if err != nil {
		return fmt.Errorf("start etcd container failed, %s", err.Error())
	}

	return nil
}

// StartUpdateDelayService 开启更新服务
func (c *Constellation) StartUpdateDelayService() error {
	// 定义一个错误队列
	errChan := make(chan error)

	// 开启一个线程，准备不断的进行更新事件的获取
	go func() {
		watchChan := c.etcdClient.Watch(
			context.Background(),
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
					err := linux_tc_api.SetInterfacesDelay(int(satPid), interfaces, interfaceDelays)
					if err != nil {
						errChan <- fmt.Errorf("set interface delay error")
					}
				}()
			}
		}
	}()

	return nil
}

// StoreToEtcd 将卫星/链路信息放到 Etcd 之中
func (c *Constellation) StoreToEtcd() (err error) {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		for _, satelliteLink := range c.AllSatelliteLinks {
			err = satelliteLink.StoreToEtcd(c.etcdClient)
			if err != nil {
				err = fmt.Errorf("store ISL to etcd error %w", err)
				return
			}
		}
	}()
	go func() {
		defer waitGroup.Done()
		for _, sat := range c.Satellites {
			if sat.Type == types.NetworkNodeType_NormalSatellite {
				normalSat, _ := sat.ActualNode.(*satellite.NormalSatellite)
				err = normalSat.StoreToEtcd(c.etcdClient)

			} else if sat.Type == types.NetworkNodeType_NormalSatellite {
				consensusSat, _ := sat.ActualNode.(*satellite.ConsensusSatellite)
				err = consensusSat.StoreToEtcd(c.etcdClient)
			}
			if err != nil {
				err = fmt.Errorf("store ISL to etcd error %w", err)
				return
			}
		}
	}()
	waitGroup.Wait()
	return nil
}

// GenerateSatelliteVethPairs 进行卫星之间的 veth pairs 的生成
func (c *Constellation) GenerateSatelliteVethPairs() error {
	description := fmt.Sprintf("%20s", "generate veth pairs")
	var taskFunc multithread.TaskFunc[*link.AbstractLink] = func(link *link.AbstractLink) error {
		err := link.GenerateVethPair()
		if err != nil {
			return err
		}
		return nil
	}
	return multithread.RunInMultiThread[*link.AbstractLink](description, taskFunc, c.AllSatelliteLinks)
}

// StartSatelliteContainers 启动卫星容器
func (c *Constellation) StartSatelliteContainers() error {
	description := fmt.Sprintf("%20s", "start satellites")
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
	return multithread.RunInMultiThread(description, taskFunc, c.Satellites)
}

// SetVethNamespaces 设置 veth 命名空间
func (c *Constellation) SetVethNamespaces() error {
	description := fmt.Sprintf("%20s", "set veth namespaces")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		normalNode, err := utils.GetNormalNodeFromAbstractNode(node)
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
	return multithread.RunInMultiThread(description, taskFunc, c.Satellites)
}
