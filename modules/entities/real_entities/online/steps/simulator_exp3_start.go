package steps

import (
	"fmt"
	"math"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
	"zhanghefan123/security_topology/utils/probs"

	"chainmaker.org/chainmaker/common/v2/random/uuid"
)

// StartExp3 进行 simulator 的运行
func (s *Simulator) StartExp3() error {
	// Phase 1: Initialization
	// Phase 1.1 initialize the weights, illegal ratios, and gains for each simDirectedPvLink,
	for _, simDirectedPvLink := range s.SimGraph.SimDirectedAbsLinks {
		simDirectedPvLink.LegalRatios = append(simDirectedPvLink.LegalRatios, 0)
		simDirectedPvLink.Weights = append(simDirectedPvLink.Weights, 1)
		simDirectedPvLink.RectifiedGains = append(simDirectedPvLink.RectifiedGains, 0)
	}
	for _, absNode := range s.SimGraph.SimAbstractNodes {
		if absNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			if pvRouter, ok := absNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				pvRouter.Weights = append(pvRouter.Weights, 1)
				pvRouter.RectifiedGains = append(pvRouter.RectifiedGains, 0)
			}
		}
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
		// Phase 2.1 judge if it is necessary to execute the event
		for index, simEvent := range s.SimEvents {
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
				// delete the event
				s.SimEvents = append(s.SimEvents[:index], s.SimEvents[index+1:]...)
			} else {
				continue
			}
		}

		// Phase 2.2 calculate the path probability
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
		// Phase 2.3 calculate the probability of choosing each edge e
		// Phase 2.3.1 calculate the explore probability for each edge based on the path probability
		for _, simPath := range s.SimGraph.AvailablePaths {
			for _, directedPvLink := range simPath.DirectedAbsLinks {
				edgeExploreProb := simPath.ExploreProbabilities[epoch-1]
				if len(directedPvLink.ExploreProbabilities) == (epoch - 1) {
					directedPvLink.ExploreProbabilities = append(directedPvLink.ExploreProbabilities, edgeExploreProb)
				} else {
					directedPvLink.ExploreProbabilities[epoch-1] += edgeExploreProb
				}
			}
		}
		// Phase 2.3.2 calculate the explore probability for each router based on the path probability
		for _, simPath := range s.SimGraph.AvailablePaths {
			for _, pvRouter := range simPath.PvRouters {
				routerExploreProb := simPath.ExploreProbabilities[epoch-1]
				if len(pvRouter.ExploreProbabilities) == (epoch - 1) {
					pvRouter.ExploreProbabilities = append(pvRouter.ExploreProbabilities, routerExploreProb)
				} else {
					pvRouter.ExploreProbabilities[epoch-1] += routerExploreProb
				}
			}
		}

		// Phase 2.4 select the path according to the probability distribution
		pathProbabilities := make([]float64, 0)
		for _, simPath := range s.SimGraph.AvailablePaths {
			pathProbabilities = append(pathProbabilities, simPath.ExploreProbabilities[epoch-1])
		}
		selectedPathIndex := probs.SampleDiscrete(pathProbabilities)
		currentEpochSelectedPath := s.SimGraph.AvailablePaths[selectedPathIndex]
		var sourceAbs *entities.SimAbstractNode
		var source *entities.SimEndHost
		var ok bool
		if sourceAbs, ok = s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source]; ok {
			source = sourceAbs.ActualNode.(*entities.SimEndHost)
			source.SetCurrentEpochSelectedPath(currentEpochSelectedPath)
		} else {
			return fmt.Errorf("cannot retrieve source")
		}

		// Phase 2.5 pull the arm and observe the gain
		// Phase 2.5.1 establish new session and remove the old session
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

		// Phase 2.5.2 send a batch of packets
		sequence, err := currentEpochSelectedPath.GenerateSampleSequence(s.SimulatorParams.NumberOfPktsPerBatch, s.SimulatorParams.SimulationStrategy)
		if err != nil {
			return fmt.Errorf("generate sample sequence failed due to %w", err)
		}
		var counts []float64
		if s.SimulatorParams.SimulationStrategy == types.SimStrategy_PerBatchBloomFilter {
			counts = make([]float64, len(currentEpochSelectedPath.PvRouters))
		} else {
			counts = make([]float64, len(currentEpochSelectedPath.NodeNameToIndexMapping))
		}
		for j := 0; j < s.SimulatorParams.NumberOfPktsPerBatch; j++ {
			// choose one node among the currentEpochSelectedPath
			idx := sequence[j]
			// create simpacket, 创建包的时候就需要知道采样哪个路由器
			var dataPacket *entities.SimPacket
			if idx != len(currentEpochSelectedPath.PvRouters) {
				dataPacket = entities.CreateSimPacket(types.SimPacketType_DataPacket, currentEpochSessionId, s.SimGraph.SimAbstractNodesMapping[currentEpochSelectedPath.PvRouters[idx].NodeName])
			} else {
				dataPacket = entities.CreateSimPacket(types.SimPacketType_DataPacket, currentEpochSessionId, currentEpochSelectedPath.NodeList[len(currentEpochSelectedPath.NodeList)-1])
			}
			// forward packet in all on-path routers
			err = s.ForwardPacket(dataPacket, currentEpochSelectedPath, s.SimulatorParams.SimulationStrategy)
			if err != nil {
				return fmt.Errorf("forward packet failed due to %v", err)
			}
			counts[idx]++
		}
		finalString := ""
		for index := range len(counts) {
			finalString += fmt.Sprintf("%f->", counts[index])
		}
		fmt.Printf("counts: %s\n", finalString)

		directedAbsLinks, directedAbsLinksMapping := currentEpochSelectedPath.GetDirectedAbsLinks() // 注意这个 abs links 既包括开头的 access link 还包括结尾的 access link
		_, pvRoutersMapping := currentEpochSelectedPath.GetPvRouters()

		if s.SimulatorParams.SimulationStrategy == types.SimStrategy_PerBatchBloomFilter {
			// Phase 2.5.3 retrieve information after sending a batch of packets
			// 1. retrieve counters 首先需要发送 request, 我们需要计算 request 到达后续路由器的 prob, 可以到达 1,2,3,D
			// 2. when prepare to transmit back, 如果从3返回, 那么可能携带了 3的信息, 然后到2丢弃, 然后1携带自己的信息返回
			// 3. if we cannot get the information of counters, it means the link or pv router on link drops the packet. we should lower down the prob of selecting these links
			var recorders []*entities.SimRecorder
			recorders, err = s.RetrieveRecorders(currentEpochSessionId, currentEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("failed to retrieve counters for sessionId %s", currentEpochSessionId)
			}
			// Phase 2.5.4 update current loss
			var packetReceivedByDest int
			packetReceivedByDest, err = recorders[len(recorders)-1].GetValidatedPacketsCount()
			if err != nil {
				return fmt.Errorf("get validated packet count failed due to %w", err)
			}
			s.SimGraph.CurrentLoss += float64(s.SimulatorParams.NumberOfPktsPerBatch) - float64(packetReceivedByDest) // 这个 batch 发送了多少个 - 目的节点最终接收了多少个

			// Phase 2.5.5 update regret
			var bestPathloss float64
			_, bestPathloss, err = s.FindBestSimPathFromGodPerspective(epoch)
			if err != nil {
				return fmt.Errorf("failed to find best sim path: %w", err)
			}

			// Phase 2.5.6 calculate average regret
			regret := s.CalculateRegret(bestPathloss, s.SimGraph.CurrentLoss, epoch)
			s.SimGraph.Regrets = append(s.SimGraph.Regrets, regret)

			// Phase 2.6 calculate legal ratio (recorders 包含最后的 EndHost 的 counter)
			fmt.Printf("path desc: %s\n", currentEpochSelectedPath.Description)
			finalString = ""
			for index, recorder := range recorders {
				if index != (len(recorders) - 1) {
					validatedPacketCnt, _ := recorder.GetValidatedPacketsCount()
					finalString += fmt.Sprintf("%d->", validatedPacketCnt)
				}
			}
			for index, recorder := range recorders {
				var estimatedLegalRatio float64
				var validatedPacketCnt int
				var validatedPacketCntBefore int
				var validatedPacketCntCurrent int
				if index == 0 { // the first is the source
					validatedPacketCnt, err = recorder.GetValidatedPacketsCount()
					if err != nil {
						return fmt.Errorf("get validated packet count failed due to: %w", err)
					}
					estimatedLegalRatio = math.Min((float64(validatedPacketCnt))/counts[0], 1.00) // packet cnt 之所以会不准, 因为使用的是布隆过滤器
					fmt.Printf("estimated legal ratio: %f\n", estimatedLegalRatio)
				} else if index == (len(recorders) - 1) { // the final access link
					validatedPacketCntBefore, err = recorders[index-1].GetValidatedPacketsCount()
					if err != nil {
						return fmt.Errorf("get validated packet count failed due to: %w", err)
					}
					estimatedLegalPacketCntBefore := float64(validatedPacketCntBefore) / counts[index-1] * float64(s.SimulatorParams.NumberOfPktsPerBatch)
					validatedPacketCntCurrent, err = recorders[index].GetValidatedPacketsCount()
					if err != nil {
						return fmt.Errorf("get validated packet count failed due to: %w", err)
					}
					estimatedLegalRatio = math.Min(float64(validatedPacketCntCurrent)/estimatedLegalPacketCntBefore, 1.00)
				} else { // the first is not the source
					validatedPacketCntBefore, err = recorders[index-1].GetValidatedPacketsCount()
					if err != nil {
						return fmt.Errorf("get validated packet count failed due to: %w", err)
					}
					validatedPacketCntCurrent, err = recorders[index].GetValidatedPacketsCount()
					if err != nil {
						return fmt.Errorf("get validated packet count failed due to: %w", err)
					}
					estimatedLegalRatio = (float64(validatedPacketCntCurrent) + float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor) / (float64(validatedPacketCntBefore) + 2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
					fmt.Printf("estimated legal raito: %f\n", estimatedLegalRatio)
				}
				directedAbsLinks[index].LegalRatios = append(directedAbsLinks[index].LegalRatios, estimatedLegalRatio) // if we not receive counter, we cannot calculate the illegal ratio
			}
		} else if s.SimulatorParams.SimulationStrategy == types.SimStrategy_PerPacketAck {
			finalString = ""
			for _, ackCounter := range source.AckCounters {
				finalString += fmt.Sprintf("%d->", ackCounter)
			}
			fmt.Printf("path: %s epoch: %d ack counter: %s\n", currentEpochSelectedPath.Description, epoch, finalString)

			// 2.5.4 update current loss
			var ackReceivedByDest int
			var packetReceivedByDest float64
			ackReceivedByDest = source.AckCounters[len(source.AckCounters)-1]
			packetReceivedByDest = float64(ackReceivedByDest) / counts[len(counts)-1] * float64(s.SimulatorParams.NumberOfPktsPerBatch)
			s.SimGraph.CurrentLoss += float64(s.SimulatorParams.NumberOfPktsPerBatch) - packetReceivedByDest // 这个 batch 发送了多少个 - 目的节点最终接收了多少个

			// 2.5.5  update regret
			var bestPathloss float64
			_, bestPathloss, err = s.FindBestSimPathFromGodPerspective(epoch)
			if err != nil {
				return fmt.Errorf("failed to find best sim path: %w", err)
			}

			// Phase 2.5.6 calculate average regret
			regret := s.CalculateRegret(bestPathloss, s.SimGraph.CurrentLoss, epoch)
			s.SimGraph.Regrets = append(s.SimGraph.Regrets, regret)

			// 2.6 calculate delivery ratio
			for index, ackCount := range source.AckCounters {
				var estimatedLegalRatio float64
				var validatedPacketCntBefore int
				var validatedPacketCntCurrent int
				if index == 0 { // the first is the source
					estimatedLegalRatio = math.Min((float64(ackCount))/counts[0], 1.00) // packet cnt 之所以会不准, 因为使用的是布隆过滤器
					fmt.Printf("estimated legal ratio: %f\n", estimatedLegalRatio)
				} else { // the first is not the source
					validatedPacketCntBefore = source.AckCounters[index-1]
					validatedPacketCntCurrent = source.AckCounters[index]
					estimatedLegalRatio = (float64(validatedPacketCntCurrent) + float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor) / (float64(validatedPacketCntBefore) + 2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
					estimatedLegalRatio = math.Min(estimatedLegalRatio, 1.00)
					fmt.Printf("estimated legal ratio: %f\n", estimatedLegalRatio)
				}
				directedAbsLinks[index].LegalRatios = append(directedAbsLinks[index].LegalRatios, estimatedLegalRatio) // if we not receive counter, we cannot calculate the illegal ratio
			}
		} else {
			return fmt.Errorf("unsupported strategy type")
		}

		// for the undetected link the legal ratio should be set to 1
		var directedAbsLink *entities.SimDirectedAbsLink
		for _, directedAbsLink = range s.SimGraph.SimDirectedAbsLinks {
			if _, ok = directedAbsLinksMapping[directedAbsLink.Description]; !ok {
				directedAbsLink.LegalRatios = append(directedAbsLink.LegalRatios, 0)
			}
		}

		// 2.7 unbiased gain estimation
		// traverse each edge
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			// 如果是选中的链路
			if _, ok = directedAbsLinksMapping[directedPvLink.Description]; ok {
				// calculate the legal ratio
				legalRatio := directedPvLink.LegalRatios[epoch]
				// calculate the estimated gain
				// estimatedGain := math.Pow(legalRatio, s.SimulatorParams.Lambda) // 假设 legal ratio = 90%，那么 estimated gain 在 lambda = 2 的时候是 81%
				estimatedGain := s.GetEstimatedGain(legalRatio)
				fmt.Printf("estimated gain: %f\n", estimatedGain)
				//fmt.Printf("legal ratio: %f, rectified gain: %f\n", legalRatio, estimatedGain)
				// calculate the rectified gain by dividing the estimated gain by the probability of choosing this edge
				// 1 / 0.25
				rectifiedGain := (estimatedGain + s.SimulatorParams.Bias) / (directedPvLink.ExploreProbabilities[epoch-1])
				fmt.Printf("link %s | explore prob: %f\n", directedPvLink.Description, directedPvLink.ExploreProbabilities[epoch-1])
				fmt.Printf("rectified gain: %f\n", rectifiedGain)
				// update the gain of this edge
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, rectifiedGain)
				// set the rectified gains of the router
				if directedPvLink.Source.Type == types.SimNetworkNodeType_PathValidationRouter {
					var sourcePvRouter *entities.SimPathValidationRouter
					if sourcePvRouter, ok = directedPvLink.Source.ActualNode.(*entities.SimPathValidationRouter); ok {
						sourcePvRouter.RectifiedGains = append(sourcePvRouter.RectifiedGains, rectifiedGain)
					} else {
						return fmt.Errorf("cannot get path validation router when calculating rectified gain: %w", err)
					}
				}
			} else { // 如果是非选中的链路
				// calculate the estimated gain
				rectifiedGain := s.SimulatorParams.Bias / directedPvLink.ExploreProbabilities[epoch-1] // 如果一个边的被探测概率很低的话, 很可能我们给到的增益会超乎想象的高, 由于我们的丢包率设置的很大, 确实有可能一个边的探索的概率会变的非常的低
				// update the gain of this edge
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, rectifiedGain)
			}
		}

		// traverse each router to set the rectified gains of the no on-path router
		for _, absSimNode := range s.SimGraph.SimAbstractNodes {
			if absSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				var nodeName string
				nodeName, err = absSimNode.GetSimNodeName()
				if err != nil {
					return fmt.Errorf("get sim node failed due to: %w", err)
				}
				var unselectedPvRouter *entities.SimPathValidationRouter
				if unselectedPvRouter, ok = pvRoutersMapping[nodeName]; !ok {
					// calculate the estimated gain
					rectifiedGain := s.SimulatorParams.Bias / unselectedPvRouter.ExploreProbabilities[epoch-1]
					// update the gain of this edge
					unselectedPvRouter.RectifiedGains = append(unselectedPvRouter.RectifiedGains, rectifiedGain)
				}
			}
		}

		// 2.8 update the weights if edges and paths
		// 2.8.1 update the edge weights
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			currentEpochWeight := directedPvLink.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*directedPvLink.RectifiedGains[epoch])
			directedPvLink.Weights = append(directedPvLink.Weights, currentEpochWeight)
		}
		// 2.8.2 update the router weights
		for _, absNode := range s.SimGraph.SimAbstractNodes {
			if absNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				var pvRouter *entities.SimPathValidationRouter
				if pvRouter, ok = absNode.ActualNode.(*entities.SimPathValidationRouter); ok {
					currentEpochWeight := pvRouter.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*pvRouter.RectifiedGains[epoch])
					pvRouter.Weights = append(pvRouter.Weights, currentEpochWeight)
				} else {
					return fmt.Errorf("cannot get actual node from abs node")
				}
			}
		}

		// 2.8.2 update the path weights
		for _, simPath := range s.SimGraph.AvailablePaths {
			// calculate the gain of path in this epoch
			simPath.Gains = append(simPath.Gains, simPath.CalculateGain(epoch, s.SimulatorParams.GainCalculationStyle))
			// calcualte the weight of path in this epoch (这里用的是当前计算出来的 gains)
			currentEpochWeight := simPath.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*simPath.Gains[epoch-1]) // 丢包率低的边, 确实 weight 会变得很大, 并且呈现出,越后面 weight 越小的情况
			// update weights
			simPath.Weights = append(simPath.Weights, currentEpochWeight)
		}
		// 2.8.3 update totol path weights
		totalPathWeights = 0.0
		for _, simPath := range s.SimGraph.AvailablePaths {
			totalPathWeights = totalPathWeights + simPath.Weights[epoch]
		}
		s.SimGraph.TotalPathWeights = totalPathWeights
		// 2.8.4 forget history smoothing
		for _, simPath := range s.SimGraph.AvailablePaths {
			historyForgettingWeight := (1-s.SimulatorParams.BalancingFactor)*simPath.Weights[epoch] + s.SimulatorParams.BalancingFactor*totalPathWeights/float64(len(s.SimGraph.AvailablePaths))
			simPath.Weights[epoch] = historyForgettingWeight
		}
	}

	return nil
}

// GetEstimatedGain 将 legal ratio 作为输入, 进行修正
func (s *Simulator) GetEstimatedGain(deliveryRatio float64) float64 {
	minDR := s.SimulatorParams.MinimumDeliveryRatio

	// 1. 处理边界：防止 deliveryRatio 为 0 或负数导致 NaN/Inf
	// 同时处理 deliveryRatio 低于最小值的情况，直接设为最小值以保证 Gain 不为负
	effectiveDR := math.Max(deliveryRatio, minDR)

	// 2. 处理 deliveryRatio 超过 1.0 的情况
	effectiveDR = math.Min(effectiveDR, 1.0)

	// 3. 计算增益
	// 当 effectiveDR == minDR 时，结果为 0.0
	// 当 effectiveDR == 1.0 时，结果为 1.0
	return 1.0 - math.Log(effectiveDR)/math.Log(minDR)
}

// EstablishSession 进行会话的建立
func (s *Simulator) EstablishSession(sessionId string, simPath *entities.SimPath) error {
	for index, simAbstractNode := range simPath.NodeList {
		if index != 0 { // the first endhost is the source
			if simAbstractNode.Type == types.SimNetworkNodeType_NormalRouter {
				// do nothing
			} else if simAbstractNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				if pvRouter, ok := simAbstractNode.ActualNode.(*entities.SimPathValidationRouter); ok {
					err := pvRouter.EstablishSession(sessionId, uint(s.SimulatorParams.SizeOfBloomFilter), uint(s.SimulatorParams.HashOfBloomFilter))
					if err != nil {
						return fmt.Errorf("establish session failed due to %v", err)
					}
				}
			} else if simAbstractNode.Type == types.SimNetworkNodeType_EndHost {
				if endHost, ok := simAbstractNode.ActualNode.(*entities.SimEndHost); ok {
					err := endHost.EstablishSession(sessionId)
					if err != nil {
						return fmt.Errorf("establish session failed due to %v", err)
					}
				}
			} else {
				return fmt.Errorf("unsupported path node type")
			}
		}
	}
	return nil
}

// DestroySession 进行会话的销毁
func (s *Simulator) DestroySession(sessionId string, simPath *entities.SimPath) error {
	for index, abstractSimNode := range simPath.NodeList {
		if index != 0 { // the first endhost is the source
			if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
				// do nothing
			} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter {
				if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
					err := pvRouter.DestroySession(sessionId)
					if err != nil {
						return fmt.Errorf("destroy session failed due to %v", err)
					}
				}
			} else if abstractSimNode.Type == types.SimNetworkNodeType_EndHost {
				if endHost, ok := abstractSimNode.ActualNode.(*entities.SimEndHost); ok {
					err := endHost.DestroySession(sessionId)
					if err != nil {
						return fmt.Errorf("destroy session failed due to %v", err)
					}
				}
			} else {
				return fmt.Errorf("unsupported path node type")
			}
		}
	}
	return nil
}

// ForwardPacket 将数据包在路径上进行转发
func (s *Simulator) ForwardPacket(packet *entities.SimPacket, selectedPath *entities.SimPath, simulationStrategy types.SimStrategy) error {
	reversePath := make([]*entities.SimAbstractNode, 0)
	for index, abstractSimNode := range selectedPath.NodeList {
		if index != 0 { // 源节点不进行处理
			// 1. 进行 normal router 的转发
			if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter {
				if router, ok := abstractSimNode.ActualNode.(*entities.SimNormalRouter); ok {
					err := router.ProcessPacket(packet)
					if err != nil {
						return fmt.Errorf("process packet failed: %w", err)
					}
					if packet.IsDropped {
						fmt.Printf("packet being dropped at %s\n", router.NodeName)
						break
					}
				}
			} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter { // 2. 进行 pv router 的转发
				if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
					dropPacket, ackPacket, err := pvRouter.ProcessPacket(packet, simulationStrategy)
					if err != nil {
						return fmt.Errorf("pv router process packet failed due to: %v", err)
					}
					if dropPacket {
						break
					}
					// 如果采样的节点是路径验证节点，则沿着反向路径进行 ack 的发送
					if ackPacket != nil {
						err = s.ForwardAck(reversePath, ackPacket, simulationStrategy)
						if err != nil {
							return fmt.Errorf("forward ack failed due to: %w", err)
						}
					}
				}
			} else if abstractSimNode.Type == types.SimNetworkNodeType_EndHost { // 3. 进行 endhost 的处理
				if endHost, ok := abstractSimNode.ActualNode.(*entities.SimEndHost); ok {
					ackPacket, err := endHost.ProcessPacket(packet, simulationStrategy)
					if err != nil {
						return fmt.Errorf("end host process packet failed due to: %v", err)
					}
					// 如果采样的节点是目的节点，沿着反向路径进行 ack 的发送
					if ackPacket != nil {
						err = s.ForwardAck(reversePath, ackPacket, simulationStrategy)
						if err != nil {
							return fmt.Errorf("forward ack failed due to %v", err)
						}
					}
				}
			} else {
				return fmt.Errorf("unsupported path node type")
			}
		}
		reversePath = append([]*entities.SimAbstractNode{abstractSimNode}, reversePath...)
	}
	return nil
}

// ForwardAck 沿着反向路径进行 ack 的发送
func (s *Simulator) ForwardAck(reversePath []*entities.SimAbstractNode, ackPacket *entities.SimPacket, simulationStrategy types.SimStrategy) error {
	for _, abstractSimNode := range reversePath {
		if abstractSimNode.Type == types.SimNetworkNodeType_NormalRouter { // 1. 进行 normal router 的处理
			if router, ok := abstractSimNode.ActualNode.(*entities.SimNormalRouter); ok {
				err := router.ProcessPacket(ackPacket)
				if err != nil {
					return fmt.Errorf("process packet failed: %w", err)
				}
				if ackPacket.IsDropped {
					break
				}
			}
		} else if abstractSimNode.Type == types.SimNetworkNodeType_PathValidationRouter { // 2. 进行 path validation router 的处理
			if pvRouter, ok := abstractSimNode.ActualNode.(*entities.SimPathValidationRouter); ok {
				dropPacket, _, err := pvRouter.ProcessPacket(ackPacket, simulationStrategy)
				if err != nil {
					return fmt.Errorf("pv router process packet failed due to: %v", err)
				}
				if dropPacket {
					break
				}
			}
		} else if abstractSimNode.Type == types.SimNetworkNodeType_EndHost { // 3. 进行 endhost 的处理
			if endHost, ok := abstractSimNode.ActualNode.(*entities.SimEndHost); ok {
				_, err := endHost.ProcessPacket(ackPacket, simulationStrategy)
				if err != nil {
					return fmt.Errorf("end host process packet failed due to: %v", err)
				}
			}
		}
	}
	return nil
}

// RetrieveRecorders A --> B --> C 则获取 B 的 bloom filter 和 C 的 counter 信息
func (s *Simulator) RetrieveRecorders(sessionId string, selectedPath *entities.SimPath) ([]*entities.SimRecorder, error) {
	simRecorders := make([]*entities.SimRecorder, 0)
	// 1. retrieve from path validation routers
	for _, pvRouter := range selectedPath.PvRouters {
		// retrieve information
		simRecorder, err := pvRouter.RetrieveRecorder(sessionId)
		if err != nil {
			// 进行错误处理
			return nil, fmt.Errorf("fail to retrieve information due to: %v", err)
		}
		simRecorders = append(simRecorders, simRecorder)
	}
	// 2. retrieve from destination
	if endHost, ok := s.SimGraph.DestinationNode.ActualNode.(*entities.SimEndHost); ok {
		recorder, err := endHost.RetrieveRecorder(sessionId)
		if err != nil {
			return nil, fmt.Errorf("retrieve recorder from destination failed due to: %w", err)
		}
		simRecorders = append(simRecorders, recorder)
	}
	return simRecorders, nil
}

// CalculateRegret 进行悔值的计算
func (s *Simulator) CalculateRegret(minimumLoss float64, currentLoss float64, currentEpoch int) float64 {
	// 应该不用 / NumberOfPktsPerBatch
	//return (currentLoss - minimumLoss) / float64(currentEpoch) / float64(s.SimulatorParams.NumberOfPktsPerBatch)
	// currentloss 可能还会比 minimumLoss 还要小, 因为 currentLoss 是通过实际转发得出的, 而 minimumLoss 则是直接取最小丢包和最大丢包的均值计算出来的
	return (currentLoss - minimumLoss) / float64(currentEpoch)
}

// FindBestSimPathFromGodPerspective 寻找到当前 epoch 最好的路径, return 最好路径以及最好路径的 losses
func (s *Simulator) FindBestSimPathFromGodPerspective(currentEpoch int) (*entities.SimPath, float64, error) {
	// losses
	minimumLoss := float64(math.MaxInt)
	var minimumLossPath *entities.SimPath = nil

	// 找到一个周期内的
	for _, simPath := range s.SimGraph.AvailablePaths {
		simPathLoss := 0.0
		for _ = range currentEpoch {
			deliveryRatio := 1.0
			for _, pvLink := range simPath.DirectedAbsLinks {
				if simNormalRouter, ok := pvLink.Intermediate.ActualNode.(*entities.SimNormalRouter); ok {
					deliveryRatio = deliveryRatio * (1 - (simNormalRouter.StartCorruptRatio+simNormalRouter.EndCorruptRatio)/2)
				} else {
					return nil, 0, fmt.Errorf("cannot get intermediate node")
				}
			}
			simPathLoss += float64(s.SimulatorParams.NumberOfPktsPerBatch) * (1.0 - deliveryRatio)
		}
		if simPathLoss < minimumLoss {
			minimumLossPath = simPath
			minimumLoss = simPathLoss
		}
	}

	return minimumLossPath, minimumLoss, nil
}
