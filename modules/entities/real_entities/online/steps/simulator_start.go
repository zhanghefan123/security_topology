package steps

import (
	"fmt"
	"math"
	"math/rand"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"

	"chainmaker.org/chainmaker/common/v2/random/uuid"
)

// Start 进行 simulator 的运行
func (s *Simulator) Start() error {
	// Phase 1: Initialization
	// Phase 1.1 initialize the weights, illegal ratios, and gains for each simDirectedPvLink,
	for _, simDirectedPvLink := range s.SimGraph.SimDirectedPvLinks {
		simDirectedPvLink.Weights = append(simDirectedPvLink.Weights, 1)
		simDirectedPvLink.IllegalRatios = append(simDirectedPvLink.IllegalRatios, 0)
		simDirectedPvLink.DropRatios = append(simDirectedPvLink.DropRatios, 0)
		simDirectedPvLink.RectifiedGains = append(simDirectedPvLink.RectifiedGains, 0)
	}
	// Phase 1.2 initialize the weights for each simPath, and calculate the total path weights
	totalPathWeights := 0.0
	for _, simPath := range s.SimGraph.AvailablePaths {
		simPath.Weights = append(simPath.Weights, 1)
		totalPathWeights += simPath.Weights[0]
	}
	s.SimGraph.TotalPathWeights = totalPathWeights

	// Phase 2: iteration
	// record the path selected in previous epoch, and the session id used in previous epoch
	var previousEpochSelectedPath *entities.SimPath = nil
	var previousEpochSessionId = ""
	for epoch := 1; epoch <= s.SimulatorParams.NumberOfEpochs; epoch++ {
		// Phase 2.0 judge if it is necessary to execute the event
		for _, simEvent := range s.SimEvents {
			if epoch == simEvent.StartEpoch {
				for _, updateRouter := range simEvent.UpdateRouters {
					if selectedAbstractNormalRouter, ok := s.SimGraph.SimAbstractNodesMapping[updateRouter.NormalRouterName]; ok {
						var selectedNormalRouter *entities.SimNormalRouter
						if selectedNormalRouter, ok = selectedAbstractNormalRouter.ActualNode.(*entities.SimNormalRouter); ok {
							err := selectedNormalRouter.Reset(updateRouter.StartDropRatio, updateRouter.EndDropRatio,
								updateRouter.StartIllegalRatio, updateRouter.EndIllegalRatio)
							if err != nil {
								return fmt.Errorf("reset router %s err due to %w", updateRouter.NormalRouterName, err)
							}
						}
					} else {
						return fmt.Errorf("cannot find the normal router with description: %s\n", updateRouter.NormalRouterName)
					}
				}
			} else {
				continue
			}
		}

		// Phase 2.1 calculate the path probability
		for _, simPath := range s.SimGraph.AvailablePaths {
			var exploreProbability float64
			if s.SimGraph.IsCoveragePath(simPath) {
				// 除了最优路径的其他一条 coverage path 的概率为 = 0.1 / 2 + 0.9 * 这条路径的权重占比
				probabilityOfSimpath := (1-s.SimulatorParams.ExploreRate)*(simPath.Weights[epoch-1])/(s.SimGraph.TotalPathWeights) + s.SimulatorParams.ExploreRate/float64(len(s.SimGraph.CoveragePaths))
				exploreProbability = probabilityOfSimpath
			} else {
				// 路径增益 = 这条路径上所有边的增益之和
				// 路径权重 = 对路径增益的指数函数
				// 0.9 * (这条路径权重) / 所有路径权重
				probabilityOfSimpath := (1 - s.SimulatorParams.ExploreRate) * (simPath.Weights[epoch-1]) / (s.SimGraph.TotalPathWeights)
				exploreProbability = probabilityOfSimpath
			}
			simPath.ExploreProbabilities = append(simPath.ExploreProbabilities, exploreProbability)
		}
		// Phase 2.2 calculate the probability of choosing each edge e
		// Phase 2.2.1 calculate the explore probability for each edge based on the path probability and the illegal ratio in last epoch
		for _, simPath := range s.SimGraph.AvailablePaths {
			//reachProbability := 0.0
			//lowerBoundReachProbability := 0.0

			for _, directedPvLink := range simPath.DirectedPvLinks {
				// -------------------------------- 如果考虑到达概率会导致这个概率非常的低 --------------------------------
				//if reachProbability == 0.0 {
				//	reachProbability = 1 - directedPvLink.IllegalRatios[epoch-1]
				//} else {
				//	reachProbability = reachProbability * (1 - directedPvLink.IllegalRatios[epoch-1])
				//}

				//if lowerBoundReachProbability == 0.0 {
				//	lowerBoundReachProbability = s.SimulatorParams.LowerBoundLegalRatio
				//} else {
				//	lowerBoundReachProbability = lowerBoundReachProbability * s.SimulatorParams.LowerBoundLegalRatio
				//}

				//edgeExploreProb := simPath.ExploreProbabilities[epoch-1] * max(reachProbability, lowerBoundReachProbability)
				//edgeExploreProb := simPath.ExploreProbabilities[epoch-1] * reachProbability
				// -------------------------------- 如果考虑到达概率会导致这个概率非常的低 --------------------------------

				edgeExploreProb := simPath.ExploreProbabilities[epoch-1]

				// modify the explore probability of this edge
				if len(directedPvLink.ExploreProbabilities) == (epoch - 1) {
					directedPvLink.ExploreProbabilities = append(directedPvLink.ExploreProbabilities, edgeExploreProb)
				} else {
					directedPvLink.ExploreProbabilities[epoch-1] += edgeExploreProb
				}
			}
		}

		// Phase 2.3 select the path according to the probability distribution
		pathProbabilities := make([]float64, 0)
		for _, simPath := range s.SimGraph.AvailablePaths {
			pathProbabilities = append(pathProbabilities, simPath.ExploreProbabilities[epoch-1])
		}
		selectedPathIndex := SampleDiscrete(pathProbabilities)
		currentEpochSelectedPath := s.SimGraph.AvailablePaths[selectedPathIndex]

		// Phase 2.4 pull the arm and observe the gain
		// Phase 2.4.1 get the sessionid
		var currentEpochSessionId = ""
		if previousEpochSelectedPath == nil {
			// 如果是第一次, 需要建立新的 session
			newSessionId := uuid.GetUUID()
			previousEpochSelectedPath = currentEpochSelectedPath
			previousEpochSessionId = newSessionId
			currentEpochSessionId = newSessionId
			// 更新选择的路径
			s.SimGraph.SelectedPaths = append(s.SimGraph.SelectedPaths, currentEpochSelectedPath)
			// 进行新的 session 的建立
			err := s.EstablishSession(currentEpochSessionId, currentEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("establish session failed due to %v", err)
			}
		} else if !entities.SamePath(previousEpochSelectedPath, currentEpochSelectedPath) {
			// 进行旧的 session 的清空
			err := s.DestroySession(previousEpochSessionId, previousEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("destroy session failed due to %v", err)
			}
			// 设置新的 session
			newSessionId := uuid.GetUUID()
			previousEpochSelectedPath = currentEpochSelectedPath
			previousEpochSessionId = newSessionId
			currentEpochSessionId = newSessionId
			// 更新选择的路径
			s.SimGraph.SelectedPaths = append(s.SimGraph.SelectedPaths, currentEpochSelectedPath)
			// 进行新的 session 的建立
			err = s.EstablishSession(currentEpochSessionId, currentEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("establish session failed due to %v", err)
			}
		} else {
			currentEpochSessionId = previousEpochSessionId
			// 更新选择的路径
			s.SimGraph.SelectedPaths = append(s.SimGraph.SelectedPaths, currentEpochSelectedPath)
		}

		// Phase 2.4.3 send a batch of packets
		for j := 0; j < s.SimulatorParams.NumberOfPktsPerBatch; j++ {
			// create simpacket
			simPacket := entities.CreatePacket(currentEpochSelectedPath, currentEpochSessionId)
			// forward packet in all on-path routers
			err := s.ForwardPacket(simPacket, currentEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("forward packet failed due to %v", err)
			}
		}

		// Phase 2.4.4 retrieve information after sending a batch of packets
		counters, err := s.RetrieveCounters(currentEpochSessionId, currentEpochSelectedPath)
		if err != nil {
			return fmt.Errorf("failed to retrieve counters for sessionId %s", currentEpochSessionId)
		}

		// Phase 2.4.5 retrieve directed pv links
		directedPvLinks, directedPvLinksMapping := currentEpochSelectedPath.GetDirectedPvLinks()

		// Phase 2.5 calculate legal and illegal ratio
		for index, counter := range counters {
			// calculate illegal raito
			estimatedIllegalRatio := (float64(counter.IllegalPackets) + float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor) / (float64(counter.IllegalPackets+counter.LegalPackets) + 2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
			directedPvLinks[index].IllegalRatios = append(directedPvLinks[index].IllegalRatios, estimatedIllegalRatio)

			// calcualate packet drop ratio
			if index == 0 {
				// handle impossible case
				currentReceivedPackets := counter.IllegalPackets + counter.LegalPackets
				if currentReceivedPackets > s.SimulatorParams.NumberOfPktsPerBatch {
					fmt.Printf("impossible current received %d, epoch sent: %d\n", currentReceivedPackets, s.SimulatorParams.NumberOfPktsPerBatch) // we need to handle this case by removing the path
					// remove the path that containing this node from the available paths
					// return fmt.Errorf("impossible")
					s.SimGraph.RemovePathContainingTheLink(directedPvLinks[index])
				} else {
					// handle normal case
					estimatedDropRatio := 1 - (float64(counter.IllegalPackets+counter.LegalPackets)+float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)/(float64(s.SimulatorParams.NumberOfPktsPerBatch)+2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
					directedPvLinks[index].DropRatios = append(directedPvLinks[index].DropRatios, estimatedDropRatio)
				}
			} else {
				// handle impossible case
				currentReceivedPackets := counter.IllegalPackets + counter.LegalPackets
				previousReceivedPackets := counters[index-1].LegalPackets + counters[index-1].IllegalPackets
				if currentReceivedPackets > previousReceivedPackets {
					fmt.Printf("impossible current received: %d, previous received: %d\n", currentReceivedPackets, previousReceivedPackets) // we need to handle this case by removing the path
					//return fmt.Errorf("impossible")
					// 注意可能是
					s.SimGraph.RemovePathContainingTheLink(directedPvLinks[index])
				} else {
					// handle possible case
					estimatedDropRatio := 1 - (float64(counter.IllegalPackets+counter.LegalPackets)+float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)/(float64(counters[index-1].LegalPackets)+2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
					directedPvLinks[index].DropRatios = append(directedPvLinks[index].DropRatios, estimatedDropRatio)
				}
			}
		}
		for _, directedPvLink := range s.SimGraph.SimDirectedPvLinks {
			if _, ok := directedPvLinksMapping[directedPvLink.Description]; !ok {
				// not detected the illegal ratio is set to 0
				directedPvLink.IllegalRatios = append(directedPvLink.IllegalRatios, 0)
				directedPvLink.DropRatios = append(directedPvLink.DropRatios, 0)
			}
		}

		// 2.6 unbiased gain estimation
		// traverse each edge
		for _, directedPvLink := range s.SimGraph.SimDirectedPvLinks {
			// 如果是选中的链路
			if _, ok := directedPvLinksMapping[directedPvLink.Description]; ok {
				// calculate the legal ratio
				legalRatio := 1 - directedPvLink.IllegalRatios[epoch]
				// calculate the deliver ratio
				deliverRatio := 1 - directedPvLink.DropRatios[epoch]
				// calculate the estimated gain
				estimatedGain := math.Pow(deliverRatio*legalRatio, s.SimulatorParams.Lambda) // 假设 legal ratio = 90%，那么 estimated gain 在 lambda = 2 的时候是 81%
				// estimatedGain := legalRatio
				// calculate the rectified gain by dividing the estimated gain by the probability of choosing this edge
				rectifiedGain := (estimatedGain + s.SimulatorParams.Bias) / directedPvLink.ExploreProbabilities[epoch-1]
				// update the gain of this edge
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, rectifiedGain)
			} else { // 如果是非选中的链路
				// calculate the estimated gain
				rectifiedGain := s.SimulatorParams.Bias / directedPvLink.ExploreProbabilities[epoch-1] // 如果一个边的被探测概率很低的话, 很可能我们给到的增益会超乎想象的高, 由于我们的丢包率设置的很大, 确实有可能一个边的探索的概率会变的非常的低
				// fmt.Println("rectified gain for unselected edge ", directedPvLink.Description, " is ", rectifiedGain, "bias is ", s.SimulatorParams.Bias, " explore probability is ", directedPvLink.ExploreProbabilities[epoch-1])
				// update the gain of this edge
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, rectifiedGain)
			}
		}

		// 2.7 update the weights if edges and paths
		// 2.7.1 update the edge weights
		for _, directedPvLink := range s.SimGraph.SimDirectedPvLinks {
			currentEpochWeight := directedPvLink.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*directedPvLink.RectifiedGains[epoch])
			directedPvLink.Weights = append(directedPvLink.Weights, currentEpochWeight)
		}
		// 2.7.2 update the path weights
		for _, simPath := range s.SimGraph.AvailablePaths {
			// calculate the gain of path in this epoch
			simPath.Gains = append(simPath.Gains, simPath.CalculateGain(epoch))
			// calcualte the weight of path in this epoch (这里用的是当前计算出来的 gains)
			currentEpochWeight := simPath.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*simPath.Gains[epoch-1]) // 丢包率低的边, 确实 weight 会变得很大, 并且呈现出,越后面 weight 越小的情况
			// update weights
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
			historyForgettingWeight := (1-s.SimulatorParams.BalancingFactor)*simPath.Weights[epoch] + s.SimulatorParams.BalancingFactor*totalPathWeights/float64(len(s.SimGraph.AvailablePaths))
			simPath.Weights[epoch] = historyForgettingWeight
			// fmt.Printf("history forggeting factor: %f percentage: %f\n", s.SimulatorParams.BalancingFactor, s.SimulatorParams.BalancingFactor*totalPathWeights/float64(len(s.SimGraph.AvailablePaths))/totalPathWeights*100)
			// 0.1 / 8 = 0.0125 * 7 = 8.75%
		}
	}

	return nil
}

// SampleDiscrete 根据输入的概率分布进行离散采样，返回采样结果的下标
func SampleDiscrete(probabilities []float64) int {
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

// EstablishSession 进行会话的建立
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

// DestroySession 进行会话的销毁
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

// ForwardPacket 将数据包在路径上进行转发
func (s *Simulator) ForwardPacket(packet *entities.SimPacket, selectedPath *entities.SimPath) error {
	for _, abstractSimNode := range selectedPath.NodeList {
		// 进行 abstractNode 的类型的判断
		if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
			if router, ok := abstractSimNode.ActualNode.(*entities.SimNormalRouter); ok {
				err := router.ProcessPacket(packet)
				if err != nil {
					return fmt.Errorf("process packet failed: %w", err)
				}
				if packet.IsDropped {
					break
				}
			}
		} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				dropPacket, err := pvRouter.ProcessPacket(packet)
				if err != nil {
					return fmt.Errorf("pv router process packet failed due to: %v", err)
				}
				if dropPacket {
					break
				}
			}
		} else {
			return fmt.Errorf("unsupported path node type")
		}
	}
	return nil
}

// RetrieveCounters A --> B --> C 则获取 B 和 C 的 counter 信息
func (s *Simulator) RetrieveCounters(sessionId string, selectedPath *entities.SimPath) ([]*entities.Counter, error) {
	counters := make([]*entities.Counter, 0)
	for _, pvRouter := range selectedPath.PvRouters {
		// retrieve information
		counter, err := pvRouter.RetrieveCounter(sessionId)
		if err != nil {
			// 进行错误处理
			return nil, fmt.Errorf("fail to retrieve information due to: %v", err)
		}
		counters = append(counters, counter)
	}
	return counters, nil
}
