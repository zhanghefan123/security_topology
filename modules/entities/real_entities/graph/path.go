package graph

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"strings"
)

type Path struct {
	NodeList []*MultipathGraphNode
}

// CreateNewGraphFromRealPaths 通过真正的节点路径创建新的 Graph
func CreateNewGraphFromRealPaths(paths []*Path) (*simple.DirectedGraph, map[string]*MultipathGraphNode, *SourceAndDest) {
	createdGraph := simple.NewDirectedGraph()
	multipathNodeMapping := map[string]*MultipathGraphNode{}
	// 遍历所有的路径来进行节点的添加
	// ----------------------------------------------------------------------------------------------------------------------
	for _, singlePath := range paths {
		// 遍历路径之中的每一个节点
		for _, multipathNode := range singlePath.NodeList {
			if _, ok := multipathNodeMapping[multipathNode.NodeName]; !ok {
				multipathNodeMapping[multipathNode.NodeName] = CreateMultipathGraphNode(multipathNode.NodeName)
				newNode := createdGraph.NewNode()
				multipathNode.Node = newNode
				createdGraph.AddNode(newNode)
			}
		}
	}
	// ----------------------------------------------------------------------------------------------------------------------
	// 通过第一条路径取出 source and dest
	source := paths[0].NodeList[0]
	destination := paths[0].NodeList[len(paths[0].NodeList)-1]
	sourceAndDestination := CreateSourceAndDest(source, destination)
	// 遍历所有的路径来添加边
	// ----------------------------------------------------------------------------------------------------------------------
	for _, singlePath := range paths {
		// 遍历路径之中的每一个节点
		for index, multipathNode := range singlePath.NodeList {
			// 当前节点和下一个节点
			if index == (len(singlePath.NodeList) - 1) {
				break
			} else {
				nextIndex := index + 1
				nextMultipathNode := singlePath.NodeList[nextIndex]
				// 构建从 currentNode 到 nextNode 的一条边
				newEdge := createdGraph.NewEdge(multipathNode, nextMultipathNode)
				createdGraph.SetEdge(newEdge)
			}
		}
	}
	// ----------------------------------------------------------------------------------------------------------------------
	return createdGraph, multipathNodeMapping, sourceAndDestination
}

func CreatePathsFromStrPaths(strPaths [][]string) []*Path {
	var paths []*Path
	multipathNodeMapping := map[string]*MultipathGraphNode{}
	// 遍历所有的路径来进行节点的获取
	// ----------------------------------------------------------------------------------------------------------------------
	for _, nodeStrList := range strPaths {
		// 遍历路径之中的每一个节点
		for _, nodeStr := range nodeStrList {
			if _, ok := multipathNodeMapping[nodeStr]; !ok {
				multipathNodeMapping[nodeStr] = CreateMultipathGraphNode(nodeStr)
			}
		}
	}
	// ----------------------------------------------------------------------------------------------------------------------
	// 构建路径
	// ----------------------------------------------------------------------------------------------------------------------
	for _, nodeStrList := range strPaths {
		singlePath := &Path{}
		// 遍历路径之中的每一个节点
		for _, nodeStr := range nodeStrList {
			// 构建路径
			singlePath.NodeList = append(singlePath.NodeList, multipathNodeMapping[nodeStr])
		}
		paths = append(paths, singlePath)
	}
	// ----------------------------------------------------------------------------------------------------------------------
	return paths
}

// ResolveMultiPathFile 解析多路径文件以得到字符串形式的路径
func ResolveMultiPathFile(multiplePathFile string) ([][]string, error) {
	var pathListInStr [][]string
	data, err := os.ReadFile(multiplePathFile)
	if err != nil {
		return pathListInStr, fmt.Errorf("read multiple path file failed, %s", err.Error())
	}
	pathsInString := strings.Split(string(data), "\n")
	for index, singlePathInString := range pathsInString {
		nodeListInString := strings.Split(singlePathInString, ",")
		pathListInStr[index] = nodeListInString
	}
	return pathListInStr, nil
}
