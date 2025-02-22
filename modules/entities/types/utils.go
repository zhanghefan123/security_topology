package types

import "fmt"

// GetPrefix 根据节点类型进行前缀的获取
func GetPrefix(typ NetworkNodeType) string {
	if typ == NetworkNodeType_NormalSatellite {
		return "ns"
	} else if typ == NetworkNodeType_GroundStation {
		return "gs"
	} else if typ == NetworkNodeType_EtcdService {
		return "es"
	} else if typ == NetworkNodeType_PositionService {
		return "ps"
	} else if typ == NetworkNodeType_Router {
		return "r"
	} else if typ == NetworkNodeType_NormalNode {
		return "nn"
	} else if typ == NetworkNodeType_ConsensusNode {
		return "cn"
	} else if typ == NetworkNodeType_ChainMakerNode {
		return "cm"
	} else if typ == NetworkNodeType_MaliciousNode {
		return "mn"
	} else if typ == NetworkNodeType_LirNode {
		return "ln"
	} else if typ == NetworkNodeType_Entrance {
		return "en"
	}
	return ""
}

// ResolveNodeType 进行节点类型的解析
func ResolveNodeType(typeString string) (*NetworkNodeType, error) {
	var result NetworkNodeType
	if value, ok := NetworkNodeType_value[typeString]; ok {
		result = NetworkNodeType(value)
		return &result, nil
	}
	return nil, fmt.Errorf("cannot resolve node type")
}
