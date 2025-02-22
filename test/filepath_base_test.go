package test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestFilePathBase(t *testing.T) {
	result := filepath.Base("test.txt")
	fmt.Println(result)
}
