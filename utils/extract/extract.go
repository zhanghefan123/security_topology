package extract

import (
	"fmt"
	"regexp"
	"strconv"
)

func NumberFromString(input string) (int, error) {
	// 编译正则表达式，匹配一个或多个数字
	re := regexp.MustCompile(`\d+`)

	// 查找匹配项
	match := re.FindString(input)

	if match == "" {
		return 0, fmt.Errorf("未找到数字")
	}

	// 将字符串转换为整数
	num, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}

	return num, nil
}
