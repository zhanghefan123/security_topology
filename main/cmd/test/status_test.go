package test

import (
	"testing"
	"zhanghefan123/security_topology/main/cmd/images"
)

func TestStatus(t *testing.T) {
	images.RetrieveStatus()
}
