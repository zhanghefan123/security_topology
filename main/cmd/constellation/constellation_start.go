package constellation

import (
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/real_entities/constellation"
	"zhanghefan123/security_topology/modules/sysconfig"
)

// Initialize 初始化函数
func Initialize() {
	sysconfig.InitLocalConfig()                                            // 初始化本地配置
	client.InitDockerClient()                                              // 初始化 docker 客户端
	constellation.ConstellationInstance = constellation.NewConstellation() // 创建一个星座
	constellation.ConstellationInstance.Init()                             // 进行星座的初始化
	constellation.ConstellationInstance.Start()                            // 进行星座的启动
	//constellation.ConstellationInstance.Print(constellation.PrintType_Satellites) // 打印星座卫星
	//constellation.ConstellationInstance.Print(constellation.PrintType_Links)      // 打印链路
}
