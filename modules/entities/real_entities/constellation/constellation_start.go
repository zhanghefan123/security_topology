package constellation

import (
	"sync"
	"zhanghefan123/security_topology/modules/docker/container_api"
	"zhanghefan123/security_topology/modules/entities/utils"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

const (
	StartSatelliteContainers = "StartSatelliteContainers"
	GenerateSatelliteLinks   = "GenerateSatelliteLinks"
	SetVethNameSpaces        = "SetVethNameSpaces"
)

type StartFunction func() error

// Start 启动
func (c *Constellation) Start() {
	startSteps := []map[string]StartFunction{
		{GenerateSatelliteLinks: c.GenerateSatelliteVethPairs}, // step1 先创建 veth pair 然后改变链路的命名空间
		{StartSatelliteContainers: c.StartSatelliteContainers}, // step2 一定要在 step1 之后，因为创建了容器后才有命名空间
		{SetVethNameSpaces: c.SetVethNamespaces},
	}
	err := c.startSteps(startSteps)
	if err != nil {
		moduleConstellationLogger.Errorf("constellation start error")
	}
}

// startSteps 调用所有的启动方法
func (c *Constellation) startSteps(startSteps []map[string]StartFunction) (err error) {
	moduleNum := len(startSteps)
	for idx, startStep := range startSteps {
		for name, startFunc := range startStep {
			if err := startFunc(); err != nil {
				moduleConstellationLogger.Errorf("start step [%s] failed, %s", name, err)
				return err
			}
			moduleConstellationLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

// GenerateSatelliteVethPairs 生成 veth pairs
func (c *Constellation) GenerateSatelliteVethPairs() error {
	linkNums := len(c.AllSatelliteLinks)
	description := "generate veth pairs"
	progressBar := progress_bar.NewProgressBar(linkNums, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(linkNums)
	for _, link := range c.AllSatelliteLinks {
		go func() {
			link.GenerateVethPair()
			waitGroup.Done()
			progress_bar.Add(progressBar, 1)
		}()
	}
	waitGroup.Wait()
	return nil
}

// StartSatelliteContainers 生成卫星容器
func (c *Constellation) StartSatelliteContainers() error {
	satelliteNumber := len(c.Satellites)
	description := "create satellites"
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			container_api.CreateContainer(satellite)
			container_api.StartContainer(satellite)
			waitGroup.Done()
			progress_bar.Add(progressBar, 1)
		}()
	}
	return nil
}

// SetVethNamespaces 进行 veth 的设置
func (c *Constellation) SetVethNamespaces() error {
	satelliteNumber := len(c.Satellites)
	description := "set veth namespaces"
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			normalNode := utils.GetNormalNodeFromAbstractNode(satellite)
			normalNode.SetVethNamespace()
			waitGroup.Done()
			progress_bar.Add(progressBar, 1)
		}()
	}
	return nil
}
