package normal_node

import (
	"fmt"
	"os"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/modules/sysconfig"
	"zhanghefan123/security_topology/modules/utils/dir"
)

var frrConfigLogger = logger.GetLogger("frr")

const (
	FrrStartInfo = `frr version 7.2.1
frr defaults traditional
hostname %s
log syslog informational
no ipv6 forwarding
service integrated-vtysh-config
!
router ospf
   redistribute connected
`

	InterfaceCommand = `interface %s
   ip ospf network point-to-point
   ip ospf hello-interval 5
   ip ospf dead-interval 20
   ip ospf retransmit-interval 5`

	FrrEndInfo = `!
line vty
!
`
)

func (normalNode *NormalNode) GenerateFrrConfig() {
	finalConfigStr := ""

	frrStartInfo := fmt.Sprintf(FrrStartInfo, normalNode.ContainerName)

	finalConfigStr += frrStartInfo

	// 遍历所有连接的子网
	area := "0.0.0.0"
	for _, subNet := range normalNode.ConnectedSubnetList {
		finalConfigStr += fmt.Sprintf("\t network %s area %s\n", subNet, area)
	}

	// 遍历所有的接口
	for _, intf := range normalNode.IfNameToInterfaceMap {
		interfaceCommand := fmt.Sprintf(InterfaceCommand, intf.IfName)
		finalConfigStr += interfaceCommand + "\n"
	}

	// 添加尾部
	finalConfigStr += FrrEndInfo

	// 获取路径
	outputDir := sysconfig.TopConfiguration.PathConfig.FrrPath.FrrHostPath

	// 进行路径的创建
	dir.GenerateDir(outputDir)

	// 创建一个文件
	f, err := os.Create(fmt.Sprintf("%s/%s.conf", outputDir, normalNode.ContainerName))
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			frrConfigLogger.Errorf("fail to close file %v", err)
		}
	}(f)
	if err != nil {
		frrConfigLogger.Errorf("error create file %v", err)
	}
	_, err = f.WriteString(finalConfigStr)
	if err != nil {
		frrConfigLogger.Errorf("fail to write file %v", err)
	}
}
