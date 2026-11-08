package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"zhanghefan123/security_topology/configs/backend_service"
	"zhanghefan123/security_topology/configs/path"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/services/http_server/params"
	"zhanghefan123/security_topology/utils/network"
)

var (
	ConfigurationFilePath = "../resources/configuration.yml"
	TopConfiguration      = &TopConfig{}
	moduleConfigLogger    = logger.GetLogger(logger.ModuleConfig)
)

type TopConfig struct {
	StartTopologyParams  *params.StartTopologyParams
	PathConfig           path.PathConfig                      `mapstructure:"path_config"`
	BackendServiceConfig backend_service.BackendServiceConfig `mapstructure:"backend_service_config"`
}

func InitLocalConfig() error {
	// 1. 进行配置的设置
	tempViper := viper.New()
	tempViper.SetConfigFile(ConfigurationFilePath)
	if err := tempViper.ReadInConfig(); err != nil {
		moduleConfigLogger.Errorf("read config error: %v", err)
	}

	// 2. 进行配置的加载
	TopConfiguration = &TopConfig{}
	if err := tempViper.Unmarshal(TopConfiguration); err != nil {
		return fmt.Errorf("unmarshal config error: %v", err)
	}

	// 3. 进行处理
	err := ProcessTopConfiguration(TopConfiguration)
	if err != nil {
		return fmt.Errorf("process top configuration error: %w", err)
	}

	return nil
}

// ProcessTopConfiguration prepares the local YAML values and exposes them as default StartTopologyParams. A topology request replaces these defaults.
func ProcessTopConfiguration(topConfiguration *TopConfig) error {
	// 将路径进行转换
	if err := topConfiguration.PathConfig.ConvertToAbsPath(); err != nil {
		return fmt.Errorf("convert to abs path error: %v", err)
	}

	// 进行本地地址的获取
	localNetworkAddr, err := network.GetLocalNetworkAddr()
	if err != nil {
		return fmt.Errorf("get local network addr error: %v", err)
	}
	TopConfiguration.BackendServiceConfig.LocalNetworkAddress = localNetworkAddr

	return nil
}
