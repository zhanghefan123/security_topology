package progress_bar

import (
	"github.com/schollz/progressbar/v3"
)

// NewProgressBar 创建指定样式的进度条
func NewProgressBar(max int, description string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowBytes(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(100), // 设置进度条的宽度，确保对齐
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[green][reset]",
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	return bar
}

// Add 进行进度条的推进
func Add(progressBar *progressbar.ProgressBar, number int) {
	_ = progressBar.Add(number)
}
