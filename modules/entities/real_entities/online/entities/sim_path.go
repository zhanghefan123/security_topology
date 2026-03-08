package entities

import (
	"fmt"
	"math"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimPath struct {
	PathId                 int                           // 路径的唯一 id
	NodeList               []*SimAbstractNode            // 可以是两种不同的 router
	PvRouters              []*SimPathValidationRouter    // 这条路径上所有的 pv router
	DirectedPvLinks        []*SimDirectedPvLink          // 这条路径上所有的 directed pv link
	DirectedPvLinksMapping map[string]*SimDirectedPvLink // 从 description 到 directed pv link 的 mapping
	Weights                []float64                     // 这条路径的历史权重
	ExploreProbabilities   []float64                     // 根据 weight 算出来的当前应该选的路径的概率
	Gains                  []float64                     // 这条路径的增益，增益是根据这条路径上所有 directed pv link 的 gain 算出来的
	Score                  float64                       // 用来进行排序的
	Description            string                        // 对这条路径的唯一描述
}

func NewSimPath() *SimPath {
	return &SimPath{
		NodeList:               make([]*SimAbstractNode, 0),
		PvRouters:              make([]*SimPathValidationRouter, 0),
		DirectedPvLinks:        make([]*SimDirectedPvLink, 0),
		DirectedPvLinksMapping: make(map[string]*SimDirectedPvLink),
		Weights:                make([]float64, 0),
		ExploreProbabilities:   make([]float64, 0),
		Gains:                  make([]float64, 0),
	}
}

func (simPath *SimPath) GetPathDescription() (string, error) {
	if len(simPath.NodeList) == 0 {
		return "", fmt.Errorf("get path description failed due to empty node list")
	}
	finalString := ""
	for index, node := range simPath.NodeList {
		simNodeName, err := node.GetSimNodeName()
		if err != nil {
			return "", fmt.Errorf("get path description failed due to: %v", err)
		}
		if index != (len(simPath.NodeList) - 1) {
			finalString += fmt.Sprintf("%s->", simNodeName)
		} else {
			finalString += simNodeName
		}
	}
	return finalString, nil
}

func (simPath *SimPath) GetDirectedPvLinks() ([]*SimDirectedPvLink, map[string]*SimDirectedPvLink) {
	return simPath.DirectedPvLinks, simPath.DirectedPvLinksMapping
}

// CalculateGain calculate the gain of this path based on the weights of the edges in this path
func (simPath *SimPath) CalculateGain(epoch int) float64 {
	// 问题1: 这里应该到底怎么进行增益的计算

	// 这里是有问题的 (path 的 gain 默认设置为了 1)
	//pathGain := 1.00
	var pathGain = 0.0
	for _, directedPvLink := range simPath.DirectedPvLinks {
		pathGain = pathGain + directedPvLink.RectifiedGains[epoch]
	}
	return pathGain
}

// FindNextPvRouterAndIndex 从 start index 开始找到第一个 pv router 以及它的 index,并且在这个 pv router 之前只能有一个 normal router，如果有超过一个 normal router 就返回 error
func FindNextPvRouterAndIndex(simPath *SimPath, startIndex int) (*SimAbstractNode, int, *SimAbstractNode, error) {
	var normalRouter *SimAbstractNode
	numberOfNormalRouters := 0
	for index := startIndex + 1; index < len(simPath.NodeList); index++ {
		if simPath.NodeList[index].Type == types.SimNetworkNodeType_NormalRouter {
			normalRouter = simPath.NodeList[index]
			numberOfNormalRouters++
			if numberOfNormalRouters > 1 {
				return nil, -1, nil, fmt.Errorf("too many normal routers within pvlink")
			}
		}
		if simPath.NodeList[index].Type == types.SimNetworkNodeType_PathValidationRouter {
			return simPath.NodeList[index], index, normalRouter, nil
		}
	}
	return nil, -1, nil, nil
}

// UpdateInfo FillDirectedPvLinksMapping and DirectedPvLinks
func (simPath *SimPath) UpdateInfo(nameToPvLinkMapping map[string]*SimDirectedPvLink) error {
	directedPvLinks := make([]*SimDirectedPvLink, 0)
	directedPvLinksMapping := make(map[string]*SimDirectedPvLink)

	startIndex := 0
	var sourceNode, targetNode, intermediateNode *SimAbstractNode
	var sourceNodeName, intermediateNodeName, targetNodeName string
	var err error
	for {
		sourceNode = simPath.NodeList[startIndex]
		targetNode, startIndex, intermediateNode, err = FindNextPvRouterAndIndex(simPath, startIndex)
		if err != nil {
			return fmt.Errorf("get directed pv links failed due to: %v", err)
		}
		if targetNode == nil {
			break
		} else {
			sourceNodeName, err = sourceNode.GetSimNodeName()
			if err != nil {
				return fmt.Errorf("get directed pv links failed due to: %v", err)
			}
			intermediateNodeName, err = intermediateNode.GetSimNodeName()
			if err != nil {
				return fmt.Errorf("get directed pv links failed due to: %v", err)
			}
			targetNodeName, err = targetNode.GetSimNodeName()
			if err != nil {
				return fmt.Errorf("get directed pv links failed due to: %v", err)
			}
			// get description of the directed pv link
			pvLinkDescription := fmt.Sprintf("%s->%s->%s", sourceNodeName, intermediateNodeName, targetNodeName)

			// get pv link from mapping
			directedPvLink := nameToPvLinkMapping[pvLinkDescription]
			// update pv router list
			pvRouter, ok := targetNode.ActualNode.(*SimPathValidationRouter)
			if !ok {
				return fmt.Errorf("target node is not a path validation router")
			}
			simPath.PvRouters = append(simPath.PvRouters, pvRouter)
			// update pv link list
			directedPvLinks = append(directedPvLinks, directedPvLink)
			if _, ok = directedPvLinksMapping[pvLinkDescription]; !ok {
				directedPvLinksMapping[pvLinkDescription] = directedPvLink
			} else {
				return fmt.Errorf("get directed pv links failed due to duplicate directed pv link description: %s", pvLinkDescription)
			}
		}
	}
	simPath.DirectedPvLinks = directedPvLinks
	simPath.DirectedPvLinksMapping = directedPvLinksMapping
	return nil
}

// CalculateScore 计算链路的得分 (换掉一条后面的链路更不重要，得分越高的链路, 越靠前)
func (simPath *SimPath) CalculateScore() {
	score := 0.0
	//destinationIndex := simPath.NodeList[len(simPath.NodeList)-1].ActualNode.(*SimPathValidationRouter).Index
	for index, directedPvLink := range simPath.DirectedPvLinks {
		intermediateNode, ok := directedPvLink.Intermediate.ActualNode.(*SimNormalRouter)
		if ok {
			score += math.Pow((intermediateNode.StartCorruptRatio+intermediateNode.EndCorruptRatio)/2, float64(index+1))
		}
		// 越靠后面的链路的优先级越高
	}
	//fmt.Printf("path desc: %s, score: %f\n", simPath.Description, score)
	simPath.Score = score
}

func SamePath(pathA, pathB *SimPath) bool {
	return pathA.Description == pathB.Description
}
