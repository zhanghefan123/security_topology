package test

import (
	"math/rand"
	"testing"
	"time"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/steps"
)

// 初始化随机种子（保证每次运行随机结果不同）
func init() {
	rand.Seed(time.Now().UnixNano())
}

// TestSampleDiscrete_Basic 测试基础功能：正确返回下标
func TestSampleDiscrete_Basic(t *testing.T) {
	// 测试用例1：只有1个概率，必须返回0
	prob1 := []float64{1.0}
	if res := steps.SampleDiscrete(prob1); res != 0 {
		t.Errorf("单一概率应返回0，实际返回%d", res)
	}

	// 测试用例2：第一个概率=1，永远返回0
	prob2 := []float64{1.0, 0.0, 0.0}
	if res := steps.SampleDiscrete(prob2); res != 0 {
		t.Errorf("100%%概率应返回0，实际返回%d", res)
	}

	// 测试用例3：最后一个概率=1，永远返回最后一个下标
	prob3 := []float64{0.0, 0.0, 1.0}
	if res := steps.SampleDiscrete(prob3); res != 2 {
		t.Errorf("100%%概率应返回2，实际返回%d", res)
	}
}

// TestSampleDiscrete_Distribution 统计测试：概率分布符合预期
func TestSampleDiscrete_Distribution(t *testing.T) {
	// 测试分布 [0.2, 0.5, 0.3]
	probs := []float64{0.2, 0.5, 0.3}
	const times = 100000 // 采样10万次

	count0 := 0
	count1 := 0
	count2 := 0

	for i := 0; i < times; i++ {
		switch steps.SampleDiscrete(probs) {
		case 0:
			count0++
		case 1:
			count1++
		case 2:
			count2++
		}
	}

	// 计算频率
	rate0 := float64(count0) / times
	rate1 := float64(count1) / times
	rate2 := float64(count2) / times

	// 允许 1% 误差
	t.Logf("分布结果：0=%.2f%%, 1=%.2f%%, 2=%.2f%%", rate0*100, rate1*100, rate2*100)
	if rate0 < 0.19 || rate0 > 0.21 {
		t.Errorf("0号概率异常：期望0.2，实际%.2f", rate0)
	}
	if rate1 < 0.49 || rate1 > 0.51 {
		t.Errorf("1号概率异常：期望0.5，实际%.2f", rate1)
	}
	if rate2 < 0.29 || rate2 > 0.31 {
		t.Errorf("2号概率异常：期望0.3，实际%.2f", rate2)
	}
}

// TestSampleDiscrete_FloatError 测试浮点误差兜底逻辑
func TestSampleDiscrete_FloatError(t *testing.T) {
	// 总和接近1，但因为浮点精度可能进不去循环内return
	probs := []float64{0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1}

	// 多次调用，确保不会崩溃，永远返回合法下标
	for i := 0; i < 1000; i++ {
		res := steps.SampleDiscrete(probs)
		if res < 0 || res >= len(probs) {
			t.Errorf("下标越界：%d", res)
		}
	}
}
