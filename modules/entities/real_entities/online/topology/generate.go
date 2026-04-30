package topology

import (
	"encoding/json"
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/params"
	"zhanghefan123/security_topology/utils/file"
)

type Description struct {
	NumberOfHops     int                      `json:"number_of_hops"`
	SourceDestParams params.SourceDestParams  `json:"source_dest_params"`
	Nodes            []params.SimNodeParam    `json:"nodes"`
	PvLinks          []params.SimAbsLinkParam `json:"pv_links"`
	Links            []params.SimLinkParam    `json:"links"`
}

// GenerateTopologyDescription 进行拓扑的生成
func GenerateTopologyDescription(numberOfHops, lowRatio, highRatio int) *Description {
	finalResult := &Description{
		NumberOfHops: numberOfHops,
	}
	FillSourceDestParams(finalResult, numberOfHops)
	indexToNodeParamMapping := FillNodes(finalResult, numberOfHops, lowRatio, highRatio)
	FillPvLinks(finalResult, numberOfHops, indexToNodeParamMapping)
	FillRealLinks(finalResult, numberOfHops, indexToNodeParamMapping)
	return finalResult
}

// MarshalTopologyDescription 将拓扑描述放到文件之中
func MarshalTopologyDescription(topologyDescription *Description) {
	bytes, err := json.MarshalIndent(topologyDescription, "", "  ")
	if err != nil {
		return
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("./%s_%d_hops.json", "topology_with", topologyDescription.NumberOfHops), string(bytes))
	if err != nil {
		return
	}
}

// FillSourceDestParams 填充 topologyDescription 之中的 source dest params
func FillSourceDestParams(topologyDesc *Description, numberOfHops int) {
	topologyDesc.SourceDestParams = params.SourceDestParams{
		Source:      "EndHost-1",
		Destination: fmt.Sprintf("EndHost-%d", 1+numberOfHops*3),
	}
}

// FillNodes 填充节点
func FillNodes(topologyDesc *Description, numberOfHops, lowRatio, highRatio int) map[int]params.SimNodeParam {
	finalMapping := map[int]params.SimNodeParam{}
	numberOfNodes := 1 + numberOfHops*3
	topologyDesc.Nodes = make([]params.SimNodeParam, numberOfNodes)
	topologyDesc.Nodes[0] = params.SimNodeParam{
		Index: 1,
		Type:  "EndHost",
	}
	finalMapping[1] = topologyDesc.Nodes[0]
	currentHop := 0
	index := 1
	for {
		topologyDesc.Nodes[index] = params.SimNodeParam{
			Index: index + 1,
			Type:  "NormalRouter",
			CorruptRatio: params.RatioDistribution{
				Start: float64(lowRatio),
				End:   float64(lowRatio),
			},
			CorruptSpecialPacketRatio: params.RatioDistribution{
				Start: 0,
				End:   0,
			},
		}
		finalMapping[index] = topologyDesc.Nodes[index]
		index += 1
		topologyDesc.Nodes[index] = params.SimNodeParam{
			Index: index + 1,
			Type:  "NormalRouter",
			CorruptRatio: params.RatioDistribution{
				Start: float64(highRatio),
				End:   float64(highRatio),
			},
			CorruptSpecialPacketRatio: params.RatioDistribution{
				Start: 0,
				End:   0,
			},
		}
		finalMapping[index] = topologyDesc.Nodes[index]
		index += 1
		if index == (numberOfNodes - 1) {
			topologyDesc.Nodes[index] = params.SimNodeParam{
				Index: index + 1,
				Type:  "EndHost",
			}
		} else {
			topologyDesc.Nodes[index] = params.SimNodeParam{
				Index: index + 1,
				Type:  "PathValidationRouter",
			}
		}
		finalMapping[index] = topologyDesc.Nodes[index]
		index += 1

		currentHop += 1

		if currentHop >= numberOfHops {
			break
		}
	}
	return finalMapping
}

// FillPvLinks 进行抽象边的填充
func FillPvLinks(topologyDesc *Description, numberOfHops int, indexToNodeParamMapping map[int]params.SimNodeParam) {
	numberOfPvLinks := numberOfHops * 2
	topologyDesc.PvLinks = make([]params.SimAbsLinkParam, numberOfPvLinks)

	currentHop := 0
	currentNodeIndex := 0
	pvLinkIndex := 0
	for {
		topologyDesc.PvLinks[pvLinkIndex] = params.SimAbsLinkParam{
			SourceNode:       indexToNodeParamMapping[currentNodeIndex+1],
			TargetNode:       indexToNodeParamMapping[currentNodeIndex+4],
			IntermediateNode: indexToNodeParamMapping[currentNodeIndex+2],
		}
		pvLinkIndex += 1

		topologyDesc.PvLinks[pvLinkIndex] = params.SimAbsLinkParam{
			SourceNode:       indexToNodeParamMapping[currentNodeIndex+1],
			TargetNode:       indexToNodeParamMapping[currentNodeIndex+4],
			IntermediateNode: indexToNodeParamMapping[currentNodeIndex+3],
		}
		pvLinkIndex += 1
		currentNodeIndex += 4
		currentHop += 1

		if currentHop >= numberOfHops {
			break
		}
	}
}

func FillRealLinks(topologyDesc *Description, numberOfHops int, indexToNodeParamMapping map[int]params.SimNodeParam) {
	numberOfRealLinks := numberOfHops * 4
	topologyDesc.Links = make([]params.SimLinkParam, numberOfRealLinks)

	currentHop := 0
	currentNodeIndex := 0
	realLinkIndex := 0
	for {
		topologyDesc.Links[realLinkIndex] = params.SimLinkParam{
			SourceNode: indexToNodeParamMapping[currentNodeIndex+1],
			TargetNode: indexToNodeParamMapping[currentNodeIndex+2],
		}
		realLinkIndex += 1

		topologyDesc.Links[realLinkIndex] = params.SimLinkParam{
			SourceNode: indexToNodeParamMapping[currentNodeIndex+2],
			TargetNode: indexToNodeParamMapping[currentNodeIndex+4],
		}
		realLinkIndex += 1

		topologyDesc.Links[realLinkIndex] = params.SimLinkParam{
			SourceNode: indexToNodeParamMapping[currentNodeIndex+1],
			TargetNode: indexToNodeParamMapping[currentNodeIndex+3],
		}
		realLinkIndex += 1

		topologyDesc.Links[realLinkIndex] = params.SimLinkParam{
			SourceNode: indexToNodeParamMapping[currentNodeIndex+3],
			TargetNode: indexToNodeParamMapping[currentNodeIndex+4],
		}
		realLinkIndex += 1
		currentNodeIndex += 4
		currentHop += 1

		if currentHop >= numberOfHops {
			break
		}

	}
}
