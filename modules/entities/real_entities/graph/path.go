package graph

import (
	"fmt"

	"gonum.org/v1/gonum/graph/simple"
)

type Path struct {
	NodeList []*Node
	Weight   float64
}

// CreateNewGraphFromRealPaths 通过真正的节点路径创建新的 Graph
func CreateNewGraphFromRealPaths(paths []*Path) (*simple.DirectedGraph, map[string]*Node, *SourceAndDest) {
	createdGraph := simple.NewDirectedGraph()
	multipathNodeMapping := map[string]*Node{}
	// 遍历所有的路径来进行节点的添加
	// ----------------------------------------------------------------------------------------------------------------------
	for _, singlePath := range paths {
		// 遍历路径之中的每一个节点
		for _, multipathNode := range singlePath.NodeList {
			if _, ok := multipathNodeMapping[multipathNode.NodeName]; !ok {
				multipathNodeMapping[multipathNode.NodeName] = multipathNode
				newNode := createdGraph.NewNode()
				multipathNode.Node = newNode
				createdGraph.AddNode(multipathNode)
			}
		}
	}
	// ----------------------------------------------------------------------------------------------------------------------
	// 通过第一条路径取出 source and dest
	//fmt.Println("------------------------------------------------------")
	for _, singlePath := range paths {
		finalPathStr := ""
		for index, node := range singlePath.NodeList {
			if index != len(singlePath.NodeList)-1 {
				finalPathStr = finalPathStr + node.NodeName + "->"
			} else {
				finalPathStr = finalPathStr + node.NodeName
			}
		}
		//fmt.Printf("path-%d: %v\n", pathIndex, finalPathStr)
	}
	//fmt.Println("------------------------------------------------------")
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
func PathToString(path *Path) string {
	finalString := ""
	for index, singleNode := range path.NodeList {
		if index != (len(path.NodeList) - 1) {
			finalString = finalString + singleNode.NodeName + "->"
		} else {
			finalString = finalString + singleNode.NodeName
		}
	}
	return finalString
}

func PrintPath(path *Path, pathIndex int) {
	finalString := ""
	for index, singleNode := range path.NodeList {
		if index != (len(path.NodeList) - 1) {
			finalString = finalString + singleNode.NodeName + "->"
		} else {
			finalString = finalString + singleNode.NodeName
		}
	}
	fmt.Printf("path-%d: %v\n", pathIndex, finalString)
}

func PrintPaths(paths []*Path) {
	for index, singlePath := range paths {
		PrintPath(singlePath, index)
	}
}
