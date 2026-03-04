package entities

import (
	"fmt"
)

type ValidationField struct {
	ValidationStr      string
	ValidationNodeName string
}

type HeaderList struct {
	SourceNodeName      string             // 这一个list的源的名字
	Depth               int                // 深度
	StartTag            int                // 起始标签
	PVF                 string             // pvf
	ValidationFieldList []*ValidationField // 验证字段的列表
	EndTag              int                // 结束标签
}

func HeaderListToString(headerList *HeaderList) string {
	start := fmt.Sprintf("source: %s, depth: %d, start tag: %d[ ", headerList.SourceNodeName, headerList.Depth, headerList.StartTag)
	middlePart := ""
	for index, validationField := range headerList.ValidationFieldList {
		if index != (len(headerList.ValidationFieldList) - 1) {
			middlePart += validationField.ValidationStr + ","
		} else {
			middlePart += validationField.ValidationStr
		}
	}
	end := fmt.Sprintf(" end tag: %d]", headerList.EndTag)
	return start + middlePart + end
}

func CreateHeaderListFromSegment(segment *Segment) *HeaderList {
	// 进行 pvf 的计算
	pvf := fmt.Sprintf("[Path_%d[PVF]]", segment.Id)
	var validationFieldList []*ValidationField
	// 进行 opv 的计算
	for index, node := range segment.Path {
		if index != 0 {
			opv := fmt.Sprintf("Path_%d[OPV_%s]", segment.Id, node.NodeName)
			validationFieldList = append(validationFieldList, &ValidationField{
				ValidationStr:      opv,
				ValidationNodeName: node.NodeName,
			})
		}
	}
	return &HeaderList{
		SourceNodeName:      segment.Path[0].NodeName,
		Depth:               segment.Depth,
		PVF:                 pvf,
		ValidationFieldList: validationFieldList,
		StartTag:            segment.Id,
		EndTag:              segment.Id,
	}
}
