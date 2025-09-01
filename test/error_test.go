package test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func FunctionReturnError() (err error) {
	err = nil
	return errors.New("test error")
}

func TestError(t *testing.T) {
	result := filepath.Join("/src", fmt.Sprintf("nodes/127.0.0.1/node%d/nodes.json", 1))
	fmt.Println(result)
}
