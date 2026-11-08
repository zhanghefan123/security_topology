package raspberrypi

type Params struct {
	NodeIDs     []int    `json:"node_ids"`
	NodeTypes   []string `json:"node_types"`
	IpAddresses []string `json:"ip_addresses"`
	Connections []string `json:"connections"`
	UserName    string   `json:"user_name"`
	Password    string   `json:"password"`
	GrpcPort    int      `json:"grpc_port"`
}
