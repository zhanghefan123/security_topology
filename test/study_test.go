package test

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestStudy(t *testing.T) {
	result := make(map[string]int, 0)
	result["age"] = 18
	bytes, err := json.Marshal(result)
	if err != nil {
		t.Error(err)
	}
	var mapAfterUnmarshal = make(map[string]any, 0)
	err = json.Unmarshal(bytes, &mapAfterUnmarshal)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(mapAfterUnmarshal)
	fmt.Printf("%T", mapAfterUnmarshal["age"])

	fmt.Println()
}

// 输出结果
// === RUN   TestStudy
// map[age:18]
// float64
// --- PASS: TestStudy (0.00s)
// PASS
