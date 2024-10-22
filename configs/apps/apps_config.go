package apps

import "zhanghefan123/security_topology/configs/apps/ipv6"

type AppsConfig struct {
	IPv6Config ipv6.IPv6Config `mapstructure:"ipv6_config"`
}
