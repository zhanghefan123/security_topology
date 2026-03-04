package steps

import (
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"fmt"
	"math"
	"math/rand"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/types"
)

// Start 进行 simulator 的运行
func (s *Simulator) Start() error {
	// Phase 1: Initialization
	for _, simDirectedPvLink := range s.SimGraph.SimDirectedPvLinks {
		simDirectedPvLink.Weights = append(simDirectedPvLink.Weights, 1)
		simDirectedPvLink.IllegalRatios = append(simDirectedPvLink.IllegalRatios, 0)
	}
	totalPathWeights := 0.0
	for _, simPath := range s.SimGraph.AvailablePaths {
		simPath.Weights = append(simPath.Weights, 1)
		totalPathWeights += simPath.Weights[0]
	}
	s.SimGraph.TotalPathWeights = totalPathWeights

	// record the path selected in previous epoch, and the session id used in previous epoch
	var selectedPathInPreviousEpoch *entities.SimPath = nil
	var previousEpochSessionId = ""

	// Phase 2: iteration
	for epoch := 1; epoch < s.SimulatorParams.NumberOfEpochs; epoch++ {
		// Phase 2.1 calculate the path probability
		for _, simPath := range s.SimGraph.AvailablePaths {
			if s.SimGraph.IsCoveragePath(simPath) {
				probabilityOfSimpath := (1-s.SimulatorParams.ExploreRate)*(simPath.Weights[epoch-1])/(s.SimGraph.TotalPathWeights) + s.SimulatorParams.ExploreRate/float64(len(s.SimGraph.CoveragePaths))
				simPath.Probability = probabilityOfSimpath
			} else {
				probabilityOfSimpath := (1 - s.SimulatorParams.ExploreRate) * (simPath.Weights[epoch-1]) / (s.SimGraph.TotalPathWeights)
				simPath.Probability = probabilityOfSimpath
			}
		}
		// Phase 2.2 calculate the probability of choosing each edge e
		// Phase 2.2.1 clear the explore probability calculated in last epoch
		s.SimGraph.ClearAllEdgesProbabilities()
		// Phase 2.2.2 calculate the explore probability for each edge based on the path probability and the illegal ratio in last epoch
		for _, simPath := range s.SimGraph.AvailablePaths {
			previousIllegalRatio := 0.0
			lowerBoundReachProbability := 0.0

			for _, directedPvLink := range simPath.DirectedPvLinks {
				if previousIllegalRatio == 0.0 {
					previousIllegalRatio = 1 - directedPvLink.IllegalRatios[epoch-1]
				} else {
					previousIllegalRatio = previousIllegalRatio * directedPvLink.IllegalRatios[epoch-1]
				}

				if lowerBoundReachProbability == 0.0 {
					lowerBoundReachProbability = 1 - 0.1
				} else {
					lowerBoundReachProbability = lowerBoundReachProbability * (1 - 0.1)
				}
				edgeExploreProb := max(previousIllegalRatio*directedPvLink.IllegalRatios[epoch-1], lowerBoundReachProbability)

				// modify the explore probability of this edge
				directedPvLink.ExploreProbability += edgeExploreProb
			}
		}

		// Phase 2.3 select the path according to the probability distribution
		pathProbabilities := make([]float64, 0)
		for _, simPath := range s.SimGraph.AvailablePaths {
			pathProbabilities = append(pathProbabilities, simPath.Probability)
		}
		selectedPathIndex := s.SampleDiscrete(pathProbabilities)
		selectedPath := s.SimGraph.AvailablePaths[selectedPathIndex]

		// Phase 2.4 pull the arm and observe the gain
		// Phase 2.4.1 get the sessionid
		var establishNewSession = false
		var currentEpochSessionId = ""
		if selectedPathInPreviousEpoch == nil {
			// 如果选择
			newSessionId := uuid.GetUUID()
			previousEpochSessionId = newSessionId
			currentEpochSessionId = newSessionId
			establishNewSession = true
			selectedPathInPreviousEpoch = selectedPath
		}
		if selectedPathInPreviousEpoch.PathDescription != selectedPath.PathDescription {
			// 进行旧的 session 的清空
			err := s.DestroySession(previousEpochSessionId, selectedPathInPreviousEpoch)
			if err != nil {
				return fmt.Errorf("destroy session failed due to %v", err)
			}
		}

		// Phase 2.4.2 establish session
		if establishNewSession {
			err := s.EstablishSession(currentEpochSessionId, selectedPath)
			if err != nil {
				return fmt.Errorf("establish session failed due to %v", err)
			}
		}
		// Phase 2.4.3 send a batch of packets
		for j := 0; j < s.SimulatorParams.NumberOfPktsPerBatch; j++ {
			// create simpacket
			simPacket := entities.CreatePacket(selectedPath, currentEpochSessionId)
			// forward packet in all on-path routers
			err := s.ForwardPacket(simPacket, selectedPath)
			if err != nil {
				return fmt.Errorf("forward packet failed due to %v", err)
			}
		}

		// Phase 2.4.4 retrieve information after sending a batch of packets
		counters, err := s.RetrieveCounters(currentEpochSessionId, selectedPath)
		if err != nil {
			return fmt.Errorf("failed to retrieve counters for sessionId %s", currentEpochSessionId)
		}

		// Phase 2.4.5 retrieve directed pv links
		directedPvLinks, directedPvLinksMapping, err := selectedPath.GetDirectedPvLinks(s.SimGraph.SimDirectedPvLinksMapping)
		if err != nil {
			return fmt.Errorf("failed to retrieve pv links for sessionId %s", currentEpochSessionId)
		}

		// Phase 2.5 calculate legal and illegal ratio
		for index, counter := range counters {
			estimatedIllegalRatio := (float64(counter.IllegalPackets) + float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor) / (float64(counter.IllegalPackets+counter.LegalPackets) + 2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
			estimatedLegalRatio := 1 - estimatedIllegalRatio
			directedPvLinks[index].EstimatedIllegalRatio = estimatedIllegalRatio
			directedPvLinks[index].EstimatedLegalRatio = estimatedLegalRatio
		}

		// 2.6 unbiased gain estimation
		// traverse each edge
		var sourceNodeName, targetNodeName string
		for _, directedPvLink := range s.SimGraph.SimDirectedPvLinks {
			sourceNodeName, targetNodeName, err = directedPvLink.GetSourceAndTargetNames()
			if err != nil {
				return fmt.Errorf("get source and target node names failed due to: %v", err)
			}
			if _, ok := directedPvLinksMapping[sourceNodeName][targetNodeName]; ok {
				// calculate the estimated gain
				estimatedGain := math.Pow(directedPvLink.EstimatedIllegalRatio, s.SimulatorParams.Lambda)
				// calculate the rectified gain by dividing the estimated gain by the probability of choosing this edge
				rectifiedGain := (estimatedGain + s.SimulatorParams.Bias) / directedPvLink.ExploreProbability
				// update the gain of this edge
				directedPvLink.RectifiedGain = rectifiedGain
			} else {
				// calculate the estimated gain
				rectifiedGain := s.SimulatorParams.Bias / directedPvLink.ExploreProbability
				// update the gain of this edge
				directedPvLink.RectifiedGain = rectifiedGain
			}
		}

		// 2.7 update the weights if edges and paths
		// 2.7.1 update the edge weights
		for _, directedPvLink := range s.SimGraph.SimDirectedPvLinks {
			currentEpochWeight := directedPvLink.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*directedPvLink.RectifiedGain)
			directedPvLink.Weights = append(directedPvLink.Weights, currentEpochWeight)
		}
		// 2.7.2 update the path weights
		for _, simPath := range s.SimGraph.AvailablePaths {
			// calculate the gain of path in this epoch
			simPath.CalculateGain()
			// calcualte the weight of path in this epoch
			currentEpochWeight := simPath.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*simPath.Gain)
			simPath.Weights = append(simPath.Weights, currentEpochWeight)
		}
		// 2.7.3 update totol path weights
		totalPathWeights = 0.0
		for _, simPath := range s.SimGraph.AvailablePaths {
			totalPathWeights = totalPathWeights + simPath.Weights[epoch]
		}
		s.SimGraph.TotalPathWeights = totalPathWeights
		// 2.7.4 forget history smoothing
		for _, simPath := range s.SimGraph.AvailablePaths {
			historyForgettingWeight := (1-s.SimulatorParams.HistoryForggetingFactor)*simPath.Weights[epoch] + s.SimulatorParams.HistoryForggetingFactor*totalPathWeights/float64(len(s.SimGraph.AvailablePaths))
			simPath.Weights[epoch] = historyForgettingWeight
		}
	}

	return nil
}

func (s *Simulator) SampleDiscrete(probabilities []float64) int {
	r := rand.Float64() // 生成 [0,1) 均匀随机数
	sum := 0.0          // 累计概率

	for i, p := range probabilities {
		sum += p      // 把概率一个个累加
		if r <= sum { // 如果随机数落在当前累计区间
			return i // 返回当前下标
		}
	}

	return len(probabilities) - 1 // 防止浮点误差兜底
}

func (s *Simulator) EstablishSession(sessionId string, simPath *entities.SimPath) error {
	for _, simAbstractNode := range simPath.NodeList {
		if simAbstractNode.Type == types.SimNetworkNodeType_NormalRouter {
			// do nothing
		} else if simAbstractNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			if pvRouter, ok := simAbstractNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				err := pvRouter.EstablishSession(sessionId)
				if err != nil {
					return fmt.Errorf("establish session failed due to %v", err)
				}
			}
		} else {
			return fmt.Errorf("unsupported path node type")
		}
	}
	return nil
}

func (s *Simulator) DestroySession(sessionId string, simPath *entities.SimPath) error {
	for _, abstractSimNode := range simPath.NodeList {
		if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
			// do nothing
		} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				err := pvRouter.DestroySession(sessionId)
				if err != nil {
					return fmt.Errorf("destroy session failed due to %v", err)
				}
			}
		} else {
			return fmt.Errorf("unsupported path node type")
		}
	}
	return nil
}

func (s *Simulator) ForwardPacket(packet *entities.SimPacket, selectedPath *entities.SimPath) error {
	for _, abstractSimNode := range selectedPath.NodeList {
		// 进行 abstractNode 的类型的判断
		if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
			if router, ok := abstractSimNode.ActualNode.(*entities.SimNormalRouter); ok {
				router.ProcessPacket(packet)
			}
		} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				err := pvRouter.ProcessPacket(packet)
				if err != nil {
					return fmt.Errorf("pv router process packet failed due to: %v", err)
				}
			}
		} else {
			return fmt.Errorf("unsupported path node type")
		}
	}
	return nil
}

func (s *Simulator) RetrieveCounters(sessionId string, selectedPath *entities.SimPath) ([]*entities.Counter, error) {
	counters := make([]*entities.Counter, 0)
	for index, abstractSimNode := range selectedPath.NodeList {
		// 这里设置为 != 0 代表的是第一个节点不用进行消息的返回
		if index != 0 {
			// 进行 abstractNode 的类型的判断
			if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
				// donothing
			} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
					// retrieve information
					counter, err := pvRouter.RetrieveInformation(sessionId)
					if err != nil {
						// 进行错误处理
						return nil, fmt.Errorf("fail to retrieve information due to: %v", err)
					}
					counters = append(counters, counter)
				}
			} else {
				return nil, fmt.Errorf("unsupported path node type")
			}
		}
	}
	return counters, nil
}
