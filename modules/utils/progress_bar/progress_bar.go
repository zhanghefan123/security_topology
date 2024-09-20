package progress_bar

import (
	"github.com/schollz/progressbar/v3"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleProgressBar = logger.GetLogger(logger.ModuleProgressBar)

func NewProgressBar(max int, description string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowBytes(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
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

func Add(progressBar *progressbar.ProgressBar, number int) {
	err := progressBar.Add(number)
	if err != nil {
		moduleProgressBar.Errorf("progress bar error")
	}
}
