package test

import (
	"fmt"
	"os"
	"testing"
	"zhanghefan123/security_topology/modules/utils/dir"
)

func TestContextManager(t *testing.T) {
	err := dir.WithContextManager("../", func() error {
		currentDir, _ := os.Getwd()
		fmt.Println(currentDir)
		return nil
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
}
