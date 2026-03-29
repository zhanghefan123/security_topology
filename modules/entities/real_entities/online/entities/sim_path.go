package entities

import (
	"fmt"
	"math"
	"math/rand"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimPath struct {
	PathId                  int                                 // 路径的唯一 id
	NodeList                []*SimAbstractNode                  // 可以是两种不同的 router
	PvRouters               []*SimPathValidationRouter          // 这条路径上所有的 pv router
	PvRoutersMapping        map[string]*SimPathValidationRouter // 这条路径上所有的 pv router 的 mapping
	DirectedAbsLinks        []*SimDirectedAbsLink               // 这条路径上所有的 directed pv link
	DirectedAbsLinksMapping map[string]*SimDirectedAbsLink      // 从 description 到 directed pv link 的 mapping
	NodeNameToIndexMapping  map[string]int                      // 从节点的名称到对应索引
	Weights                 []float64                           // 这条路径的历史权重
	ExploreProbabilities    []float64                           // 根据 weight 算出来的当前应该选的路径的概率
	Gains                   []float64                           // 这条路径的增益，增益是根据这条路径上所有 directed pv link 的 gain 算出来的
	Score                   float64                             // 用来进行排序的
	Description             string                              // 对这条路径的唯一描述
	AverageCorruptRatio     float64
	AverageDropRatio        float64
}

func NewSimPath() *SimPath {
	return &SimPath{
		NodeList:                make([]*SimAbstractNode, 0),
		PvRouters:               make([]*SimPathValidationRouter, 0),
		PvRoutersMapping:        make(map[string]*SimPathValidationRouter),
		DirectedAbsLinks:        make([]*SimDirectedAbsLink, 0),
		DirectedAbsLinksMapping: make(map[string]*SimDirectedAbsLink),
		NodeNameToIndexMapping:  make(map[string]int),
		Weights:                 make([]float64, 0),
		ExploreProbabilities:    make([]float64, 0),
		Gains:                   make([]float64, 0),
		AverageCorruptRatio:     0,
		AverageDropRatio:        0,
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

func (simPath *SimPath) GetDirectedAbsLinks() ([]*SimDirectedAbsLink, map[string]*SimDirectedAbsLink) {
	return simPath.DirectedAbsLinks, simPath.DirectedAbsLinksMapping
}

func (simPath *SimPath) GetPvRouters() ([]*SimPathValidationRouter, map[string]*SimPathValidationRouter) {
	return simPath.PvRouters, simPath.PvRoutersMapping
}

// CalculateGain calculate the gain of this path based on the weights of the edges in this path
func (simPath *SimPath) CalculateGain(epoch int, mode types.GainCalculationMode) float64 {
	// 问题1: 这里应该到底怎么进行增益的计算
	var pathGain = 0.0
	if mode == types.GainCalculationMode_SumEdgeGains {
		for _, directedPvLink := range simPath.DirectedAbsLinks {
			pathGain = pathGain + directedPvLink.RectifiedGains[epoch]
		}
		return pathGain
	} else {
		for _, directedPvLink := range simPath.DirectedAbsLinks {
			pathGain = pathGain + directedPvLink.RectifiedGains[epoch]
		}
		for _, pvRouter := range simPath.PvRouters {
			pathGain = pathGain + pvRouter.RectifiedGains[epoch]
		}
		return pathGain
	}
}

// FindNextPvRouterOrEndHost 从 start index 开始找到第一个 (pv router 或者 end host) 以及它的 index,并且在这个 pv router 之前只能有一个 normal router，如果有超过一个 normal router 就返回 error
func FindNextPvRouterOrEndHost(simPath *SimPath, startIndex int) (*SimAbstractNode, int, *SimAbstractNode, error) {
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
		if (simPath.NodeList[index].Type == types.SimNetworkNodeType_PathValidationRouter) || (simPath.NodeList[index].Type == types.SimNetworkNodeType_EndHost) {
			return simPath.NodeList[index], index, normalRouter, nil
		}
	}
	return nil, -1, nil, nil
}

// UpdateInfo FillDirectedPvLinksMapping and DirectedAbsLinks
func (simPath *SimPath) UpdateInfo(nameToPvLinkMapping map[string]*SimDirectedAbsLink) error {
	directedPvLinks := make([]*SimDirectedAbsLink, 0)
	directedPvLinksMapping := make(map[string]*SimDirectedAbsLink)

	startIndex := 0
	currentNodeIndex := 0
	var sourceNode, targetNode, intermediateNode *SimAbstractNode
	var sourceNodeName, intermediateNodeName, targetNodeName string
	var err error
	for {
		sourceNode = simPath.NodeList[startIndex]
		targetNode, startIndex, intermediateNode, err = FindNextPvRouterOrEndHost(simPath, startIndex)
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
			// record mapping
			simPath.NodeNameToIndexMapping[targetNodeName] = currentNodeIndex
			currentNodeIndex++
			// get pv link from mapping
			directedPvLink := nameToPvLinkMapping[pvLinkDescription]
			// update pv router list
			if targetNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				pvRouter, ok := targetNode.ActualNode.(*SimPathValidationRouter)
				if !ok {
					return fmt.Errorf("target node is not a path validation router")
				}
				simPath.PvRouters = append(simPath.PvRouters, pvRouter)
				simPath.PvRoutersMapping[pvRouter.NodeName] = pvRouter
			}
			// update pv link list
			directedPvLinks = append(directedPvLinks, directedPvLink)
			if _, ok := directedPvLinksMapping[pvLinkDescription]; !ok {
				directedPvLinksMapping[pvLinkDescription] = directedPvLink
			} else {
				return fmt.Errorf("get directed pv links failed due to duplicate directed pv link description: %s", pvLinkDescription)
			}
		}
	}
	simPath.DirectedAbsLinks = directedPvLinks
	simPath.DirectedAbsLinksMapping = directedPvLinksMapping
	return nil
}

// CalculateScore 计算链路的得分 (换掉一条后面的链路更不重要，得分越高的链路, 越靠前)
func (simPath *SimPath) CalculateScore() {
	score := 0.0
	//destinationIndex := simPath.NodeList[len(simPath.NodeList)-1].ActualNode.(*SimPathValidationRouter).Index
	for index, directedPvLink := range simPath.DirectedAbsLinks {
		intermediateNode, ok := directedPvLink.Intermediate.ActualNode.(*SimNormalRouter)
		if ok {
			score += math.Pow((intermediateNode.StartCorruptRatio+intermediateNode.EndCorruptRatio)/2, float64(index+1))
		}
		// 越靠后面的链路的优先级越高
	}
	//fmt.Printf("path desc: %s, score: %f\n", simPath.Description, score)
	simPath.Score = score
}

//func (simPath *SimPath) ChooseSamplePvRouter() (*SimPathValidationRouter, int) {
//	equalProbability := 1.0 / float64(len(simPath.PvRouters))
//	sampleProbabilities := make([]float64, 0)
//	for _ = range len(simPath.PvRouters) {
//		sampleProbabilities = append(sampleProbabilities, equalProbability)
//	}
//	idx := probs.SampleDiscrete(sampleProbabilities)
//	return simPath.PvRouters[idx], idx
//}

func (simPath *SimPath) GenerateSampleSequence(batchSize int, simlulationStrategy types.SimStrategy) ([]int, error) {
	sequence := make([]int, batchSize)
	if simlulationStrategy == types.SimStrategy_PerBatchBloomFilter {
		numRouters := len(simPath.PvRouters)
		// 1. 确定性分配：确保每个 Router 分配到的次数尽可能相等
		for i := 0; i < batchSize; i++ {
			sequence[i] = i % numRouters
		}

		// 2. 随机洗牌：打乱顺序，让攻击者无法预测
		// 使用加密安全的随机数种子更佳
		rand.Shuffle(len(sequence), func(i, j int) {
			sequence[i], sequence[j] = sequence[j], sequence[i]
		})

		return sequence, nil
	} else if simlulationStrategy == types.SimStrategy_PerPacketAck {
		numNodes := len(simPath.PvRouters) + 1 // +1 是最后的目的节点也可能是采样的节点
		// 1. 确定性分配：确保每个 Router 分配到的次数尽可能相等
		for i := 0; i < batchSize; i++ {
			sequence[i] = i % numNodes
		}

		// 2. 随机洗牌：打乱顺序，让攻击者无法预测
		// 使用加密安全的随机数种子更佳
		rand.Shuffle(len(sequence), func(i, j int) {
			sequence[i], sequence[j] = sequence[j], sequence[i]
		})

		return sequence, nil
	} else {
		return sequence, fmt.Errorf("not supported simulation type")
	}
}

func SamePath(pathA, pathB *SimPath) bool {
	return pathA.Description == pathB.Description
}
