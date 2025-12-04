package dir

import (
	"os"
	"path/filepath"
)

// ListFileNames 列出 Dir 之中的所有文件
func ListFileNames(dir string) ([]string, error) {
	allAvailableTopologyNames := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			allAvailableTopologyNames = append(allAvailableTopologyNames, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return allAvailableTopologyNames, nil
	}
}
