package entities

type SimEvent struct {
	StartEpoch    int                   // 开始生效的 epoch
	UpdateRouters []*UpdateNormalRouter // 更新的节点的信息
}

type UpdateNormalRouter struct {
	NormalRouterName  string  // 普通节点名称
	StartDropRatio    float64 // 起始丢包率
	EndDropRatio      float64 // 结束丢包率
	StartIllegalRatio float64 // 非法率起始值
	EndIllegalRatio   float64 // 非法率终止值
}
