package graph

import (
	"fmt"
	"gonum.org/v1/gonum/graph/topo"
	"sort"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
	"zhanghefan123/security_topology/utils/file"
)

func PrunePaths(atlasComplexGraph *Graph, requiredNumberOfPaths int) {

	// 按路径长度排序（从长到短）
	sort.Slice(atlasComplexGraph.KShortestPaths, func(i, j int) bool {
		return len(atlasComplexGraph.KShortestPaths[i].NodeList) >
			len(atlasComplexGraph.KShortestPaths[j].NodeList)
	})

	removeCount := len(atlasComplexGraph.KShortestPaths) - requiredNumberOfPaths

	// 尝试所有可能的删除组合（按长度排序后的前removeCount条）
	attempts := 0
	_ = combinations(len(atlasComplexGraph.KShortestPaths), removeCount)

	findValidateSet := false

	// 生成所有可能的删除组合
	for combination := range generateCombinations(len(atlasComplexGraph.KShortestPaths), removeCount) {
		attempts++
		//fmt.Printf("尝试组合 %d/%d: 删除索引 ", attempts, maxAttempts)

		// 构建删除后的路径集合
		remainingPaths := make([]*entities.Path, 0)
		removedIndices := make([]int, 0)

		// 标记要删除的索引
		toRemove := make(map[int]bool)
		for _, idx := range combination {
			toRemove[idx] = true
			removedIndices = append(removedIndices, idx)
			//fmt.Printf("%d ", idx)
		}
		//fmt.Println()

		// 收集保留的路径
		for idx, path := range atlasComplexGraph.KShortestPaths {
			if !toRemove[idx] {
				remainingPaths = append(remainingPaths, path)
			}
		}

		// 构建图并检查是否有环
		createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(remainingPaths)
		_, err := topo.Sort(createdGraph)

		if err != nil {
			//fmt.Printf("  仍有环，继续尝试...\n")
		} else {
			findValidateSet = true
			//fmt.Printf("  无环！找到有效路径集合\n")

			// 更新KShortestPaths为找到的有效集合
			atlasComplexGraph.KShortestPaths = remainingPaths

			// 打印结果信息
			//fmt.Printf("成功: 删除 %d 条路径后获得无环图\n", removeCount)
			//fmt.Printf("删除的路径索引: %v\n", removedIndices)
			//fmt.Printf("剩余路径数: %d\n", len(remainingPaths))

			// 打印剩余路径信息
			//for i, path := range remainingPaths {
			//fmt.Printf("路径 %d: 长度=%d\n", i+1, len(path.NodeList))
			//}

			break
		}
	}

	if !findValidateSet {
		//fmt.Printf("未找到合理路径\n")
	}

}

func GenerateDifferentNumberOfPaths() error {
	// 1. 创建图
	atlasComplexTopologyFilePath := "C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\atlas_complex_topology.json"
	atlasComplexGraph := CreateGraph(atlasComplexTopologyFilePath, "")
	// 2. 图初始化
	err := atlasComplexGraph.Init()
	if err != nil {
		//fmt.Printf("create graph error: %v", err)
	}
	// 3. 计算 length 条最短路
	atlasComplexGraph.GraphParams.NumberOfPaths = 30
	err = atlasComplexGraph.CalculateKShortestPaths()
	if err != nil {
		//fmt.Printf("calculate shortest paths failed: %v\n", err)
	}
	originalPaths := make([]*entities.Path, len(atlasComplexGraph.KShortestPaths))
	copy(originalPaths, atlasComplexGraph.KShortestPaths)
	// 4. 进行遍历生成
	for index := 2; index <= 20; index += 2 {
		// 5. 原始路径备份
		backupPaths := make([]*entities.Path, len(originalPaths))
		copy(backupPaths, originalPaths)
		//fmt.Printf("backup paths length = %d\n", len(backupPaths))
		// 4. 基于创建的路径进行图的生成
		createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(backupPaths)
		atlasComplexGraph.DirectedGraph = createdGraph
		atlasComplexGraph.KShortestPaths = backupPaths
		// 5. 进行路径的裁剪
		PrunePaths(atlasComplexGraph, index)
		// 6. 进行路径的打印
		filePath := fmt.Sprintf("C:\\zhf_projects\\security\\security_topology\\resources\\multipath\\complex\\multipath_%d.txt", index)
		err = WritePathsIntoFile(filePath, atlasComplexGraph.KShortestPaths)
		if err != nil {
			return fmt.Errorf("fail to write %s\n", filePath)
		}
	}
	return nil
}

func WritePathsIntoFile(destinationFile string, paths []*entities.Path) error {
	finalString := ""
	for index, path := range paths {
		if index != (len(paths) - 1) {
			finalString += fmt.Sprintf("%s\n", entities.PathToString(path))
		} else {
			finalString += entities.PathToString(path)
		}
	}
	err := file.WriteStringIntoFile(destinationFile, finalString)
	if err != nil {
		return fmt.Errorf("write string into file error: %v", err)
	}
	return nil
}

func CalculateNoLoopKshortestPaths(atlasComplexGraph *Graph) {
	// 当前判断是否有环路, 如果没有环路直接退出即可
	_, err := topo.Sort(atlasComplexGraph.DirectedGraph)
	if err == nil {
		fmt.Println("directly return")
		return
	}

	// 按路径长度排序（从长到短）
	sort.Slice(atlasComplexGraph.KShortestPaths, func(i, j int) bool {
		return len(atlasComplexGraph.KShortestPaths[i].NodeList) >
			len(atlasComplexGraph.KShortestPaths[j].NodeList)
	})

	maxRemovalCount := len(atlasComplexGraph.KShortestPaths)
	foundValidSet := false

	// 从删除1条路径开始尝试，逐步增加
	for removeCount := 1; removeCount <= maxRemovalCount && !foundValidSet; removeCount++ {
		//fmt.Printf("\n尝试删除 %d 条最长路径...\n", removeCount)

		// 尝试所有可能的删除组合（按长度排序后的前removeCount条）
		attempts := 0
		_ = combinations(len(atlasComplexGraph.KShortestPaths), removeCount)

		// 生成所有可能的删除组合
		for combination := range generateCombinations(len(atlasComplexGraph.KShortestPaths), removeCount) {
			attempts++
			//fmt.Printf("尝试组合 %d/%d: 删除索引 ", attempts, maxAttempts)

			// 构建删除后的路径集合
			remainingPaths := make([]*entities.Path, 0)
			removedIndices := make([]int, 0)

			// 标记要删除的索引
			toRemove := make(map[int]bool)
			for _, idx := range combination {
				toRemove[idx] = true
				removedIndices = append(removedIndices, idx)
				//fmt.Printf("%d ", idx)
			}
			fmt.Println()

			// 收集保留的路径
			for idx, path := range atlasComplexGraph.KShortestPaths {
				if !toRemove[idx] {
					remainingPaths = append(remainingPaths, path)
				}
			}

			// 构建图并检查是否有环
			createdGraph, _, _ := entities.CreateNewGraphFromRealPaths(remainingPaths)
			_, err = topo.Sort(createdGraph)

			if err != nil {
				//fmt.Printf("  仍有环，继续尝试...\n")
			} else {
				//fmt.Printf("  无环！找到有效路径集合\n")

				// 更新KShortestPaths为找到的有效集合
				atlasComplexGraph.KShortestPaths = remainingPaths
				foundValidSet = true

				// 打印结果信息
				//fmt.Printf("成功: 删除 %d 条路径后获得无环图\n", removeCount)
				//fmt.Printf("删除的路径索引: %v\n", removedIndices)
				//fmt.Printf("剩余路径数: %d\n", len(remainingPaths))

				// 打印剩余路径信息
				//for i, path := range remainingPaths {
				//fmt.Printf("路径 %d: 长度=%d\n", i+1, len(path.NodeList))
				//}

				break
			}
		}

		if !foundValidSet {
			//fmt.Printf("删除 %d 条路径的所有组合尝试完毕，均有环\n", removeCount)
		}
	}

	if !foundValidSet {
		//fmt.Printf("警告: 无法找到无环的路径集合，返回空路径\n")
		atlasComplexGraph.KShortestPaths = []*entities.Path{}
	}

}

// 组合数学函数：计算组合数 C(n, k)
func combinations(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	// 使用较小的k值计算
	if k > n-k {
		k = n - k
	}

	result := 1
	for i := 1; i <= k; i++ {
		result = result * (n - k + i) / i
	}
	return result
}

// 生成所有组合的迭代器
func generateCombinations(n, k int) <-chan []int {
	ch := make(chan []int)

	go func() {
		defer close(ch)

		if k <= 0 || k > n {
			return
		}

		// 初始化第一个组合
		combination := make([]int, k)
		for i := 0; i < k; i++ {
			combination[i] = i
		}

		// 发送第一个组合
		ch <- append([]int{}, combination...)

		// 生成后续组合
		for {
			// 找到可以增加的位置
			i := k - 1
			for i >= 0 && combination[i] == n-k+i {
				i--
			}

			if i < 0 {
				break // 所有组合已生成
			}

			// 增加该位置的值
			combination[i]++

			// 重置后面的位置
			for j := i + 1; j < k; j++ {
				combination[j] = combination[j-1] + 1
			}

			// 发送组合
			ch <- append([]int{}, combination...)
		}
	}()

	return ch
}
