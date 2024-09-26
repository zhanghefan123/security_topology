package dir

import (
	"fmt"
	"os"
	"path/filepath"
)

// Generate 进行路径上的文件夹的生成
func Generate(generatePath string) error {
	if !filepath.IsAbs(generatePath) {
		generatePath, _ = filepath.Abs(generatePath)
	}
	if _, err := os.Stat(generatePath); os.IsNotExist(err) {
		err = os.MkdirAll(generatePath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir failed for reason %v", err)
		}
	}
	return nil
}
