package services

import (
	"zhanghefan123/security_topology/configs/services/delay"
	"zhanghefan123/security_topology/configs/services/etcd"
	"zhanghefan123/security_topology/configs/services/position"
	"zhanghefan123/security_topology/configs/services/web"
)

type ServicesConfig struct {
	EtcdConfig           etcd.EtcdConfig               `mapstructure:"etcd_config"`
	WebConfig            web.WebConfig                 `mapstructure:"web_config"`
	PositionUpdateConfig position.PositionUpdateConfig `mapstructure:"position_update_config"`
	DelayUpdateConfig    delay.DelayUpdateConfig       `mapstructure:"delay_update_config"`
}
