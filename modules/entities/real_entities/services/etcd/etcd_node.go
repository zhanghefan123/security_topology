package etcd

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// EtcdNode Etcd 节点
type EtcdNode struct {
	*normal_node.NormalNode
	ClientPort int
	PeerPort   int
	DataDir    string
	EtcdName   string
}

// NewEtcdNode 创建新的 etcd 节点
func NewEtcdNode(status types.NetworkNodeStatus, clientPort, peerPort int,
	dataDir, etcdName string) *EtcdNode {
	return &EtcdNode{
		NormalNode: &normal_node.NormalNode{
			Status: status,
		},
		ClientPort: clientPort,
		PeerPort:   peerPort,
		DataDir:    dataDir,
		EtcdName:   etcdName,
	}
}
