package normal_node

import (
	"fmt"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
)

// 如何打印 frr 的log
// vtysh
// configure terminal
// log file /var/log/frr/ospf.log debugging

const (
	FrrV4StartInfo = `frr version 7.2.1
frr defaults traditional
hostname %s
log syslog debugging
service integrated-vtysh-config
!
router ospf
`

	InterfaceV4Command = `interface %s
   ip ospf area 0
   ip ospf network point-to-point
   ip ospf hello-interval 5
   ip ospf dead-interval 20
   ip ospf retransmit-interval 5`

	FrrEndV4Info = `!
line vty
!
`
)

// GenerateOspfV2FrrConfigForGroundStation 为地面站进行 ospfv2 frr config 的生成
//func (normalNode *NormalNode) GenerateOspfV2FrrConfigForGroundStation(startId int) error {
//	finalConfigStr := ""
//
//	frrStartInfo := fmt.Sprintf(FrrV4StartInfo, normalNode.ContainerName)
//
//	finalConfigStr += frrStartInfo
//
//	// 遍历所有连接的子网
//	//area := "0.0.0.0"
//	//for _, subNet := range normalNode.ConnectedIpv4SubnetList {
//	//	finalConfigStr += fmt.Sprintf("\t network %s area %s\n", subNet, area)
//	//}
//	finalConfigStr += fmt.Sprintf("\t ospf router-id %d.%d.%d.%d\n", startId+normalNode.Id, startId+normalNode.Id, startId+normalNode.Id, startId+normalNode.Id)
//	finalConfigStr += "!\n"
//
//	// 遍历所有的接口
//	for _, intf := range normalNode.IfNameToInterfaceMap {
//		interfaceCommand := fmt.Sprintf(InterfaceV4Command, intf.IfName)
//		finalConfigStr += interfaceCommand + "\n"
//	}
//
//	// 添加尾部
//	finalConfigStr += FrrEndV4Info
//
//	// 获取路径
//	// /simulation/containerName/route
//	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
//	outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "route")
//
//	// 进行路径的创建
//	err := dir.Generate(outputDir)
//	if err != nil {
//		return fmt.Errorf("GenerateOspfV2FrrConfig err: %s", err)
//	}
//
//	var f *os.File
//	// 创建一个文件
//	// /simulation/containerName/route/frr.conf
//	f, err = os.Create(fmt.Sprintf("%s/frr.conf", outputDir))
//	defer func(f *os.File) {
//		fileCloseErr := f.Close()
//		if err == nil {
//			err = fileCloseErr
//		}
//	}(f)
//	if err != nil {
//		return fmt.Errorf("error create file %v", err)
//	}
//	_, err = f.WriteString(finalConfigStr)
//	if err != nil {
//		return fmt.Errorf("fail to write file %v", err)
//	}
//	return nil
//}

// GenerateOspfV2FrrConfig 进行 frr 配置文件的生成
func (normalNode *NormalNode) GenerateOspfV2FrrConfig(routerId int) error {
	finalConfigStr := ""

	frrStartInfo := fmt.Sprintf(FrrV4StartInfo, normalNode.ContainerName)

	finalConfigStr += frrStartInfo

	// 遍历所有连接的子网
	//area := "0.0.0.0"
	//for _, subNet := range normalNode.ConnectedIpv4SubnetList {
	//	finalConfigStr += fmt.Sprintf("\t network %s area %s\n", subNet, area)
	//}
	finalConfigStr += fmt.Sprintf("\t ospf router-id %d.%d.%d.%d\n", routerId, routerId, routerId, routerId)
	finalConfigStr += "!\n"

	// 遍历所有的接口
	for _, intf := range normalNode.IfNameToInterfaceMap {
		interfaceCommand := fmt.Sprintf(InterfaceV4Command, intf.IfName)
		finalConfigStr += interfaceCommand + "\n"
	}

	// 添加尾部
	finalConfigStr += FrrEndV4Info

	// 获取路径
	// /simulation/containerName/route
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "route")

	// 进行路径的创建
	err := dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("GenerateOspfV2FrrConfig err: %s", err)
	}

	var f *os.File
	// 创建一个文件
	// /simulation/containerName/route/frr.conf
	f, err = os.Create(fmt.Sprintf("%s/frr.conf", outputDir))
	defer func(f *os.File) {
		fileCloseErr := f.Close()
		if err == nil {
			err = fileCloseErr
		}
	}(f)
	if err != nil {
		return fmt.Errorf("error create file %v", err)
	}
	_, err = f.WriteString(finalConfigStr)
	if err != nil {
		return fmt.Errorf("fail to write file %v", err)
	}
	return nil
}
