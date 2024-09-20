package lifecycle

import "zhanghefan123/security_topology/modules/entities/real_entities/constellation"

func Delete() {
	constellation.ConstellationInstance.Remove()
}
