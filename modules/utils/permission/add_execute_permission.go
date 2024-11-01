package permission

import (
	"fmt"
	"os"
)

// AddExecutePermission 为文件添加可执行的权限
func AddExecutePermission(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot add execute permission to file %s, because %w", filePath, err)
	}
	err = os.Chmod(filePath, fileInfo.Mode()|0111) // 添加可执行权限
	if err != nil {
		return fmt.Errorf("cannot add execute permission to file %s, because %w", filePath, err)
	}
	return nil
}
