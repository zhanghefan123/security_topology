package topology

import (
	"fmt"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/utils/dir"
)

// RemoveAndRegenerateMultipath 进行删除和重新生成
func (t *Topology) RemoveAndRegenerateMultipath(selectedMultipathType int, multipathFileName string) error {
	err := t.RemoveMultipathRelatedFiles()
	if err != nil {
		return fmt.Errorf("remove %w", err)
	} else {
		fmt.Printf("successfully remove multipath related files\n")
	}
	err = t.RegenerateMultipathFiles(selectedMultipathType, multipathFileName)
	if err != nil {
		return fmt.Errorf("regenerate %w", err)
	} else {
		fmt.Printf("regenerate multipath files \n")
	}
	return nil
}

// RemoveMultipathRelatedFiles 进行相关的文件的生成
func (t *Topology) RemoveMultipathRelatedFiles() error {
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath

	fileNameList := []string{
		"multipath_output_link_identifiers.txt",
		"relationship.txt",
		"num_of_paths.txt",
		"intermediate_multipath_segments.txt",
		"multipath_segments.txt",
		"multipath.txt",
	}

	dir.DeleteFileListInDir(simulationDir, fileNameList)
	return nil
}

// RegenerateMultipathFiles 进行 multipath files 的重新生成
func (t *Topology) RegenerateMultipathFiles(selectedMultipathType int, multipathFileName string) error {
	if selectedMultipathType == int(types.MultipathRoutingType_MultipathSelir) {
		configs.TopConfiguration.StartTopologyParams.PathValidationParams.MultipathConfig.MultipathRoutingType = selectedMultipathType
		configs.TopConfiguration.StartTopologyParams.PathValidationParams.MultipathConfig.MultipathFileName = multipathFileName
		delete(t.topologyInitSteps, GenerateMultipathSelirPathsAndOutputLinkIdentifiers)
		err := t.GenerateMultipathSelirPathsAndOutputLinkIdentifiers()
		if err != nil {
			return fmt.Errorf("generate atlas error %w", err)
		} else {
			fmt.Printf("regenerate multipath files for multipath selir")
		}
		return nil
	} else if selectedMultipathType == int(types.MultipathRoutingType_Atlas) {
		configs.TopConfiguration.StartTopologyParams.PathValidationParams.MultipathConfig.MultipathRoutingType = selectedMultipathType
		configs.TopConfiguration.StartTopologyParams.PathValidationParams.MultipathConfig.MultipathFileName = multipathFileName
		delete(t.topologyInitSteps, GenerateAtlasSegmentsAndOutputLinkIdentifiers)
		err := t.GenerateAtlasSegmentsAndOutputLinkIdentifiers()
		if err != nil {
			return fmt.Errorf("generate multipath selir error %w", err)
		} else {
			fmt.Printf("regenerate multipath files for atlas")
		}
		return nil
	} else {
		return fmt.Errorf("unsupported multipath routing type")
	}
}
