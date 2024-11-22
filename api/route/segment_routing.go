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
	"zhanghefan123/security_topology/modules/utils/dir"
)

// GenerateSegmentRoutingString 到单个节点的静态路由的生成
func GenerateSegmentRoutingString(destinationIp string, ipSegmentList *[]string, interfaceName string) string {
	result := strings.Join(*ipSegmentList, ",")
	return fmt.Sprintf("/bin/ip -6 route add %s encap seg6 mode encap segs %s dev %s", destinationIp, result, interfaceName)
}

// GenerateSegmentRoutingStrings  到所有节点的静态路由的生成
func GenerateSegmentRoutingStrings(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) ([]string, error) {
	var err error
	var finalResult []string
	shortestPath := path.DijkstraFrom(abstractNode, graphTmp)
	iterator := graphTmp.Nodes()
	for {
		hasNext := iterator.Next()
		if !hasNext {
			break
		}
		currentDestination := iterator.Node()
		if currentDestination.ID() != abstractNode.Node.ID() {
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
					return nil, fmt.Errorf("calcualate route error: %w", err)
				}
				targetNormal, err = GetNormalNodeFromGraphNode(target)
				if err != nil {
					return nil, fmt.Errorf("calcualate route error: %w", err)
				}
				// 找到相应的链路 -> 带有方向的
				isl := (*linksMap)[sourceNormal.ContainerName][targetNormal.ContainerName]
				if isl != nil { // 如果不为空，说明方向是对的
					ip, _, _ := net.ParseCIDR(isl.TargetInterface.Ipv6Addr)
					ipSegmentList = append(ipSegmentList, ip.String())

					if index == 0 {
						outputInterfaceName = isl.SourceInterface.IfName
					}

					// 最后一个链路
					if index == len(hopList)-2 {
						ip, _, _ = net.ParseCIDR(isl.TargetInterface.Ipv6Addr)
						destination = ip.String()
					}
				} else { // 如果为空，说明方向是反的
					isl = (*linksMap)[targetNormal.ContainerName][sourceNormal.ContainerName]
					ip, _, _ := net.ParseCIDR(isl.SourceInterface.Ipv6Addr)
					ipSegmentList = append(ipSegmentList, ip.String())

					if index == 0 {
						outputInterfaceName = isl.TargetInterface.IfName
					}

					// 最后一个链路
					if index == len(hopList)-2 {
						ip, _, _ = net.ParseCIDR(isl.SourceInterface.Ipv6Addr)
						destination = ip.String()
					}
				}

			}
			generateIpRouteString := GenerateSegmentRoutingString(destination, &ipSegmentList, outputInterfaceName)
			finalResult = append(finalResult, generateIpRouteString)
		}
	}
	return finalResult, nil
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

// CalculateAndWriteSegmentRoute 进行到其他节点的段路由的计算
func CalculateAndWriteSegmentRoute(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) error {
	var err error
	var ipRouteStrings []string
	ipRouteStrings, err = GenerateSegmentRoutingStrings(abstractNode, linksMap, graphTmp)
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
	return nil
}
