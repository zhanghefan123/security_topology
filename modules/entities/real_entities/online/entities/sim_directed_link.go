package entities

import "zhanghefan123/security_topology/modules/entities/real_entities/online/types"

type SimDirectedLinkBase struct {
	Type   types.SimDirectedLinkType
	Source *SimAbstractNode
	Target *SimAbstractNode
}

func (simDirectedLinkBase *SimDirectedLinkBase) GetSourceAndTargetNames() (string, string, error) {
	sourceName, err := simDirectedLinkBase.Source.GetSimNodeName()
	if err != nil {
		return "", "", err
	}
	targetName, err := simDirectedLinkBase.Target.GetSimNodeName()
	if err != nil {
		return "", "", err
	}
	return sourceName, targetName, nil
}

func NewSimDirectedLinkBase(linkType types.SimDirectedLinkType, source *SimAbstractNode, target *SimAbstractNode) *SimDirectedLinkBase {
	return &SimDirectedLinkBase{
		Type:   linkType,
		Source: source,
		Target: target,
	}
}

type SimDirectedRealLink struct {
	*SimDirectedLinkBase
}

func NewSimDirectedRealLink(source *SimAbstractNode, target *SimAbstractNode) *SimDirectedRealLink {
	return &SimDirectedRealLink{
		SimDirectedLinkBase: NewSimDirectedLinkBase(types.SimDirectedLinkType_SimDirectedNormalLink, source, target),
	}
}

type SimDirectedAbsLink struct {
	*SimDirectedLinkBase
	LinkType               types.SimDirectedLinkType // 链路的类型
	Intermediate           *SimAbstractNode          // pvlink 的中间节点
	Weights                []float64                 // 随时间 t 变化
	LegalRatios            []float64                 // 随时间 t 变化
	IllegalRatios          []float64                 // 非法率
	ExploreProbabilities   []float64                 // 探索概率
	RectifiedGains         []float64                 // 修正增益
	RectifiedLosses        []float64                 // 损失
	CurrentEdgeProbability float64                   // 为了 decomposition 而设置的一个临时变量
	Description            string                    // 对这个 directed pv link 的唯一描述
}

// NewSimDirectedAbsLink 进行新的抽象链路的创建
func NewSimDirectedAbsLink(linkType types.SimDirectedLinkType, pvLinkDescription string, source *SimAbstractNode, intermediate *SimAbstractNode, target *SimAbstractNode) *SimDirectedAbsLink {
	return &SimDirectedAbsLink{
		LinkType:             linkType,
		SimDirectedLinkBase:  NewSimDirectedLinkBase(types.SimDirectedLinkType_SimDirectedPvLink, source, target),
		Intermediate:         intermediate,
		Weights:              make([]float64, 0),
		LegalRatios:          make([]float64, 0),
		IllegalRatios:        make([]float64, 0),
		ExploreProbabilities: make([]float64, 0),
		RectifiedGains:       make([]float64, 0),
		RectifiedLosses:      make([]float64, 0),
		Description:          pvLinkDescription,
	}
}
