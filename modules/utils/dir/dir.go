package dir

import (
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleUtils = logger.GetLogger(logger.ModuleUtils)

func GenerateDir(generatePath string) {
	if !filepath.IsAbs(generatePath) {
		generatePath, _ = filepath.Abs(generatePath)
	}
	if _, err := os.Stat(generatePath); os.IsNotExist(err) {
		err = os.MkdirAll(generatePath, os.ModePerm)
		if err != nil {
			moduleUtils.Errorf("mkdir failed for reason %v", err)
		}
	}
}
