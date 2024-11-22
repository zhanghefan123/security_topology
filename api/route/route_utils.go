package route

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
)

// GetNormalNodeFromGraphNode 从图节点获取普通节点
func GetNormalNodeFromGraphNode(graphNode graph.Node) (*normal_node.NormalNode, error) {
	currentAbstract, ok := graphNode.(*node.AbstractNode)
	if !ok {
		return nil, fmt.Errorf("convert to normal node failed")
	}
	normalNode, err := currentAbstract.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode: %w", err)
	}
	return normalNode, nil
}

func GetAbstractLink(hopList []graph.Node, sourceIndex, targetIndex int, linksMap *map[string]map[string]*link.AbstractLink) (*link.AbstractLink, int, error) {
	var err error
	var sourceNormal *normal_node.NormalNode
	var targetNormal *normal_node.NormalNode
	source := hopList[sourceIndex]
	target := hopList[targetIndex]
	sourceNormal, err = GetNormalNodeFromGraphNode(source)
	if err != nil {
		return nil, -1, fmt.Errorf("GetNormalNodeFromGraphNode err: %w", err)
	}
	targetNormal, err = GetNormalNodeFromGraphNode(target)
	if err != nil {
		return nil, -1, fmt.Errorf("GetNormalNodeFromGraphNode err: %w", err)
	}
	// 找到相应的链路 -> 带有方向的
	abstractLink := (*linksMap)[sourceNormal.ContainerName][targetNormal.ContainerName]
	if abstractLink != nil {
		return abstractLink, abstractLink.SourceInterface.LinkIdentifier, nil
	} else {
		abstractLink = (*linksMap)[targetNormal.ContainerName][sourceNormal.ContainerName]
		return abstractLink, abstractLink.TargetInterface.LinkIdentifier, nil
	}
}
