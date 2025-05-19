package topology

import (
	"context"
	"fmt"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/types"
)

const (
	StartEtcdService       = "StartEtcdService"
	GenerateNodesVethPairs = "GenerateNodesVethPairs"
	StartNodeContainers    = "StartNodeContainers"
	SetVethNameSpaces      = "SetVethNameSpaces"
	SetLinkParameters      = "SetLinkParameters"
	StoreToEtcd            = "StoreToEtcd"
	UpdateHosts            = "UpdateHosts"
)

type StartFunction func() error

type StartModule struct {
	start         bool          // 是否启动
	startFunction StartFunction // 相应的启动函数
}

// Start 启动
func (t *Topology) Start() error {
	startSteps := []map[string]StartModule{
		{StartEtcdService: StartModule{true, t.StartEtcdService}},
		{StoreToEtcd: StartModule{true, t.StoreToEtcd}},                       // step1 将要存储的东西放到 etcd 之中
		{GenerateNodesVethPairs: StartModule{true, t.GenerateNodesVethPairs}}, // step2 先创建 veth pair 然后改变链路的命名空间
		{StartNodeContainers: StartModule{true, t.StartNodeContainers}},       // step3 一定要在 step1 之后，因为创建了容器后才有命名空间
		{SetVethNameSpaces: StartModule{true, t.SetVethNamespaces}},           // step4 一定要在 step2 之后，因为创建了容器才能设置 veth 的 namespace
		{SetLinkParameters: StartModule{true, t.SetLinkParameters}},           // step5 进行链路属性的设置
		{UpdateHosts: StartModule{true, t.UpdateHosts}},
	}
	err := t.startSteps(startSteps)
	if err != nil {
		return fmt.Errorf("constellation start error: %w", err)
	}
	return nil
}

// startStepsNum 获取启动的模块的数量
func (t *Topology) startStepsNum(startSteps []map[string]StartModule) int {
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
func (t *Topology) startSteps(startSteps []map[string]StartModule) (err error) {
	moduleNum := t.startStepsNum(startSteps)
	for idx, startStep := range startSteps {
		for name, startModule := range startStep {
			// 判断是否需要进行启动, 如果要进行启动，再调用
			if startModule.start {
				if err = startModule.startFunction(); err != nil {
					topologyLogger.Errorf("start step [%s] failed, %s", name, err)
					return err
				}
				topologyLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	return
}

func (t *Topology) UpdateHosts() error {
	finalString := ""
	for _, ordererNode := range t.FabricOrdererNodes {
		orderString := fmt.Sprintf("order%d.example.com", ordererNode.Id)
		firstIpAddressWithPrefix := ordererNode.Interfaces[0].SourceIpv4Addr
		firstIpAddress := firstIpAddressWithPrefix[:len(firstIpAddressWithPrefix)-3]
		finalString = finalString + fmt.Sprintf("%s %s\n", firstIpAddress, orderString)
	}

	for _, peerNode := range t.FabricPeerNodes {
		peerString := fmt.Sprintf("org%d.example.com", peerNode.Id)
		firstIpAddressWithPrefix := peerNode.Interfaces[0].SourceIpv4Addr
		firstIpAddress := firstIpAddressWithPrefix[:len(firstIpAddressWithPrefix)-3]
		finalString = finalString + fmt.Sprintf("%s %s\n", firstIpAddress, peerString)
	}
	fmt.Println(finalString)
	return nil
}

// StartEtcdService 开启 etcd 服务
func (t *Topology) StartEtcdService() error {
	if _, ok := t.topologyStartSteps[StartEtcdService]; ok {
		topologyLogger.Infof("StartEtcdService is already running")
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
	t.etcdService = etcdService

	// 4. 创建抽象节点
	t.abstractEtcdService = node.NewAbstractNode(types.NetworkNodeType_EtcdService, t.etcdService, nil)

	// 5. 进行容器的创建和启动
	err := container_api.CreateContainer(t.client, t.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("create etcd container failed, %s", err.Error())
	}
	err = container_api.StartContainer(t.client, t.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("start etcd container failed, %s", err.Error())
	}

	t.topologyStartSteps[StartEtcdService] = struct{}{}
	topologyLogger.Infof("execute start etcd service")

	return nil
}

// GenerateNodesVethPairs 进行节点之间的 veth pairs 的生成
func (t *Topology) GenerateNodesVethPairs() error {
	if _, ok := t.topologyStartSteps[GenerateNodesVethPairs]; ok {
		topologyLogger.Infof("GenerateNodesVethPairs is already running")
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

	t.topologyStartSteps[GenerateNodesVethPairs] = struct{}{}
	topologyLogger.Infof("generate nodes veth pairs")

	return multithread.RunInMultiThread[*link.AbstractLink](description, taskFunc, t.Links)
}

// StartNodeContainers 启动节点容器
func (t *Topology) StartNodeContainers() error {
	if _, ok := t.topologyStartSteps[StartNodeContainers]; ok {
		topologyLogger.Infof("StartNodeContainers is already running")
		return nil
	}
	description := fmt.Sprintf("%20s", "start nodes")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.CreateContainer(t.client, node)
		if err != nil {
			return err
		}
		err = container_api.StartContainer(t.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	t.topologyStartSteps[StartNodeContainers] = struct{}{}
	topologyLogger.Infof("execute start node containers")

	return multithread.RunInMultiThread(description, taskFunc, t.AllAbstractNodes)
}

// SetVethNamespaces 设置 veth 命名空间
func (t *Topology) SetVethNamespaces() error {
	if _, ok := t.topologyStartSteps[SetVethNameSpaces]; ok {
		topologyLogger.Infof("SetVethNameSpaces is already running")
		return nil
	}
	// 描述
	description := fmt.Sprintf("%20s", "set veth namespaces")
	// 每个节点执行的函数
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

	t.topologyStartSteps[SetVethNameSpaces] = struct{}{}
	topologyLogger.Infof("execute set veth namespaces")

	return multithread.RunInMultiThread(description, taskFunc, t.AllAbstractNodes)
}

// SetLinkParameters 进行链路属性的设置
func (t *Topology) SetLinkParameters() error {
	if _, ok := t.topologyStartSteps[SetLinkParameters]; ok {
		topologyLogger.Infof("SetLinkParameters is already running")
		return nil
	}
	// 描述
	description := fmt.Sprintf("%20s", "set link parameters")
	var taskFunc multithread.TaskFunc[*link.AbstractLink] = func(absLink *link.AbstractLink) error {
		err := absLink.SetLinkParams()
		if err != nil {
			return fmt.Errorf("set link params failed: %w", err)
		}
		return nil
	}

	t.topologyStartSteps[SetLinkParameters] = struct{}{}
	topologyLogger.Infof("execute set link parameters")
	return multithread.RunInMultiThread(description, taskFunc, t.Links)
}

func (t *Topology) StoreToEtcd() error {
	if _, ok := t.topologyStartSteps[StoreToEtcd]; ok {
		topologyLogger.Infof("StoreToEtcd is already running")
		return nil
	}

	startDefenceKey := configs.TopConfiguration.ChainMakerConfig.StartDefenceKey
	if t.TopologyParams.StartDefence {
		_, err := t.EtcdClient.Put(context.Background(), startDefenceKey, "true")
		if err != nil {
			return fmt.Errorf("set start defence failed: %w", err)
		}
	} else {
		_, err := t.EtcdClient.Put(context.Background(), startDefenceKey, "false")
		if err != nil {
			return fmt.Errorf("set start defence failed: %w", err)
		}
	}

	t.topologyStartSteps[StoreToEtcd] = struct{}{}
	topologyLogger.Infof("execute store to etcd")
	return nil
}
