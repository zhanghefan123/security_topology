package constellation

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"os/exec"
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

var (
	ErrRemoveConfigurationFile = errors.New("RemoveConfigurationFile")
)

type RemoveFunction func() error

// Remove 删除整个星座
func (c *Constellation) Remove() {
	removeSteps := []map[string]RemoveFunction{
		{StopSatelliteContainers: c.StopSatelliteContainers},
		{RemoveSatelliteContainers: c.RemoveSatelliteContainers},
		{RemoveLinks: c.RemoveLinks},
		{RemoveConfigurationFiles: c.RemoveConfigurationFiles},
		{RemoveEtcdService: c.RemoveEtcdService},
		{RemovePositionService: c.RemovePositionService},
		{StopLocalServices: c.StopLocalServices},
	}
	err := c.removeSteps(removeSteps)
	if err != nil {
		constellationLogger.Errorf("remove constellation failed %v", err)
	}
}

// startSteps 调用所有的启动方法
func (c *Constellation) removeSteps(removeSteps []map[string]RemoveFunction) (err error) {
	moduleNum := len(removeSteps)
	for idx, removeStep := range removeSteps {
		for name, removeFunc := range removeStep {
			if err := removeFunc(); err != nil {
				return fmt.Errorf("remove step [%s] failed, %s", name, err)
			}
			constellationLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

// StopSatelliteContainers 进行容器的停止
func (c *Constellation) StopSatelliteContainers() error {
	description := fmt.Sprintf("%20s", "stop satellites")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.StopContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}
	return multithread.RunInMultiThread(description, taskFunc, c.Satellites)
}

// RemoveSatelliteContainers 进行容器的删除
func (c *Constellation) RemoveSatelliteContainers() error {
	description := fmt.Sprintf("%20s", "remove satellites")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.RemoveContainer(c.client, node)
		if err != nil {
			return err
		}
		return nil
	}
	return multithread.RunInMultiThread(description, taskFunc, c.Satellites)
}

// RemoveLinks 进行链路的删除
func (c *Constellation) RemoveLinks() error {
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
	return multithread.RunInMultiThread(description, taskFunc, c.AllSatelliteLinks)
}

// RemoveConfigurationFiles 进行配置文件的删除
func (c *Constellation) RemoveConfigurationFiles() error {
	ConfigGeneratePath := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	if !(filepath.IsAbs(ConfigGeneratePath)) {
		configs.TopConfiguration.PathConfig.ConfigGeneratePath, _ = filepath.Abs(ConfigGeneratePath)
	}
	cmd := exec.Command("rm", "-rf", configs.TopConfiguration.PathConfig.ConfigGeneratePath+"/*")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return ErrRemoveConfigurationFile
	}
	return nil
}

// RemoveEtcdService 进行 etcd 服务的关闭
func (c *Constellation) RemoveEtcdService() error {
	err := container_api.StopContainer(c.client, c.etcdService)
	if err != nil {
		return fmt.Errorf("stop etcd service failed, %s", err)
	}
	err = container_api.RemoveContainer(c.client, c.etcdService)
	if err != nil {
		return fmt.Errorf("remove etcd service failed, %s", err)
	}
	return nil
}

// RemovePositionService 进行位置服务的关闭
func (c *Constellation) RemovePositionService() error {
	err := container_api.StopContainer(c.client, c.positionService)
	if err != nil {
		return fmt.Errorf("stop position service failed %v", err)
	}
	err = container_api.RemoveContainer(c.client, c.positionService)
	if err != nil {
		return fmt.Errorf("remove position service failed %v", err)
	}
	return nil
}

// StopLocalServices 进行本地服务的停止
func (c *Constellation) StopLocalServices() error {
	c.serviceContextCancelFunc()
	time.Sleep(1 * time.Second)
	return nil
}
