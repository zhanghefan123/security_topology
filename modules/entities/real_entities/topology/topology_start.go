package topology

import (
	"fmt"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
)

const (
	GenerateNodesVethPairs = "GenerateNodesVethPairs"
	StartNodeContainers    = "StartNodeContainers"
	SetVethNameSpaces      = "SetVethNameSpaces"
	SetLinkParameters      = "SetLinkParameters"
)

type StartFunction func() error

type StartModule struct {
	start         bool          // 是否启动
	startFunction StartFunction // 相应的启动函数
}

// Start 启动
func (t *Topology) Start() error {
	startSteps := []map[string]StartModule{
		{GenerateNodesVethPairs: StartModule{true, t.GenerateNodesVethPairs}}, // step1 先创建 veth pair 然后改变链路的命名空间
		{StartNodeContainers: StartModule{true, t.StartNodeContainers}},       // step2 一定要在 step1 之后，因为创建了容器后才有命名空间
		{SetVethNameSpaces: StartModule{true, t.SetVethNamespaces}},           // step3 一定要在 step2 之后，因为创建了容器才能设置 veth 的 namespace
		{SetLinkParameters: StartModule{true, t.SetLinkParameters}},           // step4 进行链路属性的设置
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
