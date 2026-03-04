package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/utils/extract"
)

func TestExtract(t *testing.T) {
	index, err := extract.NumberFromString("LirNode-1231")
	if err != nil {
		return
	}
	fmt.Printf("%d", index)
}
