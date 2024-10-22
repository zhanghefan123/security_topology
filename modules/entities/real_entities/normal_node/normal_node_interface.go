package normal_node

type NormalNodeInterface interface {
	GetId() int
	AppendIpv4Subnet(ipv4Subnet string)
	AppendIpv6Subnet(ipv6Subnet string)
}
