package printer

import "fmt"

func PrintMap[V any](m map[string]V) {
	for key, _ := range m {
		fmt.Println(key)
	}
}
