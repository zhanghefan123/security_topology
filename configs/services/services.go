package services

import (
	"zhanghefan123/security_topology/configs/services/etcd"
	"zhanghefan123/security_topology/configs/services/position"
)

type ServicesConfig struct {
	EtcdConfig           etcd.EtcdConfig               `mapstructure:"etcd_config"`
	PositionUpdateConfig position.PositionUpdateConfig `mapstructure:"position_update_config"`
}
