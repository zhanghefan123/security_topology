package entities

import (
	"fmt"
)

type SimPath struct {
	NodeList               []*SimAbstractNode                       //  可以是两种不同的 router
	DirectedPvLinks        []*SimDirectedPvLink                     // 这条路径上所有的 directed pv link
	DirectedPvLinksMapping map[string]map[string]*SimDirectedPvLink // 从 source name 和 target name 到 directed pv link 的 mapping
	Weights                []float64                                // 这条路径的历史权重
	Probability            float64                                  // 根据 weight 算出来的当前应该选的路径的概率
	Gain                   float64                                  // 这条路径的增益，增益是根据这条路径上所有 directed pv link 的 gain 算出来的
	PathDescription        string                                   // 对这条路径的唯一描述
}

func NewSimPath() *SimPath {
	return &SimPath{
		NodeList:               make([]*SimAbstractNode, 0),
		DirectedPvLinks:        make([]*SimDirectedPvLink, 0),
		DirectedPvLinksMapping: make(map[string]map[string]*SimDirectedPvLink),
		Weights:                make([]float64, 0),
	}
}

func (simPath *SimPath) GetPathDescription() (string, error) {
	if len(simPath.NodeList) == 0 {
		return "", fmt.Errorf("get path description failed due to empty node list")
	}
	finalString := ""
	for _, node := range simPath.NodeList {
		simNodeBase, err := node.GetSimNodeBaseFromAbstract()
		if err != nil {
			return "", fmt.Errorf("get path description failed due to: %v", err)
		}
		finalString += fmt.Sprintf("%s->", simNodeBase.NodeName)
	}
	return finalString, nil
}

func (simPath *SimPath) GetDirectedPvLinks(nameToLinkMapping map[string]map[string]*SimDirectedPvLink) ([]*SimDirectedPvLink, map[string]map[string]*SimDirectedPvLink, error) {
	directedPvLinks := make([]*SimDirectedPvLink, 0)
	directedPvLinksMapping := make(map[string]map[string]*SimDirectedPvLink)

	startIndex := 0
	var sourceNode, targetNode *SimAbstractNode
	for {
		sourceNode = simPath.NodeList[startIndex]
		targetNode, startIndex = FindNextPvRouterAndIndex(simPath, startIndex)
		if targetNode == nil {
			break
		} else {
			sourceNodeName, err := sourceNode.GetSimNodeName()
			if err != nil {
				return nil, nil, fmt.Errorf("get directed pv links failed due to: %v", err)
			}
			targetNodeName, err := targetNode.GetSimNodeName()
			if err != nil {
				return nil, nil, fmt.Errorf("get directed pv links failed due to: %v", err)
			}
			directedPvLink := nameToLinkMapping[sourceNodeName][targetNodeName]
			directedPvLinks = append(directedPvLinks, directedPvLink)
			if _, ok := directedPvLinksMapping[sourceNodeName]; !ok {
				directedPvLinksMapping[sourceNodeName] = make(map[string]*SimDirectedPvLink)
			}
			directedPvLinksMapping[sourceNodeName][targetNodeName] = directedPvLink
		}
	}
	return directedPvLinks, directedPvLinksMapping, nil
}

// CalculateGain calculate the gain of this path based on the weights of the edges in this path
func (simPath *SimPath) CalculateGain() {
	pathGain := 0.0
	for _, directedPvLink := range simPath.DirectedPvLinks {
		pathGain += directedPvLink.RectifiedGain
	}
	simPath.Gain = pathGain
}
