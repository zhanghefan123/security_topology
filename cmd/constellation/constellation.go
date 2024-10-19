package constellation

import (
	"flag"
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"zhanghefan123/security_topology/api/etcd_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/constellation"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	constellationInstance  *constellation.Constellation
	cmdConstellationLogger = logger.GetLogger(logger.ModuleMainCmdConstellation)
)

// CreateConstellationCmd 创建管理星座的命令
func CreateConstellationCmd() *cobra.Command {
	var constellationCmd = &cobra.Command{
		Use:   "constellation",
		Short: "manage constellation",
		Long:  "manage constellation",
		Run: func(cmd *cobra.Command, args []string) {
			cmdConstellationLogger.Infof("start manage the constellation")
			core()
		},
	}
	return constellationCmd
}

// 管理星座命令的核心
func core() {
	ParseFlag()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	// 启动流程
	// =======================================================
	Initialize()
	PrintExitLogo()
	// =======================================================

	<-signalChan

	// 删除流程
	// =======================================================
	Delete()
	PrintRemovedLogo()
	// =======================================================

}

// Initialize 初始化函数
func Initialize() {
	var err error // 创建错误
	var dockerClient *docker.Client
	// 初始化本地配置
	err = configs.InitLocalConfig()
	if err != nil {
		cmdConstellationLogger.Errorf("init local configuration failed: %v", err)
	}
	// 初始化 dockerClient
	dockerClient, err = client.NewDockerClient() // 创建新的 docker client
	if err != nil {
		cmdConstellationLogger.Errorf("error initilize constellation: %v", err) // 打印错误
	}
	// 初始化 etcdClient
	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
	startTime := configs.TopConfiguration.ConstellationConfig.GoStartTime
	constellationInstance = constellation.NewConstellation(dockerClient, etcdClient, startTime) // 创建一个星座, 使用的参数是 dockerClient
	constellationInstance.Init()                                                                // 进行星座的初始化
	constellationInstance.Start()                                                               // 进行星座的启动
}

// Delete 进行星座的删除
func Delete() {
	constellationInstance.Remove()
}

// ParseFlag 解析选项
func ParseFlag() {
	flag.StringVar(&configs.ConfigurationFilePath, "config", configs.ConfigurationFilePath, "config file path")
}

// PrintExitLogo 打印退出的 Logo
func PrintExitLogo() {
	cmdConstellationLogger.Infof("<------------------------------------->")
	cmdConstellationLogger.Infof("        enter ctl+c exit        ")
	cmdConstellationLogger.Infof("<------------------------------------->")
	fmt.Println()
}

// PrintRemovedLogo 打印删除的 Logo
func PrintRemovedLogo() {
	cmdConstellationLogger.Infof("<------------------------------------->")
	cmdConstellationLogger.Infof("        constellation killed        ")
	cmdConstellationLogger.Infof("<------------------------------------->")
}
