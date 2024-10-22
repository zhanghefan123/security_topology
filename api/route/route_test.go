package route

import (
	"fmt"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"testing"
)

// TestGraph 进行图的测试
func TestGraph(t *testing.T) {
	// 创建一个新的有向图
	graph := simple.NewDirectedGraph()

	// 进行节点的创建
	nodeA := graph.NewNode()
	graph.AddNode(nodeA)
	nodeB := graph.NewNode()
	graph.AddNode(nodeB)
	nodeC := graph.NewNode()
	graph.AddNode(nodeC)

	// 进行边的创建
	edgeAB := graph.NewEdge(nodeA, nodeB)
	edgeBC := graph.NewEdge(nodeB, nodeC)

	// 进行边的添加
	graph.SetEdge(edgeAB)
	graph.SetEdge(edgeBC)

	// 进行最短路径的查找
	shortestPath := path.DijkstraFrom(nodeA, graph)
	pathToB, weight := shortestPath.To(nodeB.ID())
	fmt.Println(pathToB, weight)
}
