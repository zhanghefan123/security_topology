package constellation

// 这个 constellation 命令已经被淘汰了
// --------------------------------------------------------------------------------
//var (
//	cmdConstellationLogger = logger.GetLogger(logger.ModuleMainCmdConstellation)
//)
//
//// CreateConstellationCmd 创建管理星座的命令
//func CreateConstellationCmd() *cobra.Command {
//	var constellationCmd = &cobra.Command{
//		Use:   "constellation",
//		Short: "manage constellation",
//		Long:  "manage constellation",
//		Run: func(cmd *cobra.Command, args []string) {
//			cmdConstellationLogger.Infof("start manage the constellation")
//			core()
//		},
//	}
//	return constellationCmd
//}
//
//// 管理星座命令的核心
//func core() {
//	ParseFlag()
//
//	signalChan := make(chan os.Signal, 1)
//	signal.Notify(signalChan, os.Interrupt)
//	defer signal.Stop(signalChan)
//
//	// 启动流程
//	// =======================================================
//	err := Initialize()
//	if err != nil {
//		cmdConstellationLogger.Errorf("constellation initialization error: %v", err)
//		return
//	}
//	PrintExitLogo()
//	// =======================================================
//
//	<-signalChan
//
//	// 删除流程
//	// =======================================================
//	err = Delete()
//	if err != nil {
//		cmdConstellationLogger.Errorf("delete constellation error: %v", err)
//		return
//	}
//	PrintRemovedLogo()
//	// =======================================================
//
//}
//
//// Initialize 初始化函数
//func Initialize() error {
//	var err error // 创建错误
//	var dockerClient *docker.Client
//	// 初始化本地配置
//	err = configs.InitLocalConfig()
//	if err != nil {
//		return fmt.Errorf("init local config failed: %w", err)
//	}
//	// 初始化 dockerClient
//	dockerClient, err = client.NewDockerClient() // 创建新的 docker client
//	if err != nil {
//		return fmt.Errorf("create docker client failed: %w", err)
//	}
//	// 初始化 etcdClient
//	listenAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
//	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
//	etcdClient, err := etcd_api.NewEtcdClient(listenAddr, listenPort)
//	// 获取星座创建时间
//	startTime := configs.TopConfiguration.ConstellationConfig.GoStartTime
//	// 创建一个星座, 使用的参数是 dockerClient, etcdClient, startTime
//	constellation.ConstellationInstance = constellation.NewConstellation(dockerClient, etcdClient, startTime)
//	// 进行星座的初始化
//	err = constellation.ConstellationInstance.Init()
//	// 如果初始化失败了
//	if err != nil {
//		return fmt.Errorf("init constellation failed: %w", err)
//	}
//	// 进行星座的启动
//	err = constellation.ConstellationInstance.Start()
//	// 如果星座启动失败了
//	if err != nil {
//		return fmt.Errorf("start constellation failed: %w", err)
//	}
//	return nil
//}
//
//// Delete 进行星座的删除
//func Delete() error {
//	err := constellation.ConstellationInstance.Remove()
//	if err != nil {
//		return fmt.Errorf("remove constellation failed: %w", err)
//	}
//	return nil
//}
//
//// ParseFlag 解析选项
//func ParseFlag() {
//	flag.StringVar(&configs.ConfigurationFilePath, "config", configs.ConfigurationFilePath, "config file path")
//}
//
//// PrintExitLogo 打印退出的 Logo
//func PrintExitLogo() {
//	cmdConstellationLogger.Infof("<------------------------------------->")
//	cmdConstellationLogger.Infof("        enter ctl+c exit        ")
//	cmdConstellationLogger.Infof("<------------------------------------->")
//	fmt.Println()
//}
//
//// PrintRemovedLogo 打印删除的 Logo
//func PrintRemovedLogo() {
//	cmdConstellationLogger.Infof("<------------------------------------->")
//	cmdConstellationLogger.Infof("        constellation killed        ")
//	cmdConstellationLogger.Infof("<------------------------------------->")
//}
// --------------------------------------------------------------------------------
