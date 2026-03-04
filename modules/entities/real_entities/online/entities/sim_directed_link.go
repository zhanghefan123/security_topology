package entities

import "zhanghefan123/security_topology/modules/entities/types"

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

type SimDirectedNormalLink struct {
	*SimDirectedLinkBase
}

func NewSimDirectedNormalLink(source *SimAbstractNode, target *SimAbstractNode) *SimDirectedNormalLink {
	return &SimDirectedNormalLink{
		SimDirectedLinkBase: NewSimDirectedLinkBase(types.SimDirectedLinkType_SimDirectedNormalLink, source, target),
	}
}

type SimDirectedPvLink struct {
	*SimDirectedLinkBase
	Weights               []float64 // 随时间 t 变化
	IllegalRatios         []float64 // 随时间 t 变化
	ExploreProbability    float64   // 探索概率
	EstimatedIllegalRatio float64   // 估计的非法比例
	EstimatedLegalRatio   float64   // 估计的合法比例
	RectifiedGain         float64   // 修正增益
}

func NewSimDirectedPvLink(source *SimAbstractNode, target *SimAbstractNode) *SimDirectedPvLink {
	return &SimDirectedPvLink{
		SimDirectedLinkBase: NewSimDirectedLinkBase(types.SimDirectedLinkType_SimDirectedPvLink, source, target),
		Weights:             make([]float64, 0),
		IllegalRatios:       make([]float64, 0),
	}
}
