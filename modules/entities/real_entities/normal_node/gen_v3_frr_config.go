package normal_node

import (
	"fmt"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
)

const (
	FrrV6StartInfo = `frr version 9.1.0
frr defaults traditional
hostname %s
log syslog informational
service integrated-vtysh-config
!
router ospf6
	ospf6 router-id %d.%d.%d.%d
 	redistribute connected
	area 0.0.0.0 range ::/0
exit
!
`
	InterfaceV6Command = `interface %s
	ipv6 ospf6 area 0.0.0.0
!
`

	FrrEndV6Info = `
line vty
!
`
)

// GenerateOspfV3FrrConfig 进行 frr 配置文件的生成
func (normalNode *NormalNode) GenerateOspfV3FrrConfig() error {
	finalConfigStr := ""

	frrStartInfo := fmt.Sprintf(FrrV6StartInfo, normalNode.ContainerName, normalNode.Id, normalNode.Id,
		normalNode.Id, normalNode.Id)

	finalConfigStr += frrStartInfo

	// 遍历所有的接口
	for _, intf := range normalNode.IfNameToInterfaceMap {
		interfaceCommand := fmt.Sprintf(InterfaceV6Command, intf.IfName)
		finalConfigStr += interfaceCommand
	}

	// 添加尾部
	finalConfigStr += FrrEndV6Info

	// 获取路径
	// /simulation/containerName/route
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "route")

	// 进行路径的创建
	err := dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("GenerateOspfV3FrrConfig err: %s", err)
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
