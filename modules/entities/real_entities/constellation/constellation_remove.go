package constellation

import (
	"errors"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"sync"
	"zhanghefan123/security_topology/modules/docker/container_api"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

const (
	StopSatelliteContainers   = "StopSatelliteContainers"
	RemoveSatelliteContainers = "RemoveSatelliteContainers"
	RemoveLinks               = "RemoveLinks"
)

var (
	ErrGetInterfaceFailed = errors.New("GetInterfaceFailed")
)

type RemoveFunction func() error

// Remove 删除整个星座
func (c *Constellation) Remove() {
	removeSteps := []map[string]RemoveFunction{
		{StopSatelliteContainers: c.StopSatelliteContainers},
		{RemoveSatelliteContainers: c.RemoveSatelliteContainers},
		{RemoveLinks: c.RemoveLinks},
	}
	err := c.removeSteps(removeSteps)
	if err != nil {
		moduleConstellationLogger.Errorf("remove constellation failed")
	}
}

// startSteps 调用所有的启动方法
func (c *Constellation) removeSteps(removeSteps []map[string]RemoveFunction) (err error) {
	moduleNum := len(removeSteps)
	for idx, removeStep := range removeSteps {
		for name, removeFunc := range removeStep {
			if err := removeFunc(); err != nil {
				moduleConstellationLogger.Errorf("remove step [%s] failed, %s", name, err)
				return err
			}
			moduleConstellationLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", idx+1, moduleNum, name)
		}
	}
	return
}

func (c *Constellation) StopSatelliteContainers() error {
	satelliteNumber := len(c.Satellites)
	description := "stopSatellites"
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			container_api.StopContainer(satellite)
			waitGroup.Done()
			progress_bar.Add(progressBar, 1)
		}()
	}
	return nil
}

func (c *Constellation) RemoveSatelliteContainers() error {
	satelliteNumber := len(c.Satellites)
	description := "removeSatellites"
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(satelliteNumber)
	for _, satellite := range c.Satellites {
		go func() {
			container_api.RemoveContainer(satellite)
			waitGroup.Done()
			progress_bar.Add(progressBar, 1)
		}()
	}
	return nil
}

func (c *Constellation) RemoveLinks() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		moduleConstellationLogger.Errorf("get interfaces failed, %s", err)
		return ErrGetInterfaceFailed
	}
	prefix := types.GetPrefix(c.SatelliteType)
	for _, iface := range interfaces {
		ifName := iface.Name
		if strings.Contains(ifName, prefix) {
			link, err := netlink.LinkByName(ifName)
			if err != nil {
				moduleConstellationLogger.Errorf("get link failed, %s", err)
				continue
			}
			if err := netlink.LinkDel(link); err != nil {
				moduleConstellationLogger.Errorf("delete link failed, %s", err)
				continue
			}
			moduleConstellationLogger.Infof("delete interface %s success", ifName)
		}
	}
	return nil
}
