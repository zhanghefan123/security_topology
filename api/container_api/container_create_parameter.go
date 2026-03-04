package container_api

import "zhanghefan123/security_topology/modules/entities/abstract_entities/node"

type ContainerCreateParameter struct {
	NumberOfNodes int
	AbstractNode  *node.AbstractNode
}
