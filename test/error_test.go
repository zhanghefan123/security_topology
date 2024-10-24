package test

import (
	"errors"
	"fmt"
	"testing"
)

func FunctionReturnError() (err error) {
	err = nil
	return errors.New("test error")
}

func TestError(t *testing.T) {
	if err := FunctionReturnError(); err != nil {
		fmt.Println(err)
	}
}
