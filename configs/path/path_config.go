package path

import (
	"fmt"
	"path/filepath"
)

type PathConfig struct {
	ResourcesPath       string  `mapstructure:"resources_path"`
	ConfigGeneratePath  string  `mapstructure:"config_generate_path"`
	FrrPath             FrrPath `mapstructure:"frr_path"`
	AddressMappingPath  string  `mapstructure:"address_mapping_path"`
	RealTimePositionDir string  `mapstructure:"real_time_position_dir"`
	GottyPath           string  `mapstructure:"gotty_path"`
}

// ConvertToAbsPath 将可能的相对路径全部转换为绝对路径
func (pathConfig *PathConfig) ConvertToAbsPath() error {
	var err error
	if !(filepath.IsAbs(pathConfig.ResourcesPath)) {
		pathConfig.ResourcesPath, err = filepath.Abs(pathConfig.ResourcesPath)
		if err != nil {
			return fmt.Errorf("get absolute path of resources path: %w", err)
		}
	}
	if !(filepath.IsAbs(pathConfig.ConfigGeneratePath)) {
		pathConfig.ConfigGeneratePath, err = filepath.Abs(pathConfig.ConfigGeneratePath)
		if err != nil {
			return fmt.Errorf("get absolute path of config generate path: %w", err)
		}
	}
	if !(filepath.IsAbs(pathConfig.FrrPath.FrrHostPath)) {
		pathConfig.FrrPath.FrrHostPath, err = filepath.Abs(pathConfig.FrrPath.FrrHostPath)
		if err != nil {
			return fmt.Errorf("get absolute path of frr host path: %w", err)
		}
	}
	return nil
}
