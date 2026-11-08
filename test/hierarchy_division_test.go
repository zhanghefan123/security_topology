package test

import (
	"fmt"
	"testing"
	atlasinit "zhanghefan123/security_topology/modules/entities/real_entities/atlas/init"
)

func TestHierarchyDivision(t *testing.T) {
	err := atlasinit.GenerateDifferentNumberOfPaths()
	if err != nil {
		fmt.Printf("generate different number of paths error: %v", err)
	}
}
