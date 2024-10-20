package root

import (
	"github.com/spf13/cobra"
)

// CreateRootCmd 创建根命令
func CreateRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "main",
		Short: "Cmd interface of security images",
		Long:  "Cmd interface of security images",
	}
	return rootCmd
}
