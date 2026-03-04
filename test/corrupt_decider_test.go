package test

import (
	"fmt"
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/corrupt_decider"
)

func TestCorruptDecider(t *testing.T) {
	minRate, maxRate := 0.1, 0.3

	dropCount := 0
	totalCount := 100000

	uniformCorruptDecider, err := corrupt_decider.CreateUniformCorruptDecider(minRate, maxRate)
	if err != nil {
		fmt.Printf("create uniform corrupt decider failed: %v", err)
	}

	for index := 0; index < totalCount; index++ {
		if uniformCorruptDecider.ShouldDrop() {
			dropCount += 1
		}
	}

	fmt.Printf("actual packet drop ratio: %f", float64(dropCount)/float64(totalCount))

}
