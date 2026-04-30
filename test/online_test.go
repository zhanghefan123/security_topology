package test

import (
	"fmt"
	"math"
	"testing"
	"zhanghefan123/security_topology/cmd/online_securest_path"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

func TestOnlineSecurestPath(t *testing.T) {
	online_securest_path.StartSimulation(online_securest_path.SimulationGraphPath,
		online_securest_path.ExperimentResultsDir, types.SimAlgorithm_OSMD)
}

func TestDecrease(t *testing.T) {
	learningRate := 0.2
	initialProb := 0.9
	for _ = range 5 {
		initialProb = initialProb * math.Exp(-learningRate*1)
		fmt.Println(initialProb)
	}
}
