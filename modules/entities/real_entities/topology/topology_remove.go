package topology

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
)

const (
	StopNodeContainers       = "StopNodeContainers"
	RemoveNodeContainers     = "RemoveNodeContainers"
	RemoveLinks              = "RemoveLinks"
	RemoveConfigurationFiles = "RemoveConfigurationFiles"
)

// RemoveFunction 删除函数
type RemoveFunction func() error

// RemoveModule 删除模块
type RemoveModule struct {
	remove         bool           // 是否进行删除 -> 只有相应的模块启动了才需要进行删除
	removeFunction RemoveFunction // 相应的删除函数
}

// Remove 删除整个星座
func (t *Topology) Remove() error {
	removeSteps := []map[string]RemoveModule{
		{StopNodeContainers: RemoveModule{true, t.StopNodeContainers}},
		{RemoveNodeContainers: RemoveModule{true, t.RemoveNodeContainers}},
		{RemoveLinks: RemoveModule{true, t.RemoveLinks}},
		{RemoveConfigurationFiles: RemoveModule{true, t.RemoveConfigurationFiles}},
	}
	err := t.removeSteps(removeSteps)
	if err != nil {
		return fmt.Errorf("stop topology error: %w", err)
	}
	return nil
}

// removeStepsNum 获取删除模块的数量
func (t *Topology) removeStepsNum(removeSteps []map[string]RemoveModule) int {
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
func (t *Topology) removeSteps(removeSteps []map[string]RemoveModule) (err error) {
	moduleNum := t.removeStepsNum(removeSteps)
	count := 0
	for _, removeStep := range removeSteps {
		for name, removeModule := range removeStep {
			if removeModule.remove {
				if err = removeModule.removeFunction(); err != nil {
					return fmt.Errorf("remove step [%s] failed, %s", name, err)
				}
				topologyLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", count+1, moduleNum, name)
				count += 1
			}
		}
	}
	return
}

// StopNodeContainers 进行容器的停止
func (t *Topology) StopNodeContainers() error {
	if _, ok := t.topologyStopSteps[StopNodeContainers]; ok {
		topologyLogger.Infof("already execute stop node containers")
		return nil
	}

	description := fmt.Sprintf("%20s", "stop nodes")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.StopContainer(t.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	t.topologyStopSteps[StopNodeContainers] = struct{}{}
	topologyLogger.Infof("execute stop node containers")

	return multithread.RunInMultiThread(description, taskFunc, t.AllAbstractNodes)
}

// RemoveNodeContainers 进行容器的删除
func (t *Topology) RemoveNodeContainers() error {
	if _, ok := t.topologyStopSteps[RemoveNodeContainers]; ok {
		topologyLogger.Infof("already execute remove node containers")
		return nil
	}

	description := fmt.Sprintf("%20s", "remove nodes")
	var taskFunc multithread.TaskFunc[*node.AbstractNode] = func(node *node.AbstractNode) error {
		err := container_api.RemoveContainer(t.client, node)
		if err != nil {
			return err
		}
		return nil
	}

	t.topologyStopSteps[RemoveNodeContainers] = struct{}{}
	topologyLogger.Infof("execute remove node containers")

	return multithread.RunInMultiThread(description, taskFunc, t.AllAbstractNodes)
}

// RemoveLinks 进行链路的删除
func (t *Topology) RemoveLinks() error {
	if _, ok := t.topologyStopSteps[RemoveLinks]; ok {
		topologyLogger.Infof("already execute remove links")
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

	t.topologyStopSteps[RemoveLinks] = struct{}{}
	topologyLogger.Infof("execute remove links %s", description)

	return multithread.RunInMultiThread(description, taskFunc, t.Links)
}

// RemoveConfigurationFiles 进行配置文件的删除
func (t *Topology) RemoveConfigurationFiles() error {

	if _, ok := t.topologyStopSteps[RemoveConfigurationFiles]; ok {
		topologyLogger.Infof("already execute remove configuration files")
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

	t.topologyStopSteps[RemoveConfigurationFiles] = struct{}{}
	topologyLogger.Infof("execute remove configuration files")

	return nil
}
