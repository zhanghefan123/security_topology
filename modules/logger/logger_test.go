package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLogger(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{"component"},
		TimestampFormat: "2006-01-02 15:04:05",
	})
	starterLogger := logrus.WithFields(logrus.Fields{
		"component": "Starter",
	})
	starterLogger.Info("Hello World") // 输出结果 2024-09-18 15:39:47 [INFO] [Starter] Hello World
}
