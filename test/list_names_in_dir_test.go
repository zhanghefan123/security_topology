package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/utils/dir"
)

func TestListNamesInDir(t *testing.T) {
	allFileNames, err := dir.ListFileNames("./")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(allFileNames)
}
