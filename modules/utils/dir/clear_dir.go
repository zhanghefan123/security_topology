package dir

import (
	"os"
	"path/filepath"
)

func ClearDir(dir string) error {
	// 读取目录下的所有内容
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 遍历并删除每一项
	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		err = os.RemoveAll(filePath) // 删除文件或目录
		if err != nil {
			return err
		}
	}
	return nil
}
