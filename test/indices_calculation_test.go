package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/calculation"
)

func TestIndiciesCalculation(t *testing.T) {
	err := calculation.GenerateDifferentPaths()
	if err != nil {
		fmt.Printf("err due to: %v", err)
	}
	//calculation.Test()
}
