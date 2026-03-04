package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph"
)

func TestHierarchyDivision(t *testing.T) {
	err := graph.GenerateDifferentNumberOfPaths()
	if err != nil {
		fmt.Printf("generate different number of paths error: %v", err)
	}
}
