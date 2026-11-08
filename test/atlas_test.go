package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/atlas/generate"
)

func TestAtlas(t *testing.T) {
	err := configs.InitLocalConfig()
	if err != nil {
		fmt.Printf("init local config error: %v", err)
	}
	resourcesFilePath := configs.TopConfiguration.PathConfig.ResourcesPath
	multipathFileName := "./multipath_10.txt"
	multipathFilePath := filepath.Join(resourcesFilePath, fmt.Sprintf("multipath/complex/%s", multipathFileName))
	_, segments, _, _ := generate.GenerateAtlasPathsAndSegmentsViaPathsFile(multipathFilePath)

	fmt.Println("--------------------------- segments ----------------------------")
	for _, segment := range segments {
		fmt.Println(segment)
	}
	fmt.Println("--------------------------- segments ----------------------------")
}
