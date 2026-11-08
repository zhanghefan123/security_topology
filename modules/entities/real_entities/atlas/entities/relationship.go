package entities

import (
	"fmt"
	"strconv"
	"strings"
)

type Relationship struct {
	SplitNodeId        int
	ForwardingTargetId int
	NumberOfPaths      int
	PathIds            *[]int
}

func (r *Relationship) String() string {
	finalString := fmt.Sprintf("%d,%d,%d,", r.SplitNodeId, r.ForwardingTargetId, r.NumberOfPaths)
	for pathIndex, pathId := range *r.PathIds {
		if pathIndex != (len(*r.PathIds) - 1) {
			finalString += fmt.Sprintf("%d,", pathId)
		} else {
			finalString += fmt.Sprintf("%d", pathId)
		}
	}
	return finalString
}

func (r *Relationship) BeautifulPrint() string {
	if r == nil {
		return "<nil Relationship>"
	}
	if r.PathIds == nil {
		return fmt.Sprintf("Split Node: %d  →  Next Hop: %d  │  Paths (0): []",
			r.SplitNodeId, r.ForwardingTargetId)
	}
	ids := make([]string, 0, len(*r.PathIds))
	for _, id := range *r.PathIds {
		ids = append(ids, strconv.Itoa(id))
	}
	return fmt.Sprintf("Split Node: %d  →  Next Hop: %d  │  Paths (%d): [%s]",
		r.SplitNodeId, r.ForwardingTargetId, len(*r.PathIds), strings.Join(ids, ", "))
}
