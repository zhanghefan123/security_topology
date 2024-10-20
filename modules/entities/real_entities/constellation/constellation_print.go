package constellation

import (
	"github.com/kr/pretty"
	"zhanghefan123/security_topology/modules/entities/types"
)

type PrintType int32

const (
	PrintType_Satellites PrintType = 0
	PrintType_Links      PrintType = 1
)

// Print 进行星座之中的卫星或者链路的打印
func (c *Constellation) Print(printType PrintType) {
	if printType == PrintType_Satellites {
		if c.SatelliteType == types.NetworkNodeType_NormalSatellite {
			for _, satellite := range c.NormalSatellites {
				prettyOutput := pretty.Sprint(satellite)
				constellationLogger.Infof(prettyOutput)
			}
		} else if c.SatelliteType == types.NetworkNodeType_ConsensusSatellite {
			for _, satellite := range c.ConsensusSatellites {
				prettyOutput := pretty.Sprint(satellite)
				constellationLogger.Infof(prettyOutput)
			}
		}

	} else if printType == PrintType_Links {
		allLinks := append(c.AllSatelliteLinks)
		for _, link := range allLinks {
			prettyOutput := pretty.Sprint(link)
			constellationLogger.Infof(prettyOutput)
		}
	}
}
