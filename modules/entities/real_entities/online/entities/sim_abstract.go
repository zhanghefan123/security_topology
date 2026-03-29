package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"

	"gonum.org/v1/gonum/graph"
)

type SimAbstractNode struct {
	graph.Node
	Type       types.SimNetworkNodeType
	ActualNode interface{}
}

func NewSimAbstract(nodeType types.SimNetworkNodeType, actualNode interface{}, graphNode graph.Node) *SimAbstractNode {
	return &SimAbstractNode{
		Node:       graphNode,
		Type:       nodeType,
		ActualNode: actualNode,
	}
}

func (simAbstractNode *SimAbstractNode) GetSimNodeBaseFromAbstract() (*SimNodeBase, error) {
	switch simAbstractNode.Type {
	case types.SimNetworkNodeType_EndHost:
		if endHost, ok := simAbstractNode.ActualNode.(*SimEndHost); ok {
			return endHost.SimNodeBase, nil
		}
	case types.SimNetworkNodeType_NormalRouter:
		if simNormalRouter, ok := simAbstractNode.ActualNode.(*SimNormalRouter); ok {
			return simNormalRouter.SimNodeBase, nil
		}
	case types.SimNetworkNodeType_PathValidationRouter:
		if simPathValidationRouter, ok := simAbstractNode.ActualNode.(*SimPathValidationRouter); ok {
			return simPathValidationRouter.SimNodeBase, nil
		}
	}
	return nil, fmt.Errorf("cannot get simNoedBase from abstract")
}

func (simAbstractNode *SimAbstractNode) GetPotentialFromAbstract() (float64, error) {
	switch simAbstractNode.Type {
	case types.SimNetworkNodeType_EndHost:
		if endHost, ok := simAbstractNode.ActualNode.(*SimEndHost); ok {
			return endHost.Potential, nil
		}
	case types.SimNetworkNodeType_NormalRouter:
		return -1, fmt.Errorf("cannot get potential from normal router")
	case types.SimNetworkNodeType_PathValidationRouter:
		if simPathValidationRouter, ok := simAbstractNode.ActualNode.(*SimPathValidationRouter); ok {
			return simPathValidationRouter.Potential, nil
		}
	}
	return -1, fmt.Errorf("cannot get simNoedBase from abstract")
}

func (simAbstractNode *SimAbstractNode) GetSimNodeName() (string, error) {
	simNodeBase, err := simAbstractNode.GetSimNodeBaseFromAbstract()
	if err != nil {
		return "", fmt.Errorf("get sim node failed: %v", err)
	} else {
		return simNodeBase.NodeName, nil
	}
}
