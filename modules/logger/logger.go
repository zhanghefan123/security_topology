package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

const (
	ModuleMainCmdHttpService   = "MAIN_CMD_HTTP_SERVICE"
	ModuleMainCmdConstellation = "MAIN_CMD_CONSTELLATION"
	ModuleMainCmdImages        = "MAIN_CMD_IMAGES"
	ModuleMainCmdTest          = "MAIN_CMD_TEST"
	ModuleConfig               = "CONFIG"
	ModuleConstellation        = "CONSTELLATION"
	ModuleTopology             = "TOPOLOGY"
	ModuleHttpService          = "HTTP_SERVICE"
	ModulePrepare              = "PREPARE"
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
