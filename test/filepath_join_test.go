package test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestFilepathJoin(t *testing.T) {
	result := filepath.Join("/home/zhf", "satellite1", ".conf")
	fmt.Println(result)
}
