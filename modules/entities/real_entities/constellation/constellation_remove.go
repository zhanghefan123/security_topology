package constellation

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"zhanghefan123/security_topology/modules/docker/container_api"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/sysconfig"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

const (
	StopSatelliteContainers   = "StopSatelliteContainers"
	RemoveSatelliteContainers = "RemoveSatelliteContainers"
	RemoveLinks               = "RemoveLinks"
	RemoveConfigurationFiles  = "RemoveConfigurationFiles"
)

var (
	ErrGetInterfaceFailed      = errors.New("GetInterfaceFailed")
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
	}
	err := c.removeSteps(removeSteps)
	if err != nil {
		ConstellationLogger.Errorf("remove constellation failed")
	}
}

// startSteps 调用所有的启动方法
func (c *Constellation) removeSteps(removeSteps []map[string]RemoveFunction) (err error) {
	moduleNum := len(removeSteps)
	for idx, removeStep := range removeSteps {
		for name, removeFunc := range removeStep {
			if err := removeFunc(); err != nil {
				ConstellationLogger.Errorf("remove step [%s] failed, %s", name, err)
				return err
			}
			ConstellationLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

func (c *Constellation) StopSatelliteContainers() error {
	satelliteNumber := len(c.Satellites)
	description := fmt.Sprintf("%20s", "stop satellites")
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			container_api.StopContainer(satellite)
			waitGroup.Done()
			progressBar.Add(1)
		}()
	}
	waitGroup.Wait()
	if progressBar.IsFinished() {
		fmt.Println()
	}
	return nil
}

func (c *Constellation) RemoveSatelliteContainers() error {
	satelliteNumber := len(c.Satellites)
	description := fmt.Sprintf("%20s", "remove satellites")
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			container_api.RemoveContainer(satellite)
			waitGroup.Done()
			progressBar.Add(1)
		}()
	}
	waitGroup.Wait()
	if progressBar.IsFinished() {
		fmt.Println()
	}
	return nil
}

func (c *Constellation) RemoveLinks() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		ConstellationLogger.Errorf("get interfaces failed, %s", err)
		return ErrGetInterfaceFailed
	}
	prefix := types.GetPrefix(c.SatelliteType)
	for _, iface := range interfaces {
		ifName := iface.Name
		if strings.Contains(ifName, prefix) {
			link, err := netlink.LinkByName(ifName)
			if err != nil {
				ConstellationLogger.Errorf("get link failed, %s", err)
				continue
			}
			if err := netlink.LinkDel(link); err != nil {
				// ConstellationLogger.Errorf("delete link failed, %s", err)
				continue
			}
			ConstellationLogger.Infof("delete interface %s success", ifName)
		}
	}
	return nil
}

// RemoveConfigurationFiles 进行配置文件的删除
func (c *Constellation) RemoveConfigurationFiles() error {
	ConfigGeneratePath := sysconfig.TopConfiguration.PathConfig.ConfigGeneratePath
	if !(filepath.IsAbs(ConfigGeneratePath)) {
		sysconfig.TopConfiguration.PathConfig.ConfigGeneratePath, _ = filepath.Abs(ConfigGeneratePath)
	}
	cmd := exec.Command("rm", "-rf", sysconfig.TopConfiguration.PathConfig.ConfigGeneratePath+"/*")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return ErrRemoveConfigurationFile
	}
	return nil
}
