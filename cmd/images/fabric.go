package images

import (
	"fmt"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
)

func buildImageForFabricPeer() error {
	fabricProjectPath := configs.TopConfiguration.FabricConfig.FabricProjectPath
	err := dir.WithContextManager(fabricProjectPath, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("build fabric peer image err: %s", err.Error())
	} else {
		return nil
	}
}

func buildImageForFabricOrder() error {
	fabricProjectPath := configs.TopConfiguration.FabricConfig.FabricProjectPath
	err := dir.WithContextManager(fabricProjectPath, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("build fabric peer image err: %s", err.Error())
	} else {
		return nil
	}
}
