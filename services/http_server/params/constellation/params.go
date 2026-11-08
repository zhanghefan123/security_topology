package constellation

import "time"

type SatelliteParams struct {
	Type    int32 `json:"type"`
	P2PPort int   `json:"p2p_port"`
	RPCPort int   `json:"rpc_port"`
}

type Params struct {
	StartTime                string          `json:"start_time"`
	GoStartTime              time.Time       `json:"-"`
	SatelliteConfig          SatelliteParams `json:"satellite_config"`
	ISLBandwidth             int             `json:"isl_bandwidth"`
	GSLBandwidth             int             `json:"gsl_bandwidth"`
	SatelliteAvailableGSLs   int             `json:"satellite_available_gsls"`
	MinimumElevationAngle    int             `json:"minimum_elevation_angle"`
	TimeStepKey              string          `json:"time_step_key"`
	MinimumElevationAngleKey string          `json:"minimum_elevation_angle_key"`
	UpdatePosition           bool            `json:"update_position"`
}
