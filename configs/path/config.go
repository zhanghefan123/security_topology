package path

import (
	"fmt"
	"path/filepath"
)

type PathConfig struct {
	ResourcesPath                  string `mapstructure:"resources_path"`
	ConfigGeneratePath             string `mapstructure:"config_generate_path"`
	GottyPath                      string `mapstructure:"gotty_path"`
	TemplatesFilePath              string `mapstructure:"templates_file_path"`
	ChainmakerGoProjectPath        string `mapstructure:"chainmaker_go_project_path"`
	ChainmakerCryptoGenProjectPath string `mapstructure:"chainmaker_crypto_gen_project_path"`
	FiscoBcosProjectPath           string `mapstructure:"fisco_bcos_project_path"`
	FiscoBcosExamplePath           string `mapstructure:"fisco_bcos_example_path"`
	FiscoBcosConsolePath           string `mapstructure:"fisco_bcos_console_path"`
	FabricProjectPath              string `mapstructure:"fabric_project_path"`
	FabricSamplesPath              string `mapstructure:"fabric_samples_path"`
	FabricNetworkPath              string `mapstructure:"fabric_network_path"`
}

// ConvertToAbsPath 将可能的相对路径全部转换为绝对路径
func (pathConfig *PathConfig) ConvertToAbsPath() error {
	var err error
	if !(filepath.IsAbs(pathConfig.GottyPath)) {
		pathConfig.GottyPath, err = filepath.Abs(pathConfig.GottyPath)
		if err != nil {
			return fmt.Errorf("get absolute path of gotty error: %w", err)
		}
	}
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
	return nil
}
