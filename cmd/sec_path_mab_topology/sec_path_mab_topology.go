package sec_path_mab_topology

import (
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/cmd/tools"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/topology"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	secPathTopologyLogger = logger.GetLogger(logger.ModuleMainCmdSecPathTopology)
)

// CreateSecPathMabTopologyCommand 使用方式 ./cmd sec_path_mab -s 3 -lr 50000 -hr 250000
func CreateSecPathMabTopologyCommand() *cobra.Command {
	var secPathMabTopologyCommand = &cobra.Command{
		Use:   "sec_path_mab",
		Short: "sec_path_mab",
		Long:  "sec_path_mab",
		Run: func(cmd *cobra.Command, args []string) {
			secPathTopologyLogger.Infof("start sec_path_mab")
			core()
		},
	}
	tools.AttachFlags(secPathMabTopologyCommand, []string{tools.FlagNameOfSecPathMabHops,
		tools.FlagNameOfLowRatio, tools.FlagNameOfHighRatio})
	return secPathMabTopologyCommand
}

// core sec_path_mab
func core() {
	topologyDescription := topology.GenerateTopologyDescription(variables.UserSelectedNumberOfHops,
		variables.UserSelectedLowRatio,
		variables.UserSelectedHighRatio)
	topology.MarshalTopologyDescription(topologyDescription)
}
