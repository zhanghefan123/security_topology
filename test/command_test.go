package test

import (
	"testing"
	"zhanghefan123/security_topology/modules/utils/execute"
)

func TestExecuteCommand(t *testing.T) {
	err := execute.Command("pwd", []string{})
	if err != nil {
		t.Error(err)
	}
}
