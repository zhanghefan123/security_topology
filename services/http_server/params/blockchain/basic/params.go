package basic

import "zhanghefan123/security_topology/modules/entities/types"

type BasicParams struct {
	ConsensusThreadCount int    `json:"consensus_thread_count"`
	BlockchainTypeStr    string `json:"blockchain_type"`
	BlockchainType       types.ChainType
	ConsensusType        string `json:"consensus_type"`
	StartDefence         bool   `json:"start_defence"`
	DdosWarningRate      int    `json:"ddos_warning_rate"`
}
