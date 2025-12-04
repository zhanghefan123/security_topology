package route

import (
	"fmt"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"net"
	"os"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/utils/dir"
)

// GenerateSegmentRoutingString 到单个节点的静态路由的生成
func GenerateSegmentRoutingString(destinationIp string, ipSegmentList *[]string, interfaceName string) string {
	result := strings.Join(*ipSegmentList, ",")
	return fmt.Sprintf("/bin/ip -6 route add %s encap seg6 mode encap segs %s dev %s", destinationIp, result, interfaceName)
}

// GenerateSegmentRoutingStrings  到所有节点的静态路由的生成
func GenerateSegmentRoutingStrings(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) ([]string, map[string][]string, map[string]int, error) {
	var err error
	var finalResult []string
	var destinationIpv6AddressMapping = map[string][]string{}
	var destinationPathLengthMapping = map[string]int{}
	shortestPath := path.DijkstraFrom(abstractNode, graphTmp)
	iterator := graphTmp.Nodes()
	for {
		hasNext := iterator.Next()
		if !hasNext {
			break
		}
		currentDestination := iterator.Node()
		if currentDestination.ID() != abstractNode.Node.ID() {
			// 拿到目的节点的名称
			// -----------------------------------------
			var currentDestinationNormal *normal_node.NormalNode
			currentDestinationNormal, err = GetNormalNodeFromGraphNode(currentDestination)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("calcualate route error: %w", err)
			}
			// 不用计算到地面站的路径
			if currentDestinationNormal.Type == types.NetworkNodeType_GroundStation {
				continue
			}
			// -----------------------------------------
			ipSegmentList := make([]string, 0)
			// 在这里进行路由的计算
			hopList, _ := shortestPath.To(currentDestination.ID())
			if len(hopList) == 2 {
				continue
			}
			// 出接口的名称
			var outputInterfaceName string
			// 目的地址
			var destination string
			// 通过 hopList 找到对应的 linkList
			for index := 0; index < len(hopList)-1; index++ {
				var sourceNormal *normal_node.NormalNode
				var targetNormal *normal_node.NormalNode
				source := hopList[index]
				target := hopList[index+1]
				sourceNormal, err = GetNormalNodeFromGraphNode(source)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("calcualate route error: %w", err)
				}
				targetNormal, err = GetNormalNodeFromGraphNode(target)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("calcualate route error: %w", err)
				}
				// 找到相应的链路 -> 带有方向的
				isl := (*linksMap)[sourceNormal.ContainerName][targetNormal.ContainerName]
				if isl != nil { // 如果不为空，说明方向是对的
					ip, _, _ := net.ParseCIDR(isl.TargetInterface.SourceIpv6Addr) // 这里使用的是 目标的ip 因为是正向的
					ipSegmentList = append(ipSegmentList, ip.String())

					if index == 0 {
						outputInterfaceName = isl.SourceInterface.IfName
					}

					// 最后一个链路
					if index == len(hopList)-2 {
						ip, _, _ = net.ParseCIDR(isl.TargetInterface.SourceIpv6Addr) // 这里使用的是 目标的ip 因为是正向的
						destination = ip.String()
						// 将结果进行存储
						destinationIpv6AddressMapping[currentDestinationNormal.ContainerName] = []string{destination, isl.TargetInterface.IfName}
					}
				} else { // 如果为空，说明方向是反的
					isl = (*linksMap)[targetNormal.ContainerName][sourceNormal.ContainerName]
					ip, _, _ := net.ParseCIDR(isl.SourceInterface.SourceIpv6Addr) // 这里使用的是 源的ip 因为是反向的
					ipSegmentList = append(ipSegmentList, ip.String())

					if index == 0 {
						outputInterfaceName = isl.TargetInterface.IfName
					}

					// 最后一个链路
					if index == len(hopList)-2 {
						ip, _, _ = net.ParseCIDR(isl.SourceInterface.SourceIpv6Addr) // 这里使用的是 源的ip 因为是反向的
						destination = ip.String()
						// 将结果进行存储
						destinationIpv6AddressMapping[currentDestinationNormal.ContainerName] = []string{destination, isl.SourceInterface.IfName}
					}
				}
			}

			destinationPathLengthMapping[currentDestinationNormal.ContainerName] = len(ipSegmentList)

			generateIpRouteString := GenerateSegmentRoutingString(destination, &ipSegmentList, outputInterfaceName)
			finalResult = append(finalResult, generateIpRouteString)
		}
	}
	return finalResult, destinationIpv6AddressMapping, destinationPathLengthMapping, nil
}

// WriteSegmentRoutingStringsIntoFile 将段路由信息写入到文件之中
func WriteSegmentRoutingStringsIntoFile(containerName string, IPRouteStringList []string) (err error) {
	// simulation 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	// route dir 文件的位置
	outputDir := filepath.Join(simulationDir, containerName, "route")
	// 进行文件夹的生成
	err = dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("write route error: %w", err)
	}
	// 文件的路径
	filePath := filepath.Join(outputDir, "srv6.txt")
	// 创建写入文件
	var ipv6SegmentRouteFile *os.File
	ipv6SegmentRouteFile, err = os.Create(filePath)
	defer func() {
		closeErr := ipv6SegmentRouteFile.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err != nil {
		return fmt.Errorf("calculate route error: %w", err)
	}
	// 进行实际的内容的写入
	_, err = ipv6SegmentRouteFile.WriteString(strings.Join(IPRouteStringList, "\n"))
	if err != nil {
		return fmt.Errorf("write route error: %w", err)
	}
	return nil
}

// WriteIpv6DestinationMappingIntoFile 将映射写入文件之中
func WriteIpv6DestinationMappingIntoFile(containerName string, mapping map[string][]string) error {
	// 最终的写入文件
	// simulation 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	// route dir 文件的位置
	outputDir := filepath.Join(simulationDir, containerName, "route")
	// 文件的路径
	filePath := filepath.Join(outputDir, "ipv6_destination.txt")
	// 进行文件夹的生成
	err := dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("cannot generate output dir: %w", err)
	}
	// 创建写入文件
	var ipv6DestinationMappingFile *os.File
	ipv6DestinationMappingFile, err = os.Create(filePath)
	defer func() {
		closeErr := ipv6DestinationMappingFile.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err != nil {
		return fmt.Errorf("calculate ipv6 destination mapping error: %w", err)
	}
	// 进行实际的文件的写入
	finalString := ""
	for key, value := range mapping {
		finalString += fmt.Sprintf("%s->%s->%s\n", key, value[0], value[1])
	}
	_, err = ipv6DestinationMappingFile.WriteString(finalString)
	if err != nil {
		return fmt.Errorf("write ipv6 destination mapping error: %w", err)
	}
	return nil
}

func WriteDestinationPathLengthMappingIntoFile(containerName string, destinationPathLengthMapping map[string]int) (err error) {
	// simulation 文件夹的位置
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	// route dir 文件的位置
	outputDir := filepath.Join(simulationDir, containerName, "route")
	// 进行文件夹的生成
	err = dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("write route error: %w", err)
	}
	// 文件的路径
	filePath := filepath.Join(outputDir, "destination_path_length.txt")
	// 创建写入文件
	var ipv6SegmentRouteFile *os.File
	ipv6SegmentRouteFile, err = os.Create(filePath)
	defer func() {
		closeErr := ipv6SegmentRouteFile.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err != nil {
		return fmt.Errorf("calculate ipv6 destination path length mapping error: %w", err)
	}
	finalString := ""
	for key, value := range destinationPathLengthMapping {
		finalString += fmt.Sprintf("%s->%d\n", key, value)
	}
	// write content
	_, err = ipv6SegmentRouteFile.WriteString(finalString)
	if err != nil {
		return fmt.Errorf("write destination path length mapping error: %w", err)
	}
	return nil
}

// CalculateAndWriteSegmentRoute 进行到其他节点的段路由的计算
func CalculateAndWriteSegmentRoute(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) error {
	var err error
	var ipRouteStrings []string
	var destinationIpv6AddressMapping map[string][]string
	var destinationPathLengthMapping map[string]int
	ipRouteStrings, destinationIpv6AddressMapping, destinationPathLengthMapping, err = GenerateSegmentRoutingStrings(abstractNode, linksMap, graphTmp)
	if err != nil {
		return fmt.Errorf("generate segment routing strings failed: %w", err)
	}
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get normal node from abstract node failed: %w", err)
	}
	err = WriteSegmentRoutingStringsIntoFile(normalNode.ContainerName, ipRouteStrings)
	if err != nil {
		return fmt.Errorf("write segment routing strings into file failed: %w", err)
	}
	err = WriteIpv6DestinationMappingIntoFile(normalNode.ContainerName, destinationIpv6AddressMapping)
	if err != nil {
		return fmt.Errorf("write ipv6 destination mapping failed: %w", err)
	}
	// output to each destination's path length
	err = WriteDestinationPathLengthMappingIntoFile(normalNode.ContainerName, destinationPathLengthMapping)
	if err != nil {
		return fmt.Errorf("write destination path length mapping failed: %w", err)
	}
	return nil
}
