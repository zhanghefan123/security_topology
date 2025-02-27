package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"zhanghefan123/security_topology/configs/apps"
	"zhanghefan123/security_topology/configs/chainmaker"
	"zhanghefan123/security_topology/configs/consensus"
	"zhanghefan123/security_topology/configs/constellation"
	"zhanghefan123/security_topology/configs/fabric"
	"zhanghefan123/security_topology/configs/images"
	"zhanghefan123/security_topology/configs/network"
	"zhanghefan123/security_topology/configs/path"
	"zhanghefan123/security_topology/configs/path_validation"
	"zhanghefan123/security_topology/configs/resources"
	"zhanghefan123/security_topology/configs/services"
	"zhanghefan123/security_topology/modules/logger"
	networkUtils "zhanghefan123/security_topology/modules/utils/network"
)

var (
	ConfigurationFilePath = "../resources/configuration.yml"
	TopConfiguration      = &TopConfig{}
	moduleConfigLogger    = logger.GetLogger(logger.ModuleConfig)
)

type TopConfig struct {
	NetworkConfig        network.NetworkConfig                `mapstructure:"network_config"`
	ConsensusConfig      consensus.ConsensusConfig            `mapstructure:"consensus_config"`
	ConstellationConfig  constellation.ConstellationConfig    `mapstructure:"constellation_config"`
	ChainMakerConfig     chainmaker.ChainMakerConfig          `mapstructure:"chain_maker_config"`
	ImagesConfig         images.ImagesConfig                  `mapstructure:"images_config"`
	PathConfig           path.PathConfig                      `mapstructure:"path_config"`
	ServicesConfig       services.ServicesConfig              `mapstructure:"services_config"`
	AppsConfig           apps.AppsConfig                      `mapstructure:"apps_config"`
	ResourcesConfig      resources.ResourcesConfig            `mapstructure:"resources_config"`
	PathValidationConfig path_validation.PathValidationConfig `mapstructure:"path_validation_config"`
	FabricConfig         fabric.FabricConfig                  `mapstructure:"fabric_config"`
}

var availableOspfVersions = map[string]struct{}{
	"ospfv2": {},
	"ospfv3": {},
}

func NewDefaultTopConfig() *TopConfig {
	return &TopConfig{}
}

// InitLocalConfig 进行配置的初始化
func InitLocalConfig() error {
	var err error
	// 初始化解析器
	tempViper := viper.New()
	// 设置yml文件的地址
	configFilePath := ConfigurationFilePath
	tempViper.SetConfigFile(configFilePath)
	// 将yml加载
	if err = tempViper.ReadInConfig(); err != nil {
		moduleConfigLogger.Errorf("read config error: %v", err)
	}
	// 将yml进行反序列化
	TopConfiguration = NewDefaultTopConfig()
	if err = tempViper.Unmarshal(TopConfiguration); err != nil {
		return fmt.Errorf("unmarshal config error: %v", err)
	}
	// 进行顶级配置对象的处理
	if err = ProcessTopConfiguration(TopConfiguration); err != nil {
		return fmt.Errorf("process top configuration error: %v", err)
	}
	// 进行处理后的配置的打印
	PrintLocalConfig()
	return nil
}

// ProcessTopConfiguration 进行顶级配置对象的处理
func ProcessTopConfiguration(topConfiguration *TopConfig) error {
	// 解析 chainmaker-go 的一些内部路径
	chainMakerGoProjectPath := topConfiguration.ChainMakerConfig.ChainMakerGoProjectPath
	topConfiguration.ChainMakerConfig.CryptoGenProjectPath = filepath.Join(chainMakerGoProjectPath, "/tools/chainmaker-cryptogen")
	// 将路径转换为绝对路径
	err := topConfiguration.PathConfig.ConvertToAbsPath()
	if err != nil {
		return fmt.Errorf("convert to abs path error: %v", err)
	}
	// 进行时间的解析
	topConfiguration.ConstellationConfig.GoStartTime = ResolveStartTime()
	// 获取本地的网络地址
	localNetworkAddr, err := networkUtils.GetLocalNetworkAddr()
	if err != nil {
		return fmt.Errorf("get local network addr error: %v", err)
	}
	// 进行本地网路地址的设置
	topConfiguration.NetworkConfig.LocalNetworkAddress = localNetworkAddr
	// 进行 ospf 版本号的判断
	if _, ok := availableOspfVersions[topConfiguration.NetworkConfig.OspfVersion]; !ok {
		return fmt.Errorf("unsupported OSPF version: %s", topConfiguration.NetworkConfig.OspfVersion)
	}
	// 进行 constellation bandwidth 的更新
	topConfiguration.ConstellationConfig.ISLBandwidth = topConfiguration.ConstellationConfig.ISLBandwidth * 1e6
	// 没有错误,返回 nil
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
