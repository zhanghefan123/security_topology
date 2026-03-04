package entities

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"zhanghefan123/security_topology/modules/entities/types"
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

func (simAbstractNode *SimAbstractNode) GetSimNodeName() (string, error) {
	simNodeBase, err := simAbstractNode.GetSimNodeBaseFromAbstract()
	if err != nil {
		return "", fmt.Errorf("get sim node failed: %v", err)
	} else {
		return simNodeBase.NodeName, nil
	}
}
