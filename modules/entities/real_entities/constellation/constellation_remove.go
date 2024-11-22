package constellation

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"os"
	"path/filepath"
	"time"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
)

const (
	StopSatelliteContainers   = "StopSatelliteContainers"
	RemoveSatelliteContainers = "RemoveSatelliteContainers"
	RemoveLinks               = "RemoveLinks"
	RemoveConfigurationFiles  = "RemoveConfigurationFiles"
	RemoveEtcdService         = "RemoveEtcdService"
	RemovePositionService     = "RemovePositionService"
	StopLocalServices         = "StopLocalServices"
)

type RemoveFunction func() error

type RemoveModule struct {
	remove         bool           // 是否进行删除 -> 只有相应的模块启动了才需要进行删除
	removeFunction RemoveFunction // 相应的删除函数
}

// Remove 删除整个星座
func (c *Constellation) Remove() error {

	removePositionService := configs.TopConfiguration.ServicesConfig.PositionUpdateConfig.Enabled
	removeUpdatedDelayService := configs.TopConfiguration.ServicesConfig.DelayUpdateConfig.Enabled

	removeSteps := []map[string]RemoveModule{
		{StopSatelliteContainers: RemoveModule{true, c.StopSatelliteContainers}},
		{RemoveSatelliteContainers: RemoveModule{true, c.RemoveSatelliteContainers}},
		{RemoveLinks: RemoveModule{true, c.RemoveLinks}},
		{RemoveConfigurationFiles: RemoveModule{true, c.RemoveConfigurationFiles}},
		{RemoveEtcdService: RemoveModule{true, c.RemoveEtcdService}},
		{RemovePositionService: RemoveModule{removePositionService, c.RemovePositionService}},
		{StopLocalServices: RemoveModule{removeUpdatedDelayService, c.StopLocalServices}},
	}
	err := c.removeSteps(removeSteps)
	if err != nil {
		return fmt.Errorf("remove constellation failed %w", err)
	}
	return nil
}

// removeModuleNum 获取删除模块的数量
func (c *Constellation) removeStepsNum(removeSteps []map[string]RemoveModule) int {
	result := 0
	for _, removeStep := range removeSteps {
		for _, removeModule := range removeStep {
			if removeModule.remove {
				result += 1
			}
		}
	}
	return result
}

// startSteps 调用所有的启动方法
func (c *Constellation) removeSteps(removeSteps []map[string]RemoveModule) (err error) {
	moduleNum := c.removeStepsNum(removeSteps)
	count := 0
	for _, removeStep := range removeSteps {
		for name, removeModule := range removeStep {
			if removeModule.remove {
				if err = removeModule.removeFunction(); err != nil {
					return fmt.Errorf("remove step [%s] failed, %s", name, err)
				}
				constellationLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", count+1, moduleNum, name)
				count += 1
			}
		}
	}
	return
}

// StopSatelliteContainers 进行容器的停止
func (c *Constellation) StopSatelliteContainers() error {
	if _, ok := c.systemStopSteps[StopSatelliteContainers]; ok {
		constellationLogger.Infof("already execute stop satellite containers")
		return nil
	}

	description := fmt.Sprintf("%20s", "stop satellites")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.StopContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	c.systemStopSteps[StopSatelliteContainers] = struct{}{}
	constellationLogger.Infof("execute stop satellite containers")

	return multithread.RunInMultiThread(description, taskFunc, c.AllAbstractNodes)
}

// RemoveSatelliteContainers 进行容器的删除
func (c *Constellation) RemoveSatelliteContainers() error {
	if _, ok := c.systemStopSteps[RemoveSatelliteContainers]; ok {
		constellationLogger.Infof("already execute remove satellite containers")
		return nil
	}

	description := fmt.Sprintf("%20s", "remove satellites")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.RemoveContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	c.systemStopSteps[RemoveSatelliteContainers] = struct{}{}
	constellationLogger.Infof("execute remove satellite containers")

	return multithread.RunInMultiThread(description, taskFunc, c.AllAbstractNodes)
}

// RemoveLinks 进行链路的删除
func (c *Constellation) RemoveLinks() error {
	if _, ok := c.systemStopSteps[RemoveLinks]; ok {
		constellationLogger.Infof("already execute remove links")
		return nil
	}

	description := fmt.Sprintf("%20s", "remove links")
	var taskFunc multithread.TaskFunc[*link.AbstractLink] = func(link *link.AbstractLink) error {
		sourceIfName := link.SourceInterface.IfName
		veth, err := netlink.LinkByName(sourceIfName)
		if err != nil {
			return nil // 查找不到可能是已经删掉了
		}
		err = netlink.LinkDel(veth)
		if err != nil {
			return err // 删不了是真正的错误，要返回
		}
		return nil
	}

	c.systemStopSteps[RemoveLinks] = struct{}{}
	constellationLogger.Infof("execute remove links %s", description)

	return multithread.RunInMultiThread(description, taskFunc, c.AllSatelliteLinks)
}

// RemoveConfigurationFiles 进行配置文件的删除
func (c *Constellation) RemoveConfigurationFiles() error {

	if _, ok := c.systemStopSteps[RemoveConfigurationFiles]; ok {
		constellationLogger.Infof("already execute remove configuration files")
		return nil
	}

	ConfigGeneratePath := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	if !(filepath.IsAbs(ConfigGeneratePath)) {
		configs.TopConfiguration.PathConfig.ConfigGeneratePath, _ = filepath.Abs(ConfigGeneratePath)
	}

	// 不要通过命令行 rm -rf 的方法进行删除
	err := os.RemoveAll(ConfigGeneratePath)
	if err != nil {
		return fmt.Errorf("execute remove configuration files failed")
	}

	c.systemStopSteps[RemoveConfigurationFiles] = struct{}{}
	constellationLogger.Infof("execute remove configuration files")

	return nil
}

// RemoveEtcdService 进行 etcd 服务的关闭
func (c *Constellation) RemoveEtcdService() error {
	if _, ok := c.systemStopSteps[RemoveEtcdService]; ok {
		constellationLogger.Infof("already execute remove etcd service")
		return nil
	}

	err := container_api.StopContainer(c.client, c.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("stop etcd service failed, %s", err)
	}
	err = container_api.RemoveContainer(c.client, c.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("remove etcd service failed, %s", err)
	}

	c.systemStopSteps[RemoveEtcdService] = struct{}{}
	constellationLogger.Infof("execute remove etcd service")

	return nil
}

// RemovePositionService 进行位置服务的关闭
func (c *Constellation) RemovePositionService() error {
	if _, ok := c.systemStopSteps[RemovePositionService]; ok {
		constellationLogger.Infof("already execute remove position service")
		return nil
	}

	err := container_api.StopContainer(c.client, c.abstractPositionService)
	if err != nil {
		return fmt.Errorf("stop position service failed %v", err)
	}
	err = container_api.RemoveContainer(c.client, c.abstractPositionService)
	if err != nil {
		return fmt.Errorf("remove position service failed %v", err)
	}

	c.systemStopSteps[RemovePositionService] = struct{}{}
	constellationLogger.Infof("execute remove position service")

	return nil
}

// StopLocalServices 进行本地服务的停止
func (c *Constellation) StopLocalServices() error {
	if _, ok := c.systemStopSteps[StopLocalServices]; ok {
		constellationLogger.Infof("already execute stop local services")
		return nil
	}

	c.serviceContextCancelFunc()
	time.Sleep(1 * time.Second)

	c.systemStopSteps[StopLocalServices] = struct{}{}
	constellationLogger.Infof("execute stop local services")
	return nil
}
