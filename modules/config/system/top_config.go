package system

import (
	"github.com/spf13/viper"
	"reflect"
	"zhanghefan123/security_topology/modules/config/system/consensus"
	"zhanghefan123/security_topology/modules/config/system/constellation"
	"zhanghefan123/security_topology/modules/config/system/network"
	"zhanghefan123/security_topology/modules/config/system/path"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	ConfigurationFilePath string     = "../resources/configuration.yml"
	TopConfiguration      *TopConfig = &TopConfig{}
	moduleConfigLogger               = logger.GetLogger(logger.ModuleConfig)
)

type TopConfig struct {
	NetworkConfig       network.NetworkConfig             `mapstructure:"network_config"`
	ConsensusConfig     consensus.ConsensusConfig         `mapstructure:"consensus_config"`
	ConstellationConfig constellation.ConstellationConfig `mapstructure:"constellation_config"`
	PathConfig          path.PathConfig                   `mapstructure:"path_config"`
}

func NewDefaultTopConfig() *TopConfig {
	return &TopConfig{}
}

// InitLocalConfig 进行配置的初始化
func InitLocalConfig() {
	tempViper := viper.New()
	configFilePath := ConfigurationFilePath
	tempViper.SetConfigFile(configFilePath)
	if err := tempViper.ReadInConfig(); err != nil {
		moduleConfigLogger.Errorf("read config error: %v", err)
	}
	TopConfiguration = NewDefaultTopConfig()
	if err := tempViper.Unmarshal(TopConfiguration); err != nil {
		moduleConfigLogger.Errorf("unmarshal config error: %v", err)
	}
	PrintLocalConfig()
	return
}

// PrintLocalConfig 打印日志
func PrintLocalConfig() {
	// 根据传入的值获取反射对象 -> 注意获取反射对象的一定需要是对象而非反射对象
	reflectObj := reflect.ValueOf(*TopConfiguration)
	// 遍历
	for i := 0; i < reflectObj.NumField(); i++ {
		// 获取字段类型
		field := reflectObj.Type().Field(i)
		// 获取字段值
		value := reflectObj.Field(i)
		// 打印字段名和字段值
		moduleConfigLogger.Infof("%s: %+v", field.Name, value)
	}
}
