package file

import (
	"fmt"
	"io"
	"os"
)

func ReadFile(fileName string) (string, error) {
	var bytesContent []byte
	file, err := os.Open(fileName)
	defer func() {
		errClose := file.Close()
		if err == nil {
			err = errClose
		}
	}()
	if err != nil {
		return "", fmt.Errorf("open file %s failed due to %v\n", fileName, err)
	}
	bytesContent, err = io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read file %s failed due to %v\n", fileName, err)
	}
	stringContent := string(bytesContent)
	return stringContent, nil
}
