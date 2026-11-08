package network

type Params struct {
	BaseV4NetworkAddress string `json:"base_v4_network_address"`
	BaseV6NetworkAddress string `json:"base_v6_network_address"`
	ValidationListenPort int    `json:"validation_listen_port"`
	EnableFrr            bool   `json:"enable_frr"`
	OspfVersion          string `json:"ospf_version"`
	EnableSRv6           bool   `json:"enable_srv6"`
}
