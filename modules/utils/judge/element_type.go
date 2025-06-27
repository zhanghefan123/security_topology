package judge

import "zhanghefan123/security_topology/modules/entities/types"

func IsBlockChainType(networkNodeType types.NetworkNodeType) bool {
	if networkNodeType == types.NetworkNodeType_ChainMakerNode {
		return true
	}
	if networkNodeType == types.NetworkNodeType_FabricOrderNode {
		return true
	}
	return false
}
