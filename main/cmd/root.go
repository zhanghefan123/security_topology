package cmd

import (
	"github.com/spf13/cobra"
)

// CreateRootCmd 创建根命令
func CreateRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "main",
		Short: "Cmd interface of security topology",
		Long:  "Cmd interface of security topology",
	}
	return rootCmd
}
