package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"strconv"
	"strings"
	"time"
	"zhanghefan123/security_topology/configs/apps"
	"zhanghefan123/security_topology/configs/consensus"
	"zhanghefan123/security_topology/configs/constellation"
	"zhanghefan123/security_topology/configs/network"
	"zhanghefan123/security_topology/configs/path"
	"zhanghefan123/security_topology/configs/services"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	ConfigurationFilePath = "../resources/configuration.yml"
	TopConfiguration      = &TopConfig{}
	ConstellationGraph    = simple.NewDirectedGraph()
	moduleConfigLogger    = logger.GetLogger(logger.ModuleConfig)
)

type TopConfig struct {
	NetworkConfig       network.NetworkConfig             `mapstructure:"network_config"`
	ConsensusConfig     consensus.ConsensusConfig         `mapstructure:"consensus_config"`
	ConstellationConfig constellation.ConstellationConfig `mapstructure:"constellation_config"`
	PathConfig          path.PathConfig                   `mapstructure:"path_config"`
	ServicesConfig      services.ServicesConfig           `mapstructure:"services_config"`
	AppsConfig          apps.AppsConfig                   `mapstructure:"apps_config"`
}

func NewDefaultTopConfig() *TopConfig {
	return &TopConfig{}
}

// InitLocalConfig 进行配置的初始化
func InitLocalConfig() error {
	tempViper := viper.New()
	configFilePath := ConfigurationFilePath
	tempViper.SetConfigFile(configFilePath)
	if err := tempViper.ReadInConfig(); err != nil {
		moduleConfigLogger.Errorf("read config error: %v", err)
	}
	TopConfiguration = NewDefaultTopConfig()
	if err := tempViper.Unmarshal(TopConfiguration); err != nil {
		return fmt.Errorf("unmarshal config error: %v", err)
	}
	err := TopConfiguration.PathConfig.ConvertToAbsPath()
	if err != nil {
		return fmt.Errorf("convert to abs path error: %v", err)
	}
	TopConfiguration.ConstellationConfig.GoStartTime = ResolveStartTime()
	PrintLocalConfig()
	return nil
}

// ResolveStartTime 解析初始化的时间
func ResolveStartTime() time.Time {
	startTime := TopConfiguration.ConstellationConfig.StartTime
	result := strings.Split(startTime, "|")
	yearInt, _ := strconv.Atoi(result[0])
	monthInt, _ := strconv.Atoi(result[1])
	dateInt, _ := strconv.Atoi(result[2])
	hourInt, _ := strconv.Atoi(result[3])
	minuteInt, _ := strconv.Atoi(result[4])
	secondInt, _ := strconv.Atoi(result[5])
	return time.Date(yearInt, time.Month(monthInt), dateInt, hourInt, minuteInt, secondInt, 0, time.Local)
}

// PrintLocalConfig 打印日志
func PrintLocalConfig() {
	fmt.Println()
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
	fmt.Println()
}
