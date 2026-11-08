package dir

import (
	"fmt"
	"os"
	"path/filepath"
)

func DeleteFileListInDir(directory string, fileNameList []string) {
	fileNameSet := make(map[string]struct{}, len(fileNameList))
	for _, fileName := range fileNameList {
		fileNameSet[fileName] = struct{}{}
	}

	err := filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := fileNameSet[d.Name()]; ok {
			if err = os.Remove(path); err != nil {
				return err
			}
			fmt.Printf("Deleted: %s\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("DeleteFileList error: %v\n", err)
	}
}
