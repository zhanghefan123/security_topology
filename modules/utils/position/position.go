package position

import (
	"time"
	"unicode"
)

// GetYearAndDay 获取时间的年份和日期
func GetYearAndDay(currentTime time.Time) (int, float64) {
	year := currentTime.Year()

	// 计算天数的小数部分
	day := float64(currentTime.Nanosecond()) / 1e6 // 将纳秒转换为毫秒
	day += float64(currentTime.Second())           // 将秒数加到毫秒上
	day /= 60                                      // 转换为分钟
	day += float64(currentTime.Minute())           // 将分钟加到结果中
	day /= 60                                      // 转换为小时
	day += float64(currentTime.Hour())             // 将小时加到结果中
	day /= 24                                      // 转换为天数

	// 计算当前日期是该年的第几天
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, currentTime.Location())
	daysSinceStartOfYear := currentTime.Sub(startOfYear).Hours() / 24

	// 加上已经过去的天数
	day += daysSinceStartOfYear

	return year % 100, day
}

// TleCheckSum 进行校验和的计算
func TleCheckSum(tle string) int {
	sumNum := 0
	for _, c := range tle {
		// 检查字符是否是数字
		if unicode.IsDigit(c) {
			sumNum += int(c - '0') // 将字符转换为数字
		} else if c == '-' {
			// 负号计算
			sumNum += 1
		}
	}
	// 返回校验和的最后一位
	return sumNum % 10
}
