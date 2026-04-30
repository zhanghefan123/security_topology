package nodes

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
)

type PacketProcessor interface {
	ProcessPacket(packet *entities.SimPacket)
}
