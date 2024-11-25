package route

import (
	"fmt"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/utils/dir"
)

// GenerateLiRRoutingString 为 path_validation 到某个地点生成路由
func GenerateLiRRoutingString(sourceNode, destinationNode int64, linkIdentifiers, nodeIds []int) string {
	length := len(linkIdentifiers)
	finalString := ""
	for index := 0; index < length; index++ {
		if index != length-1 {
			finalString += fmt.Sprintf("%d,%d,", linkIdentifiers[index], nodeIds[index])
		} else {
			finalString += fmt.Sprintf("%d,%d", linkIdentifiers[index], nodeIds[index])
		}
	}
	return fmt.Sprintf("%d,%d,%d,%s", sourceNode, destinationNode, len(linkIdentifiers), finalString)
}

// GenerateLiRRoutingStrings 为 path_validation 到所有目的节点生成路由
func GenerateLiRRoutingStrings(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) ([]string, error) {
	var err error
	var linkIdentifier int
	var sourceNode int64
	var destinationNode int64
	var finalResult []string
	shortestPath := path.DijkstraFrom(abstractNode, graphTmp)
	iterator := graphTmp.Nodes()
	for {
		linkIdentifiers := make([]int, 0)
		nodeIds := make([]int, 0)
		hasNext := iterator.Next()
		if !hasNext {
			break
		}
		currentDestination := iterator.Node()
		if currentDestination.ID() != abstractNode.Node.ID() {
			hopList, _ := shortestPath.To(currentDestination.ID())
			sourceNode = hopList[0].ID() + 1                   // 这里使用 + 1 的原因是 graph Node 的 ID 从 0 开始
			destinationNode = hopList[len(hopList)-1].ID() + 1 // 这里使用 + 1 的原因是 graph Node 的 ID 从 0 开始
			for index := 0; index < len(hopList)-1; index++ {
				sourceIndex := index
				targetIndex := index + 1
				_, linkIdentifier, err = GetAbstractLink(hopList, sourceIndex, targetIndex, linksMap)
				if err != nil {
					return finalResult, fmt.Errorf("get abstract link failed with err %w", err)
				}
				linkIdentifiers = append(linkIdentifiers, linkIdentifier)
				nodeIds = append(nodeIds, int(hopList[targetIndex].ID()+1))
			}
		} else {
			continue
		}
		generateLiRRouteString := GenerateLiRRoutingString(sourceNode, destinationNode, linkIdentifiers, nodeIds)
		finalResult = append(finalResult, generateLiRRouteString)
	}
	return finalResult, nil
}

// WriteLiRRoutingStringsIntoFile 将生成的路由条目写到文件之中
func WriteLiRRoutingStringsIntoFile(containerName string, LiRRoutingStringList []string) error {
	var err error
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
	filePath := filepath.Join(outputDir, "path_validation.txt")
	// 创建写入文件
	var lirRouteFile *os.File
	lirRouteFile, err = os.Create(filePath)
	defer func() {
		closeErr := lirRouteFile.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err != nil {
		return fmt.Errorf("write route error: %w", err)
	}
	// 进行实际的内容的写入
	_, err = lirRouteFile.WriteString(strings.Join(LiRRoutingStringList, "\n"))
	if err != nil {
		return fmt.Errorf("write route error: %w", err)
	}
	return nil
}

// GenerateLiRRoute 生成单个 LiR 节点的路由条目
func GenerateLiRRoute(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) (string, error) {
	var err error
	var lirRouteStrings []string
	lirRouteStrings, err = GenerateLiRRoutingStrings(abstractNode, linksMap, graphTmp)
	if err != nil {
		return "", fmt.Errorf("generate path_validation routing strings failed: %w", err)
	}
	return strings.Join(lirRouteStrings, "\n"), nil
}

// CalculateAndWriteLiRRoutes 进行路由的计算以及文件的写入
func CalculateAndWriteLiRRoutes(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) error {
	var err error
	var lirRouteStrings []string
	lirRouteStrings, err = GenerateLiRRoutingStrings(abstractNode, linksMap, graphTmp)
	if err != nil {
		return fmt.Errorf("generate path_validation routing strings failed: %w", err)
	}
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("get normal node failed: %w", err)
	}
	err = WriteLiRRoutingStringsIntoFile(normalNode.ContainerName, lirRouteStrings)
	if err != nil {
		return fmt.Errorf("write path_validation routing strings failed: %w", err)
	}
	return nil
}
