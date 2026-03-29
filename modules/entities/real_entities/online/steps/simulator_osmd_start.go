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
	// 1. initialize all the edges probability
	s.InitAllEdgesProbabilities()

	// 2. enter iteration
	var previousEpochSelectedPath *entities.SimPath = nil
	var previousEpochSessionId = ""
	for epoch := 1; epoch <= s.SimulatorParams.NumberOfEpochs; epoch++ {
		s.SetCurrentEdgesProbability(epoch)

		// 2.1 decompose
		pathMapping, err := s.DecomposeToGetPathProbabilities()
		if err != nil {
			return fmt.Errorf("DecomposeToGetPathProbabilities() failed: %v", err)
		}

		// 2.2 sample discrete
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

		// 2.3 set current epoch selected path
		var sourceAbs *entities.SimAbstractNode
		var source *entities.SimEndHost
		var ok bool
		if sourceAbs, ok = s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source]; ok {
			source = sourceAbs.ActualNode.(*entities.SimEndHost)
			source.SetCurrentEpochSelectedPath(currentEpochSelectedPath)
		} else {
			return fmt.Errorf("cannot retrieve source")
		}

		// 2.4 pull the arm and get responses (send a batch of packets on the current epoch selected path)
		// 2.4.1 establish new session and remove the old session
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
		// 2.4.2 send a batch of packets
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

		finalString = ""
		for _, ackCounter := range source.AckCounters {
			finalString += fmt.Sprintf("%d->", ackCounter)
		}
		fmt.Printf("path: %s epoch: %d ack counter: %s\n", currentEpochSelectedPath.Description, epoch, finalString)

		// 2.4.3 update current loss
		var ackReceivedByDest int
		var packetReceivedByDest float64
		ackReceivedByDest = source.AckCounters[len(source.AckCounters)-1]
		packetReceivedByDest = float64(ackReceivedByDest) / counts[len(counts)-1] * float64(s.SimulatorParams.NumberOfPktsPerBatch)
		s.SimGraph.CurrentLoss += float64(s.SimulatorParams.NumberOfPktsPerBatch) - packetReceivedByDest // 这个 batch 发送了多少个 - 目的节点最终接收了多少个

		// 2.4.4  update regret
		var bestPathloss float64
		_, bestPathloss, err = s.FindBestSimPathFromGodPerspective(epoch)
		if err != nil {
			return fmt.Errorf("failed to find best sim path: %w", err)
		}

		// 2.4.5 calculate average regret
		regret := s.CalculateRegret(bestPathloss, s.SimGraph.CurrentLoss, epoch)
		s.SimGraph.Regrets = append(s.SimGraph.Regrets, regret)

		// 2.4.6 calculate delivery ratio
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

		// for the undetected link the legal ratio should be set to 1
		var directedAbsLink *entities.SimDirectedAbsLink
		for _, directedAbsLink = range s.SimGraph.SimDirectedAbsLinks {
			if _, ok = directedAbsLinksMapping[directedAbsLink.Description]; !ok {
				directedAbsLink.LegalRatios = append(directedAbsLink.LegalRatios, 0)
			}
		}

		// 2.4.7 calculate unbiased gain
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			// if the link is picked up
			if _, ok = directedAbsLinksMapping[directedPvLink.Description]; ok {
				legalRatio := directedPvLink.LegalRatios[epoch]
				rectifiedGain := legalRatio / directedPvLink.ExploreProbabilities[epoch-1]
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, rectifiedGain)
			} else {
				directedPvLink.RectifiedGains = append(directedPvLink.RectifiedGains, 0)
			}
		}

		// 2.4.8 update weight
		for _, directedPvLink := range s.SimGraph.SimDirectedAbsLinks {
			currentEpochWeight := directedPvLink.Weights[epoch-1] * math.Exp(s.SimulatorParams.LearningRate*directedPvLink.RectifiedGains[epoch])
			directedPvLink.Weights = append(directedPvLink.Weights, currentEpochWeight)
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

func (s *Simulator) SetCurrentEdgesProbability(epoch int) {
	for _, simAbsLink := range s.SimGraph.SimDirectedAbsLinks {
		simAbsLink.CurrentEdgeProbability = simAbsLink.ExploreProbabilities[epoch-1]
	}
}

func (s *Simulator) DecomposeToGetPathProbabilities() (map[string]float64, error) {
	pathMapping := make(map[string]float64)
	// -------------------------------------- 寻找很多路 --------------------------------------
	for {
		startPoint := s.SimGraph.SimAbstractNodesMapping[s.SimGraph.GraphParams.SourceDestParams.Source] // 拿到起始节点
		currentPathDesc := ""                                                                            // 当前路径的描述
		currentPath := make([]*entities.SimDirectedAbsLink, 0)                                           // 当前的路径
		pathProbabilitiesList := make([]float64, 0)                                                      // 当前路径的各条边的概率
		minimumEdgeProbability := math.MaxFloat64
		// -------------------------------------- 寻找一条路 --------------------------------------
	FindPathLoop:
		for {
			// 获取起始节点的名称
			startNodeName, _ := startPoint.GetSimNodeName()
			// 寻找所有的出节点
			pvRouterOrEndHostOutNodes := s.SimGraph.RealGraph.From(startPoint.ID())
			// -------------------------------------- 寻找一条边 (概率大于0的边) --------------------------------------
			// 遍历所有的出节点找到一个概率大于 0 的边
			for pvRouterOrEndHostOutNodes.Next() {
				// 获取出节点
				normalGraphNode := pvRouterOrEndHostOutNodes.Node()
				// 获取当前节点和出节点对应的边
				if absToNormalNode, ok := normalGraphNode.(*entities.SimAbstractNode); ok {
					// 获取普通节点名称
					normalNodeName, _ := absToNormalNode.GetSimNodeName()
					normalOutNodes := s.SimGraph.RealGraph.From(normalGraphNode.ID())
					normalOutNodes.Next()
					pvRouterOrEndHostGraphNode := normalOutNodes.Node()
					var absPvRouterOrEndHost *entities.SimAbstractNode
					if absPvRouterOrEndHost, ok = pvRouterOrEndHostGraphNode.(*entities.SimAbstractNode); ok {
						pvRouterOrEndHostName, _ := absPvRouterOrEndHost.GetSimNodeName()
						// 获取当前的 link
						absLink := s.SimGraph.SimDirectedAbsLinksMapping[fmt.Sprintf("%s->%s->%s", startNodeName, normalNodeName, pvRouterOrEndHostName)]
						// 判断链路的概率是否大于 0
						if absLink.CurrentEdgeProbability > 0 {
							currentPathDesc += fmt.Sprintf("%s->%s->%s", startNodeName, normalNodeName, pvRouterOrEndHostName)
							pathProbabilitiesList = append(pathProbabilitiesList, absLink.CurrentEdgeProbability)
							currentPath = append(currentPath, absLink)
							if absLink.CurrentEdgeProbability < minimumEdgeProbability {
								minimumEdgeProbability = absLink.CurrentEdgeProbability
							}
							startPoint = absPvRouterOrEndHost
							if startNodeName == s.SimGraph.GraphParams.SourceDestParams.Destination {
								pathMapping[currentPathDesc[:len(currentPathDesc)-2]] = minimumEdgeProbability
								// 进行所选的这条路的遍历, 将所有边的概率减去这条路径上的边的概率的最小值
								for _, inPathLink := range currentPath {
									inPathLink.CurrentEdgeProbability -= minimumEdgeProbability
								}
								break FindPathLoop
							}
						}
					} else {
						return nil, fmt.Errorf("pvRouterOrEndHostGraphNode cannot transformed into SimAbstractNode")
					}
				} else {
					return nil, fmt.Errorf("cannot find the node")
				}
			}
			// -------------------------------------- 寻找一条边 (概率大于0的边) --------------------------------------
		}
		// -------------------------------------- 寻找一条路 --------------------------------------

		// 判断当前是否还有链路的概率大于 0
		allZero := true
		for _, simAbsLink := range s.SimGraph.SimDirectedAbsLinks {
			if simAbsLink.CurrentEdgeProbability > 0 {
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

func (s *Simulator) ProjectionBackToLegalPlane(source, destination *entities.SimEndHost, epoch int) error {
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
			currentEndHost := currentNode.ActualNode.(*entities.SimEndHost)
			if destination.NodeName == currentEndHost.NodeName {
				currentEndHost.Potential = 1.0
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
					var targetPotential float64
					targetPotential, err = outputEdge.Target.GetPotentialFromAbstract()
					if err != nil {
						return fmt.Errorf("cannot get potential from abstract: %w", err)
					}
					nodePotential += outputEdge.CurrentEdgeProbability * targetPotential
				}
				// 进行节点的 potential 的更新
				currentEndHost.Potential = nodePotential
			}
		} else if currentNode.Type == types.SimNetworkNodeType_PathValidationRouter {
			currentPathValidationRouter := currentNode.ActualNode.(*entities.SimPathValidationRouter)
			// 寻找所有的出边, 把他们的 w * potential 都拿过来
			var allOutputEdges []*entities.SimDirectedAbsLink
			allOutputEdges, err = s.SimGraph.FindAllOutputEdges(currentNode)
			if err != nil {
				return fmt.Errorf("cannot find all output edges: %w", err)
			}
			// 进行 potential 的更新
			nodePotential := 0.0
			for _, outputEdge := range allOutputEdges {
				var targetPotential float64
				targetPotential, err = outputEdge.Target.GetPotentialFromAbstract()
				if err != nil {
					return fmt.Errorf("cannot get potential from abstract: %w", err)
				}
				nodePotential += outputEdge.CurrentEdgeProbability * targetPotential
			}
			// 进行节点的 potential 的更新
			currentPathValidationRouter.Potential = nodePotential
		} else {
			return fmt.Errorf("cannot turn node into abstract node")
		}
	}
	// 3. 按照正向的顺序进行排序
	for index := 0; index < len(pvRouterOrEndHosts); index++ {
		currentNode := pvRouterOrEndHosts[index]
		var allOutputEdges []*entities.SimDirectedAbsLink
		allOutputEdges, err = s.SimGraph.FindAllOutputEdges(currentNode)
		for _, outputEdge := range allOutputEdges {
			var currentPotential float64
			currentPotential, err = currentNode.GetPotentialFromAbstract()
			if err != nil {
				return fmt.Errorf("get current potential failed due to: %w", err)
			}
			var destPotential float64
			destPotential, err = outputEdge.Target.GetPotentialFromAbstract()
			if err != nil {
				return fmt.Errorf("get destination potential failed due to: %w", err)
			}
			outputEdge.ExploreProbabilities = append(outputEdge.ExploreProbabilities, outputEdge.Weights[epoch]*destPotential/currentPotential)
		}
	}
	return nil
}
