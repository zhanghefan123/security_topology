package normal_node

import (
	"fmt"
	"os"
)

// GenerateIfnameToLidMapping 进行接口名称到Lid映射文件的生成
func (normalNode *NormalNode) GenerateIfnameToLidMapping() (err error) {
	finalString := ""
	err = os.MkdirAll(fmt.Sprintf("/configuration/%s/interface", normalNode.ContainerName), os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdir for interface error: %w", err)
	}
	filePath := fmt.Sprintf("/configuration/%s/interface/interface.txt", normalNode.ContainerName)
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
