package apps

type IPv6Config struct {
	ServerPort int `json:"server_port"`
}

type Params struct {
	IPv6Config IPv6Config `json:"ipv6_config"`
}
