package steps

import (
	"fmt"
	"math"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
	"zhanghefan123/security_topology/utils/probs"

	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"gonum.org/v1/gonum/graph/topo"
)

func (s *Simulator) StartOsmd() error {
	// 1. 让所有的路径的概率相同 (并计算出每条边的概率)
	s.InitAllEdgesProbabilities()

	// 2. 进入循环
	var previousEpochSelectedPath *entities.SimPath = nil
	var previousEpochSessionId = ""
	for epoch := 1; epoch <= s.SimulatorParams.NumberOfEpochs; epoch++ {
		// 2.0 execute events
		for index, simEvent := range s.SimEvents {
			if epoch == simEvent.StartEpoch {
				for _, updateRouter := range simEvent.UpdateRouters {
					if selectedAbstractNormalRouter, ok := s.SimGraph.SimAbstractNodesMapping[updateRouter.NormalRouterName]; ok {
						var selectedNormalRouter *entities.SimNormalRouter
						if selectedNormalRouter, ok = selectedAbstractNormalRouter.ActualNode.(*entities.SimNormalRouter); ok {
							err := selectedNormalRouter.Reset(updateRouter.StartCorruptRatio, updateRouter.EndCorruptRatio,
								updateRouter.StartCorruptSpecialRatio, updateRouter.EndCorruptSpecialRatio)
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

		// 2.1 设置每条边的当前的概率
		s.SetCurrentEdgesProbability(epoch)

		// 2.2 根据边的概率进行流分解
		pathMapping, err := s.DecomposeToGetPathProbabilities() // 第一轮运行的结果是正确的
		if err != nil {
			return fmt.Errorf("DecomposeToGetPathProbabilities() failed: %v", err)
		}
		fmt.Println("--------------------------------------")
		for pathDesc, probability := range pathMapping {
			fmt.Println(pathDesc)
			fmt.Printf("path id: %d, probability: %f\n", s.SimGraph.AvailablePathMapping[pathDesc].PathId, probability)
		}
		fmt.Println("--------------------------------------")

		// 2.2 决策当前路径
		currentEpochSelectedPath := s.DetermineCurrentEpochSelectedPath(pathMapping)

		// 2.3 设置当前所选路径
		var sourceAbs *entities.SimAbstractNode
		var source *entities.SimEndHost
		var ok bool
		if sourceAbs, ok = s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source]; ok {
			source = sourceAbs.ActualNode.(*entities.SimEndHost)
			source.SetCurrentEpochSelectedPath(currentEpochSelectedPath)
		} else {
			return fmt.Errorf("cannot retrieve source")
		}

		// 2.4 拉杆操作
		// 2.4.1 建立新 session, 移除旧 session
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
			err = s.EstablishSession(currentEpochSessionId, currentEpochSelectedPath)
			if err != nil {
				return fmt.Errorf("establish session failed due to %v", err)
			}
		} else if !entities.SamePath(previousEpochSelectedPath, currentEpochSelectedPath) {
			// 进行旧的 session 的清空
			err = s.DestroySession(previousEpochSessionId, previousEpochSelectedPath)
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

		// 2.4.2 发送一批数据包
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
			// 选择抽样的路由器
			idx := sequence[j]
			// 创建包的时候就需要知道采样哪个路由器
			var dataPacket *entities.SimPacket
			if idx != len(currentEpochSelectedPath.PvRouters) {
				dataPacket = entities.CreateSimPacket(types.SimPacketType_DataPacket, currentEpochSessionId, s.SimGraph.SimAbstractNodesMapping[currentEpochSelectedPath.PvRouters[idx].NodeName])
			} else {
				dataPacket = entities.CreateSimPacket(types.SimPacketType_DataPacket, currentEpochSessionId, currentEpochSelectedPath.NodeList[len(currentEpochSelectedPath.NodeList)-1])
			}
			// 进行数据包的转发
			err = s.ForwardPacket(dataPacket, currentEpochSelectedPath, s.SimulatorParams.SimulationStrategy)
			if err != nil {
				return fmt.Errorf("forward packet failed due to %v", err)
			}
			counts[idx]++
		}

		// 2.4.3 更新当前的损失 - 用来计算悔值
		var ackReceivedByDest int
		var packetReceivedByDest float64
		ackReceivedByDest = source.AckCounters[len(source.AckCounters)-1]
		packetReceivedByDest = float64(ackReceivedByDest) / counts[len(counts)-1] * float64(s.SimulatorParams.NumberOfPktsPerBatch)
		s.SimGraph.CurrentLoss += float64(s.SimulatorParams.NumberOfPktsPerBatch) - packetReceivedByDest // 这个 batch 发送了多少个 - 目的节点最终接收了多少个

		// 2.4.4  计算悔值
		var bestPathloss float64
		_, bestPathloss, err = s.FindBestSimPathFromGodPerspective(epoch)
		if err != nil {
			return fmt.Errorf("failed to find best sim path: %w", err)
		}

		// 2.4.5 计算平均悔值
		regret := s.CalculateRegret(bestPathloss, s.SimGraph.CurrentLoss, epoch)
		s.SimGraph.Regrets = append(s.SimGraph.Regrets, regret)

		// 2.4.6 计算传输失败率
		directedAbsLinks, directedAbsLinksMapping := currentEpochSelectedPath.GetDirectedAbsLinks() // 注意这个 abs links 既包括开头的 access link 还包括结尾的 access link

		for index, ackCount := range source.AckCounters {
			var estimatedLegalRatio float64
			var validatedPacketCntBefore int
			var validatedPacketCntCurrent int
			if index == 0 { // the first is the source
				estimatedLegalRatio = math.Min((float64(ackCount))/counts[0], 1.00) // packet cnt 之所以会不准, 因为使用的是布隆过滤器

			} else { // the first is not the source
				validatedPacketCntBefore = source.AckCounters[index-1]
				validatedPacketCntCurrent = source.AckCounters[index]
				estimatedLegalRatio = (float64(validatedPacketCntCurrent) + float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor) / (float64(validatedPacketCntBefore) + 2*float64(s.SimulatorParams.NumberOfPktsPerBatch)*s.SimulatorParams.LaplaceSmoothingFactor)
				estimatedLegalRatio = math.Min(estimatedLegalRatio, 1.00)
			}
			directedAbsLinks[index].IllegalRatios = append(directedAbsLinks[index].IllegalRatios, 1-estimatedLegalRatio)
		}

		// 未进行探测的链路的非法率为 0
		var directedAbsLink *entities.SimDirectedAbsLink
		for _, directedAbsLink = range s.SimGraph.SimDirectedAbsLinks {
			if _, ok = directedAbsLinksMapping[directedAbsLink.Description]; !ok {
				directedAbsLink.IllegalRatios = append(directedAbsLink.IllegalRatios, 0)
			}
		}

		// 2.4.7 计算无偏估计
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			// if the link is picked up
			if _, ok = directedAbsLinksMapping[directedPvLink.Description]; ok {
				illegalRatio := directedPvLink.IllegalRatios[epoch-1]
				estimatedLoss := 1 - s.GetEstimatedGain(1-illegalRatio)
				rectifiedLoss := estimatedLoss / directedPvLink.ExploreProbabilities[epoch-1]
				directedPvLink.RectifiedLosses = append(directedPvLink.RectifiedLosses, rectifiedLoss)
			} else {
				directedPvLink.RectifiedLosses = append(directedPvLink.RectifiedLosses, 0)
			}
		}

		// 2.4.8 更新各个节点的权重
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			// rectified loss 最大为   probability * exp(-1 * learning rate)
			currentEpochWeight := directedPvLink.ExploreProbabilities[epoch-1] * math.Exp(-s.SimulatorParams.LearningRate*directedPvLink.RectifiedLosses[epoch-1])
			directedPvLink.Weights = append(directedPvLink.Weights, currentEpochWeight)
		}

		// 2.4.9 将权重重新进行投影
		err = s.ProjectionBackToLegalPlane(
			s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source],
			s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Destination],
			epoch)
		if err != nil {
			return fmt.Errorf("project back to legal plane failed due to: %w", err)
		}
	}

	return nil
}

func (s *Simulator) InitAllEdgesProbabilities() {
	edgeTotalProbability := 0.0
	numberOfPaths := len(s.SimGraph.AvailablePaths)
	singlePathProbability := 1.0 / float64(numberOfPaths)
	for _, simPath := range s.SimGraph.AvailablePaths {
		for _, simAbsLink := range simPath.DirectedAbsLinks {
			if len(simAbsLink.ExploreProbabilities) == 0 {
				simAbsLink.ExploreProbabilities = append(simAbsLink.ExploreProbabilities, singlePathProbability)
			} else {
				simAbsLink.ExploreProbabilities[0] += singlePathProbability
			}
			edgeTotalProbability += singlePathProbability
		}
	}
}

func (s *Simulator) DetermineCurrentEpochSelectedPath(pathMapping map[string]float64) *entities.SimPath {
	pathProbabilities := make([]float64, 0)
	for _, simPath := range s.SimGraph.AvailablePaths {
		if probability, ok := pathMapping[simPath.Description]; ok {
			pathProbabilities = append(pathProbabilities, probability)
		} else {
			pathProbabilities = append(pathProbabilities, 0)
		}
	}
	selectedPathIndex := probs.SampleDiscrete(pathProbabilities)
	currentEpochSelectedPath := s.SimGraph.AvailablePaths[selectedPathIndex]
	return currentEpochSelectedPath
}

func (s *Simulator) SetCurrentEdgesProbability(epoch int) {
	for _, simAbsLink := range s.SimGraph.SimDirectedAbsLinks {
		simAbsLink.CurrentEdgeProbability = simAbsLink.ExploreProbabilities[epoch-1]
	}
}

func (s *Simulator) DecomposeToGetPathProbabilities() (map[string]float64, error) {
	epsilon := 1e-9
	pathMapping := make(map[string]float64)
	// -------------------------------------- 寻找很多路 --------------------------------------\
OutLoop:
	for {
		startPoint := s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source] // 拿到起始节点
		currentPathDesc := ""                                                                            // 当前路径的描述
		currentPath := make([]*entities.SimDirectedAbsLink, 0)                                           // 当前的路径
		pathProbabilitiesList := make([]float64, 0)                                                      // 当前路径的各条边的概率
		minimumEdgeProbability := math.MaxFloat64
		// -------------------------------------- 寻找一条路 --------------------------------------
	FindPathLoop:
		for {
			// 获取节点所有的出边
			allOutputEdges, err := s.SimGraph.FindAllOutputEdges(startPoint)
			if err != nil {
				return nil, fmt.Errorf("cannot find all output edges due to %w", err)
			}
			// 进行所有的抽象边的遍历
			moved := false
			for _, outputEdge := range allOutputEdges {
				if outputEdge.CurrentEdgeProbability > epsilon {
					var sourceName, normalRouterName string
					sourceName, _ = outputEdge.Source.GetSimNodeName()
					normalRouterName, _ = outputEdge.Intermediate.GetSimNodeName()
					currentPathDesc += fmt.Sprintf("%s,%s,", sourceName, normalRouterName)
					pathProbabilitiesList = append(pathProbabilitiesList, outputEdge.CurrentEdgeProbability)
					currentPath = append(currentPath, outputEdge)
					startPoint = outputEdge.Target
					// 更新最小的概率
					if outputEdge.CurrentEdgeProbability < minimumEdgeProbability {
						minimumEdgeProbability = outputEdge.CurrentEdgeProbability
					}
					moved = true
					break
				}
			}

			// 如果没有找到任何可用出边（遇到死胡同或残渣），强制退出寻找，防止死循环
			if !moved {
				break OutLoop
			}

			var startNodeName string
			startNodeName, err = startPoint.GetSimNodeName()
			if err != nil {
				return nil, fmt.Errorf("get start node name failed due to: %w", err)
			}
			if startNodeName == s.SimGraph.GraphParams.SourceDestParams.Destination {
				currentPathDesc += startNodeName
				pathMapping[currentPathDesc] += minimumEdgeProbability
				// 进行所选的这条路的遍历, 将所有边的概率减去这条路径上的边的概率的最小值
				for _, inPathLink := range currentPath {
					inPathLink.CurrentEdgeProbability -= minimumEdgeProbability
				}
				break FindPathLoop
			}
			// -------------------------------------- 寻找一条边 (概率大于0的边) --------------------------------------
		}
		// -------------------------------------- 寻找一条路 --------------------------------------

		// 判断当前是否还有链路的概率大于 0
		allZero := true
		for _, simAbsLink := range s.SimGraph.SimDirectedAbsLinks {
			if simAbsLink.CurrentEdgeProbability > epsilon {
				allZero = false
				break
			}
		}
		if allZero {
			break
		}
	}
	// -------------------------------------- 寻找很多路 --------------------------------------
	return pathMapping, nil
}

func (s *Simulator) ProjectionBackToLegalPlane(source, destination *entities.SimAbstractNode, epoch int) error {
	// 0. 获取目的节点的名称
	destinationNodeName, err := destination.GetSimNodeName()
	if err != nil {
		return fmt.Errorf("error to get destination node name due to: %w", err)
	}

	// 1. 首先进行拓扑排序
	sortedNodes, err := topo.Sort(s.SimGraph.RealGraph)
	if err != nil {
		return fmt.Errorf("cannot sort nodes: %w", err)
	}

	// 2. 将 normal Nodes 进行删除
	pvRouterOrEndHosts := make([]*entities.SimAbstractNode, 0)
	for index := 0; index < len(sortedNodes); index++ {
		currentNode := sortedNodes[index]
		if absNode, ok := currentNode.(*entities.SimAbstractNode); ok {
			if absNode.Type != types.SimNetworkNodeType_NormalRouter {
				pvRouterOrEndHosts = append(pvRouterOrEndHosts, absNode)
			}
		} else {
			return fmt.Errorf("cannot turn node into abstract node")
		}
	}

	// 2. 进行逆拓扑排序
	for index := len(pvRouterOrEndHosts) - 1; index >= 0; index-- {
		currentNode := pvRouterOrEndHosts[index]
		// 2.1 进行设置
		if currentNode.Type == types.SimNetworkNodeType_EndHost {
			var currentNodeName string
			currentNodeName, err = currentNode.GetSimNodeName()
			if err != nil {
				return fmt.Errorf("get current node name failed due to: %w", err)
			}
			if destinationNodeName == currentNodeName {
				currentNode.Potential = 1.0
			} else {
				// 寻找所有的出边, 把他们的 w * potential 都拿过来
				var allOutputEdges []*entities.SimDirectedAbsLink
				allOutputEdges, err = s.SimGraph.FindAllOutputEdges(currentNode)
				if err != nil {
					return fmt.Errorf("cannot find all output edges: %w", err)
				}
				// 进行 potential 的更新
				nodePotential := 0.0
				for _, outputEdge := range allOutputEdges {
					nodePotential += outputEdge.Weights[epoch-1] * outputEdge.Target.Potential
				}
				// 进行节点的 potential 的更新
				currentNode.Potential = nodePotential
			}
		} else if currentNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			// 寻找所有的出边, 把他们的 w * potential 都拿过来
			var allOutputEdges []*entities.SimDirectedAbsLink
			allOutputEdges, err = s.SimGraph.FindAllOutputEdges(currentNode)
			if err != nil {
				return fmt.Errorf("cannot find all output edges: %w", err)
			}
			// 进行 potential 的更新
			nodePotential := 0.0
			for _, outputEdge := range allOutputEdges {
				nodePotential += outputEdge.Weights[epoch-1] * outputEdge.Target.Potential
			}
			// 进行节点的 potential 的更新
			currentNode.Potential = nodePotential
		} else {
			return fmt.Errorf("cannot turn node into abstract node")
		}
	}

	// 3. 根据之前计算的潜力计算每条边应该分配的流量
	// 3.1 首先将源的 flow 设置为 1
	source.Flow = 1
	// 3.2 按照正向的拓扑排序进行节点的遍历
	for index := 0; index < len(pvRouterOrEndHosts); index++ {
		currentNode := pvRouterOrEndHosts[index]
		var allOutputEdges []*entities.SimDirectedAbsLink
		allOutputEdges, err = s.SimGraph.FindAllOutputEdges(currentNode)
		for _, outputEdge := range allOutputEdges {
			// 避免分母的除0错误
			if outputEdge.Source.Potential > 0 {
				// 计算当前的边的概率
				currentEdgeProbability := outputEdge.Source.Flow * outputEdge.Weights[epoch-1] * outputEdge.Target.Potential / outputEdge.Source.Potential
				outputEdge.ExploreProbabilities = append(outputEdge.ExploreProbabilities, currentEdgeProbability)
				// 更新流的强度
				outputEdge.Target.Flow += currentEdgeProbability
			} else {
				outputEdge.ExploreProbabilities = append(outputEdge.ExploreProbabilities, 0)
			}
		}
	}

	// 4. 进行所有的 potential 和 flow 的清空
	for _, abstractNode := range s.SimGraph.SimAbstractNodes {
		abstractNode.Potential = 0
		abstractNode.Flow = 0
	}
	return nil
}
