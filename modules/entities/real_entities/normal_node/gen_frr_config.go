package normal_node

import (
	"fmt"
	"os"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
)

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

// GenerateFrrConfig 进行 frr 配置文件的生成
func (normalNode *NormalNode) GenerateFrrConfig() error {
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
	outputDir := configs.TopConfiguration.PathConfig.FrrPath.FrrHostPath

	// 进行路径的创建
	err := dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("GenerateFrrConfig err: %s", err)
	}

	var f *os.File
	// 创建一个文件
	f, err = os.Create(fmt.Sprintf("%s/%s.conf", outputDir, normalNode.ContainerName))
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			err = fmt.Errorf("fail to close file %w", err)
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
