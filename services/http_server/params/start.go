package params

import (
	"fmt"
	"time"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/services/http_server/params/apps"
	"zhanghefan123/security_topology/services/http_server/params/blockchain"
	"zhanghefan123/security_topology/services/http_server/params/constellation"
	"zhanghefan123/security_topology/services/http_server/params/images"
	"zhanghefan123/security_topology/services/http_server/params/network"
	"zhanghefan123/security_topology/services/http_server/params/path_validation"
	"zhanghefan123/security_topology/services/http_server/params/raspberrypi"
	"zhanghefan123/security_topology/services/http_server/params/resources"
	"zhanghefan123/security_topology/services/http_server/params/services"
	"zhanghefan123/security_topology/services/http_server/params/topo"
)

var (
	StartTopologyParamsInstance = &StartTopologyParams{}
	ModuleParamsLogger          = logger.GetLogger(logger.ModuleParams)
)

type StartTopologyParams struct {
	BlockChainParams     blockchain.BlockChainParams          `json:"block_chain_params"`
	PathValidationParams path_validation.PathValidationParams `json:"path_validation_params"`
	TopologyParams       topo.TopologyParams                  `json:"topology_params"`
	NetworkConfig        network.Params                       `json:"network_config"`
	ConstellationConfig  constellation.Params                 `json:"constellation_config"`
	ImagesConfig         images.Params                        `json:"images_config"`
	ServicesConfig       services.Params                      `json:"services_config"`
	AppsConfig           apps.Params                          `json:"apps_config"`
	ResourcesConfig      resources.Params                     `json:"resources_config"`
	RaspberryPiConfig    raspberrypi.Params                   `json:"raspberrypi_config"`
}

// ProcessStartTopologyParams applies the derived values that cannot be sent by
// chain_simulation, such as absolute paths, local addresses, and parsed times.
func ProcessStartTopologyParams(startTopologyParams *StartTopologyParams) error {
	if startTopologyParams == nil {
		return fmt.Errorf("start topology params must not be nil")
	}

	// constellation modification
	goStartTime, err := resolveStartTime(startTopologyParams.ConstellationConfig.StartTime)
	if err != nil {
		return err
	}
	startTopologyParams.ConstellationConfig.GoStartTime = goStartTime
	startTopologyParams.ConstellationConfig.ISLBandwidth *= 1e6

	// resolve blockchain type
	startTopologyParams.BlockChainParams.BasicBlockChainParmas.BlockchainType = blockchain.ResolveBlockChainType(
		startTopologyParams.BlockChainParams.BasicBlockChainParmas.BlockchainTypeStr,
	)

	// get validate ospf version
	if err = validateOspfVersion(startTopologyParams.NetworkConfig.OspfVersion); err != nil {
		return err
	}
	return nil
}

func resolveStartTime(value string) (time.Time, error) {
	result, err := time.ParseInLocation("2006|1|2|15|4|5", value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse constellation start time %q: %w", value, err)
	}
	return result, nil
}

func validateOspfVersion(version string) error {
	var availableOspfVersions = map[string]struct{}{
		"ospfv2": {},
		"ospfv3": {},
	}
	if _, ok := availableOspfVersions[version]; !ok {
		return fmt.Errorf("unsupported OSPF version: %s", version)
	}
	return nil
}
