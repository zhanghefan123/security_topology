package satellites

import "zhanghefan123/security_topology/modules/entities/real_entities/normal_node"

type SatelliteInterface interface {
	normal_node.NormalNodeInterface
	GetOrbitId() int
	GetIndexInOrbit() int
}
