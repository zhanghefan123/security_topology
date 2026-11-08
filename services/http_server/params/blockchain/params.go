package blockchain

import (
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/services/http_server/params/blockchain/basic"
	"zhanghefan123/security_topology/services/http_server/params/blockchain/chainmaker"
	"zhanghefan123/security_topology/services/http_server/params/blockchain/fabric"
	"zhanghefan123/security_topology/services/http_server/params/blockchain/fisco_bcos"
)

// BlockChainParams 联盟链相关参数
type BlockChainParams struct {
	BasicBlockChainParmas basic.BasicParams           `json:"basic_block_chain_params"`
	ChainMakerParams      chainmaker.ChainMakerParams `json:"chain_maker_params"`
	FabricParams          fabric.FabricParams         `json:"fabric_params"`
	FiscoBcosParams       fisco_bcos.FiscoBcosParams  `json:"fisco_bcos_params"`
}

func ResolveBlockChainType(blockChainTypeString string) types.ChainType {
	if blockChainTypeString == "长安链" {
		return types.ChainType_ChainMaker
	} else if blockChainTypeString == "fabric" {
		return types.ChainType_HyperledgerFabric
	} else if blockChainTypeString == "fisco-bcos" {
		return types.ChainType_FiscoBcos
	} else {
		return types.ChainType_NonChain
	}
}
