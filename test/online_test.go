package test

import (
	"testing"
	"zhanghefan123/security_topology/cmd/online_securest_path"
)

func TestOnlineSecurestPath(t *testing.T) {
	online_securest_path.StartOnlineSecurestPathSimulation(online_securest_path.SimulationGraphPath,
		online_securest_path.ExperimentResultsDir)
}
