package types

func GetPrefix(typ NetworkNodeType) string {
	if typ == NetworkNodeType_NormalSatellite {
		return "ns"
	} else if typ == NetworkNodeType_ConsensusSatellite {
		return "cs"
	}
	return ""
}
