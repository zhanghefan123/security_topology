package steps

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
)

const (
	InitGraphFromConfigurationFile = "InitGraphFromConfigurationFile"
)

type InitFunction func() error

type InitModule struct {
	Init         bool
	InitFunction InitFunction
}

// Init 进行初始化
func (s *Simulator) Init() error {
	initSteps := []map[string]InitModule{
		{InitGraphFromConfigurationFile: InitModule{true, s.InitGraphFromConfigurationFile}},
	}
	err := s.InitializeSteps(initSteps)
	if err != nil {
		return fmt.Errorf("simulator init failed: %w", err)
	}
	return nil
}

// InitStepsNum 计算需要进行初始化的步骤的数量
func (s *Simulator) InitStepsNum(initSteps []map[string]InitModule) int {
	result := 0
	for _, initStep := range initSteps {
		for _, initModule := range initStep {
			if initModule.Init {
				result += 1
			}
		}
	}
	return result
}

// InitializeSteps 按步骤进行初始化
func (s *Simulator) InitializeSteps(initSteps []map[string]InitModule) (err error) {
	fmt.Println()
	moduleNum := s.InitStepsNum(initSteps)
	for idx, initStep := range initSteps {
		for name, initModule := range initStep {
			if initModule.Init {
				if err = initModule.InitFunction(); err != nil {
					return fmt.Errorf("init step [%s] failed, %s", name, err)
				}
				fmt.Printf("SIMULATOR INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	fmt.Println()
	return
}

func (s *Simulator) InitGraphFromConfigurationFile() error {
	s.SimGraph = entities.NewSimGraph()
	// step1: load graph params
	err := s.SimGraph.LoadGraphParamsFromConfigurationFile(s.SimulationGraphPath)
	if err != nil {
		return fmt.Errorf("init graph from configuration file failed, %s", err)
	}
	// step2: load nodes
	err = s.SimGraph.LoadNodesFromNodeParams()
	if err != nil {
		return fmt.Errorf("init graph nodes from configuration file failed, %s", err)
	}
	// step3: load source and destination
	err = s.SimGraph.LoadSourceAndDest()
	if err != nil {
		return fmt.Errorf("load source and destination from configuration file failed, %s", err)
	}
	// step4: load links
	err = s.SimGraph.LoadLinksFromLinkParams()
	if err != nil {
		return fmt.Errorf("init graph links from configuration file failed, %s", err)
	}
	// step5: load coverage paths
	err = s.SimGraph.LoadCoveragePathsFromParams()
	if err != nil {
		return fmt.Errorf("load coverage paths from configuration file failed, %s", err)
	}
	// step6: calculate k shortest paths
	err = s.SimGraph.CalculateKShortestPaths()
	if err != nil {
		return fmt.Errorf("calculate k shortest paths failed, %s", err)
	}
	return nil
}
