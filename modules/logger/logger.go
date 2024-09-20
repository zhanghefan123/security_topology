package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

const (
	// 1. main 下的模块

	ModuleMainCmdConstellation = "MAIN_CMD_CONSTELLATION"
	ModuleMainCmdImages        = "MAIN_CMD_IMAGES"
	ModuleMainCmdStatus        = "MAIN_CMD_STATUS"

	// 2. containerconfig 下的模块
	ModuleFrrConfig = "FRR_CONFIG"

	ModuleConfig = "CONFIG"
	ModuleUtils  = "UTILS"

	ModuleNormalNode    = "NORMAL_NODE"
	ModuleConstellation = "CONSTELLATION"

	ModuleAbstractEntities = "ABSTRACT_ENTITIES"
	ModuleProgressBar      = "PROGRESS_BAR"
	ModuleContainerManager = "CONTAINER_MANAGER"
)

func init() {
	InitLogger() // 初始化日志记录器
}

func InitLogger() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{"component"},
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// GetLogger 通过输入的模块名称来获取相应的 logger
func GetLogger(ModuleName string) *logrus.Entry {
	logger := logrus.WithFields(logrus.Fields{
		"component": ModuleName,
	})
	return logger
}
