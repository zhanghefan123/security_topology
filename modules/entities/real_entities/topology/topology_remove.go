package topology

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/vishvananda/netlink"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/api/multithread"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/webshell"
	"zhanghefan123/security_topology/utils/dir"
	"zhanghefan123/security_topology/utils/execute"
)

const (
	DeleteWebShells           = "DeleteWebShells"
	StopNodeContainers        = "StopNodeContainers"
	RemoveNodeContainers      = "RemoveNodeContainers"
	RemoveLinks               = "RemoveLinks"
	RemoveChainCodeContainers = "RemoveChainCodeContainers"
	RemoveEtcdService         = "RemoveEtcdService"
	RemoveChainMakerFiles     = "RemoveChainMakerFiles"
	RemoveFabricFiles         = "RemoveFabricFiles"
	RemoveVolumes             = "RemoveVolumes"
	RemoveDefaultRoutes       = "RemoveDefaultRoutes"
	RemoveAllChainCodeImages  = "RemoveAllChainCodeImages"
	RemoveAllFiscoBcosFiles   = "RemoveAllFiscoBcosFiles"
	RemoveAllSimulationDirs   = "RemoveAllSimulationDirs"
	StopLocalServices         = "StopLocalServices"
)

// RemoveFunction 删除函数
type RemoveFunction func() error

// RemoveModule 删除模块
type RemoveModule struct {
	remove         bool           // 是否进行删除 -> 只有相应的模块启动了才需要进行删除
	removeFunction RemoveFunction // 相应的删除函数
}

// Remove 删除整个拓扑
func (t *Topology) Remove() error {

	removeSteps := []map[string]RemoveModule{
		{DeleteWebShells: RemoveModule{true, t.DeleteWebShells}},
		{RemoveChainCodeContainers: RemoveModule{true, t.RemoveChaincodeContainers}},
		{StopNodeContainers: RemoveModule{true, t.StopNodeContainers}},
		{RemoveNodeContainers: RemoveModule{true, t.RemoveNodeContainers}},
		{RemoveLinks: RemoveModule{true, t.RemoveLinks}},
		{RemoveEtcdService: RemoveModule{true, t.RemoveEtcdService}},
		{RemoveChainMakerFiles: RemoveModule{t.ChainMakerEnabled, t.RemoveChainMakerFiles}},
		{RemoveFabricFiles: RemoveModule{t.FabricEnabled, t.RemoveFabricFiles}},
		{RemoveVolumes: RemoveModule{true, t.RemoveVolumes}},
		{RemoveDefaultRoutes: RemoveModule{t.FabricEnabled, t.RemoveDefaultRoutes}},
		{RemoveAllChainCodeImages: RemoveModule{t.FabricEnabled, t.RemoveAllChainCodeImages}},
		{RemoveAllFiscoBcosFiles: RemoveModule{t.FiscoBcosEnabled, t.RemoveAllFiscoBcosFiles}},
		{RemoveAllSimulationDirs: RemoveModule{true, t.RemoveAllSimulationDirs}},
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

// DeleteWebShells 进行所有的 web shell 的删除
func (t *Topology) DeleteWebShells() error {
	if _, ok := t.topologyStopSteps[DeleteWebShells]; ok {
		topologyLogger.Infof("already delete web shells")
		return nil
	}

	for pid, _ := range webshell.WebShellPids {
		killCmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
		err := killCmd.Start()
		if err != nil {
			return fmt.Errorf("webshell kill failed: %w", err)
		}
	}

	t.topologyStopSteps[DeleteWebShells] = struct{}{}
	topologyLogger.Infof("delete web shells")
	return nil
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

// RemoveEtcdService 进行 etcd 服务的关闭
func (t *Topology) RemoveEtcdService() error {
	if _, ok := t.topologyStopSteps[RemoveEtcdService]; ok {
		topologyLogger.Infof("already execute remove etcd service")
		return nil
	}

	err := container_api.StopContainer(t.client, t.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("stop etcd service failed, %s", err)
	}
	err = container_api.RemoveContainer(t.client, t.abstractEtcdService)
	if err != nil {
		return fmt.Errorf("remove etcd service failed, %s", err)
	}

	t.topologyStopSteps[RemoveEtcdService] = struct{}{}
	topologyLogger.Infof("execute remove etcd service")

	return nil
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

// RemoveChaincodeContainers 删除 fabric 链码对应的容器
func (t *Topology) RemoveChaincodeContainers() error {
	if _, ok := t.topologyStopSteps[RemoveChainCodeContainers]; ok {
		topologyLogger.Infof("already execute remove chaincode containers")
		return nil
	}

	chainCodeContainerFilter := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: "dev-peer",
	})

	containers, err := t.client.ContainerList(context.Background(), types.ContainerListOptions{
		All:     true,
		Filters: chainCodeContainerFilter,
	})
	if err != nil {
		return fmt.Errorf("get chaincode containers failed: %w", err)
	}

	for _, chainCodeContainer := range containers {
		err = t.client.ContainerStop(
			context.Background(),
			chainCodeContainer.ID,
			container.StopOptions{})
		if err != nil {
			return fmt.Errorf("stop container error: %w", err)
		}
		err = t.client.ContainerRemove(
			context.Background(),
			chainCodeContainer.ID,
			dockerTypes.ContainerRemoveOptions{})
		if err != nil {
			return fmt.Errorf("container remove failed: %v", err)
		}
	}

	t.topologyStopSteps[RemoveChainCodeContainers] = struct{}{}
	topologyLogger.Infof("execute remove chaincode containers")

	return nil
}

// RemoveChainMakerFiles 进行长安链相关文件的删除
func (t *Topology) RemoveChainMakerFiles() error {
	if _, ok := t.topologyStopSteps[RemoveChainMakerFiles]; ok {
		topologyLogger.Infof("already execute remove chainmaker files")
		return nil
	}

	chainMakerGoProjectPath := configs.TopConfiguration.ChainMakerConfig.ChainMakerGoProjectPath
	multiNodePath := path.Join(chainMakerGoProjectPath, "scripts/docker/multi_node")
	testDataPath := path.Join(chainMakerGoProjectPath, "tools/cmc/testdata")
	configPath := path.Join(multiNodePath, "config")
	dataPath := path.Join(multiNodePath, "data")
	logPath := path.Join(multiNodePath, "log")
	cmdTestCryptoConfigData := path.Join("../cmd", "testdata/crypto-config")

	deleteDirs := []string{"./build/", "./crypto-config/", testDataPath, configPath, dataPath, logPath, cmdTestCryptoConfigData}

	for _, deleteDir := range deleteDirs {
		err := os.RemoveAll(deleteDir)
		if err != nil {
			return fmt.Errorf("execute remove chainmaker files failed")
		}
	}

	t.topologyStopSteps[RemoveChainMakerFiles] = struct{}{}
	topologyLogger.Infof("execute remove chainmaker files")
	return nil
}

// RemoveFabricFiles 进行 Fabric 所生成的配置文件的删除
func (t *Topology) RemoveFabricFiles() error {
	if _, ok := t.topologyStopSteps[RemoveFabricFiles]; ok {
		topologyLogger.Infof("already execute remove fabric files")
		return nil
	}

	testNetworkPath := configs.TopConfiguration.FabricConfig.FabricNetworkPath
	organizationsPath := path.Join(testNetworkPath, "organizations")
	ordererOrganizationsPath := path.Join(organizationsPath, "ordererOrganizations")
	peerOrganizationsPath := path.Join(organizationsPath, "peerOrganizations")
	cryptogenFiles := path.Join(organizationsPath, "cryptogen/")

	commandStr := fmt.Sprintf("-rf %s %s", ordererOrganizationsPath, peerOrganizationsPath)

	err := execute.Command("rm", strings.Split(commandStr, " "))
	if err != nil {
		return fmt.Errorf("remove fabric files failed: %w", err)
	}

	// 不能使用 rm -rf * go 无法识别
	err = dir.ClearDir(cryptogenFiles)
	if err != nil {
		return fmt.Errorf("remove fabric files failed: %w", err)
	}

	t.topologyStopSteps[RemoveFabricFiles] = struct{}{}
	topologyLogger.Infof("execute remove fabric files")
	return nil
}

// RemoveVolumes 删除所有的容器卷映射
func (t *Topology) RemoveVolumes() error {
	if _, ok := t.topologyStopSteps[RemoveVolumes]; ok {
		topologyLogger.Infof("already execute remove volumes")
		return nil
	}

	volumes, err := t.client.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		return fmt.Errorf("get volumes failed: %w", err)
	}
	for _, dockerVolume := range volumes.Volumes {
		err = t.client.VolumeRemove(context.Background(), dockerVolume.Name, true)
		if err != nil {
			return fmt.Errorf("remove volume %s failed: %w", dockerVolume.Name, err)
		}
	}

	t.topologyStopSteps[RemoveVolumes] = struct{}{}
	topologyLogger.Infof("execute remove volumes")
	return nil
}

// RemoveDefaultRoutes 删除 fabric 的默认路由
func (t *Topology) RemoveDefaultRoutes() error {
	if _, ok := t.topologyStopSteps[RemoveDefaultRoutes]; ok {
		topologyLogger.Infof("already execute remove default routes")
		return nil
	}

	// 只需要到 peer 以及到第一个 orderer 的路由即可  (一定要等路由收敛之后再安装链码)
	// -----------------------------------------------------------------
	for _, abstractNode := range t.FabricPeerAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return err
		}
		firstInterface := normalNode.Interfaces[0]
		deleteRouteCommand := fmt.Sprintf("del -host %s gw %s", firstInterface.SourceIpv4Addr[:len(firstInterface.SourceIpv4Addr)-3], normalNode.DockerZeroNetworkAddress)
		fmt.Println(deleteRouteCommand)
		err = execute.Command("route", strings.Split(deleteRouteCommand, " "))
		if err != nil {
			return fmt.Errorf("del default route failed: %w", err)
		}
	}

	t.topologyStopSteps[RemoveDefaultRoutes] = struct{}{}
	topologyLogger.Infof("execute remove default routes")
	return nil
}

// RemoveAllChainCodeImages 删除 chaincode 镜像
func (t *Topology) RemoveAllChainCodeImages() error {
	if _, ok := t.topologyStopSteps[RemoveAllChainCodeImages]; ok {
		topologyLogger.Infof("already execute remove all chaincode images")
		return nil
	}

	images, err := t.client.ImageList(context.Background(), dockerTypes.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("get chaincode images failed: %w", err)
	}

	for _, image := range images {
		if len(image.RepoTags) != 0 {
			if strings.Contains(image.RepoTags[0], "dev-peer") {
				_, err = t.client.ImageRemove(context.Background(), image.ID, dockerTypes.ImageRemoveOptions{
					Force:         true,
					PruneChildren: true,
				})
				if err != nil {
					return fmt.Errorf("remove chaincode image failed: %w", err)
				}
				fmt.Printf("remove chaincode image %s success\n", image.RepoTags[0])
			}
		}
	}

	t.topologyStopSteps[RemoveAllChainCodeImages] = struct{}{}
	topologyLogger.Infof("execute remove all chaincode images")
	return nil
}

// RemoveAllFiscoBcosFiles 删除 fisco bcos example 下的文件
func (t *Topology) RemoveAllFiscoBcosFiles() error {
	if _, ok := t.topologyStopSteps[RemoveAllFiscoBcosFiles]; ok {
		topologyLogger.Infof("already remove all fisco bcos files")
		return nil
	}

	dirToDelete := filepath.Join(configs.TopConfiguration.FiscoBcosConfig.ExamplePath, "nodes/")
	err := os.RemoveAll(dirToDelete)
	if err != nil {
		return fmt.Errorf("failed to delete the fisco bcos generated files")
	}

	t.topologyStopSteps[RemoveAllFiscoBcosFiles] = struct{}{}
	return nil
}

// RemoveAllSimulationDirs 删除所有的 simulations 文件夹下的内容
func (t *Topology) RemoveAllSimulationDirs() error {
	if _, ok := t.topologyStopSteps[RemoveAllSimulationDirs]; ok {
		topologyLogger.Infof("remove all simulation dirs called")
		return nil
	}

	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath

	err := dir.ClearDir(simulationDir)
	if err != nil {
		return fmt.Errorf("clear dir failed due to: %v\n", err)
	}

	t.topologyStopSteps[RemoveAllSimulationDirs] = struct{}{}
	return nil
}
