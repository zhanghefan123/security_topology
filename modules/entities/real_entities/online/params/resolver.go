package params

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/types"
)

// ResolveSimNodeType 进行节点类型的解析
func ResolveSimNodeType(typeString string) (types.SimNetworkNodeType, error) {
	var result types.SimNetworkNodeType
	if value, ok := types.SimNetworkNodeType_value[typeString]; ok {
		result = types.SimNetworkNodeType(value)
		return result, nil
	}
	return types.SimNetworkNodeType_NormalRouter, fmt.Errorf("cannot resolve node type")
}

// ResolveSimNodeName 进行节点类型的解析
func ResolveSimNodeName(nodeParam *SimNodeParam) (string, error) {
	nodeType, err := ResolveSimNodeType(nodeParam.Type)
	if err != nil {
		return "", fmt.Errorf("resolve node type failed, %s", err.Error())
	}
	nodeName := fmt.Sprintf("%s-%d", nodeType.String(), nodeParam.Index)
	return nodeName, nil
}
