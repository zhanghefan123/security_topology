package normal_node

import (
	"fmt"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
)

// GenerateIfnameToLidMapping 进行接口名称到Lid映射文件的生成
func (normalNode *NormalNode) GenerateIfnameToLidMapping() (err error) {
	// 最后的写入的内容
	finalString := ""
	// simulationDir 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	// interface dir 文件架的位置
	outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "interface")
	// 进行文件夹的创建
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdir for interface error: %w", err)
	}
	filePath := filepath.Join(outputDir, "interface.txt")
	for interfaceName, networkIntf := range normalNode.IfNameToInterfaceMap {
		finalString += fmt.Sprintf("%s->%d\n", interfaceName, networkIntf.LinkIdentifier)
	}
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Error opening file %s: %s\n", filePath, err)
	}
	defer func() {
		errClose := file.Close()
		if err == nil {
			err = errClose
		}
	}()
	_, err = file.WriteString(finalString)
	if err != nil {
		return fmt.Errorf("Error writing to file %s: %s\n", filePath, err)
	}
	return nil
}
