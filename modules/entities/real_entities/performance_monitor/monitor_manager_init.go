package performance_monitor

import (
	"context"
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
)

const (
	CreateServiceContext = "CreateServiceContext" // 生成上下文->方便取消
	InitializeRecorder   = "InitializeRecorder"
)

type InitFunction func() error

type InitModule struct {
	init         bool
	initFunction InitFunction
}

// Init 进行初始化
func (mm *MonitorManager) Init() error {
	initSteps := []map[string]InitModule{
		{CreateServiceContext: InitModule{true, mm.CreateServiceContext}},
		{InitializeRecorder: InitModule{true, mm.InitializeRecorder}},
	}
	err := mm.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		return fmt.Errorf("monitor manager init failed: %w", err)
	}
	return nil
}

func (mm *MonitorManager) initStepsNum(initSteps []map[string]InitModule) int {
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
func (mm *MonitorManager) initializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := mm.initStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.init {
				if err = initModule.initFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				monitorManagerLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

// CreateServiceContext 创建程序上下文
func (mm *MonitorManager) CreateServiceContext() (err error) {
	if _, ok := mm.MonitorManagerInitSteps[CreateServiceContext]; ok {
		monitorManagerLogger.Infof("CreateServiceContext is already running")
		return nil
	}
	mm.ServiceContext, mm.serviceContextCancelFunc = context.WithCancel(context.Background())

	mm.MonitorManagerInitSteps[CreateServiceContext] = struct{}{}
	monitorManagerLogger.Infof("execute create context")
	return nil
}

func (mm *MonitorManager) InitializeRecorder() error {
	if _, ok := mm.MonitorManagerInitSteps[InitializeRecorder]; ok {
		monitorManagerLogger.Infof("already initialize recorder")
		return nil
	}

	// 进行所有的节点的遍历
	for _, abstractChainNode := range topology.Instance.AllChainAbstractNodes {
		normalNode, err := abstractChainNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("cannot get normal node from abstract node: %v", err)
		}
		InitializeRecorderForContainer(normalNode.ContainerName)
	}

	mm.MonitorManagerInitSteps[InitializeRecorder] = struct{}{}
	monitorManagerLogger.Infof("initialize recorder finished")
	return nil
}
