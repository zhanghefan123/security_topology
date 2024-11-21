package tools

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"zhanghefan123/security_topology/cmd/variables"
)

// InitFlagSet 初始化可用的选项的集合
func InitFlagSet() *pflag.FlagSet {
	flags := &pflag.FlagSet{}
	flags.StringVarP(&variables.UserSelectedImage,
		FlagNameOfImageName,
		FlagNameShortHandOfImageName,
		variables.UserSelectedImage,
		fmt.Sprintf("available images: %s %s %s %s %s %s %s %s %s %s %s",
			variables.ImageUbuntuWithSoftware, variables.ImagePythonEnv,
			variables.ImageGoEnv, variables.ImageNormalSatellite,
			variables.ImageNameEtcd, variables.ImageNamePosition,
			variables.ImageNameRouter, variables.ImageNameNormalNode,
			variables.ImageNameConsensusNode, variables.ImageNameMaliciousNode,
			variables.ImageLiRNode))
	flags.StringVarP(&variables.UserSelectedOperation,
		FlagNameOfOperationType,
		FlagNameShortHandOfOperationType,
		variables.UserSelectedOperation,
		fmt.Sprintf("available operations: %s %s %s",
			variables.OperationBuild, variables.OperationRebuild,
			variables.OperationRemove))
	return flags
}

// AttachFlags 将选项放到对应的命令上
func AttachFlags(cmd *cobra.Command, flagNames []string) {
	initializedFlags := InitFlagSet()
	cmdFlags := cmd.Flags()
	for _, flagName := range flagNames {
		if flag := initializedFlags.Lookup(flagName); flag != nil {
			cmdFlags.AddFlag(flag)
			_ = cmd.MarkFlagRequired(flagName)
		}
	}
}
