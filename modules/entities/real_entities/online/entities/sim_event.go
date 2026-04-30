package entities

type SimEvent struct {
	StartEpoch    int                   // 开始生效的 epoch
	UpdateRouters []*UpdateNormalRouter // 更新的节点的信息
}

type UpdateNormalRouter struct {
	NormalRouterName         string  // 普通节点名称
	StartCorruptRatio        float64 // 起始破坏率
	EndCorruptRatio          float64 // 结束破坏率
	StartCorruptSpecialRatio float64 // 起始ack破坏率
	EndCorruptSpecialRatio   float64 // 结束ack破坏率
}
