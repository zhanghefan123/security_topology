package performance_monitor

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestInterfaceRate(t *testing.T) {
	str := "eth0:    7122      73    0    0    0     0          0         0     1026      11    0    0    0     0       0          0"
	r := regexp.MustCompile("[^\\s]+")
	res := r.FindAllString(str, -1)
	fmt.Println(res)
	fmt.Println(len(res))
}

func TestBreak(t *testing.T) {
	for {
		select {
		case <-time.After(time.Second):
			fmt.Println("一秒后执行")
			break
		}
	}
}
