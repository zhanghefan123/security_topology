package route

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
)

type StaticRoute struct {
	DestinationNetworkSegment string
	Gateway                   string
}

// GenerateStaticRoutes 进行静态路由的生成
func GenerateStaticRoutes(abstractNode *node.AbstractNode, linksMap *map[string]map[string]*link.AbstractLink, graphTmp *simple.DirectedGraph) ([]StaticRoute, error) {
	allStaticRoutes := make([]StaticRoute, 0)
	shortestPath := path.DijkstraFrom(abstractNode, graphTmp)
	nodeIterator := graphTmp.Nodes()
	for {
		hasNext := nodeIterator.Next()
		if !hasNext {
			break
		}
		currentDestination := nodeIterator.Node()
		if currentDestination.ID() != abstractNode.ID() {
			hopList, _ := shortestPath.To(currentDestination.ID())
			if len(hopList) == 2 {
				// 如果只有两个节点，说明是直接相连的
				continue
			}
			// 获取第一个链路和最后一个链路的两个端点
			// ------------------------------------------
			firstLinkSource := hopList[0]
			firstLinkTarget := hopList[1]
			firstLinkSourceNormal, err := GetNormalNodeFromGraphNode(firstLinkSource)
			if err != nil {
				return nil, fmt.Errorf("cannot get first link source normal node: %w", err)
			}
			firstLinkTargetNormal, err := GetNormalNodeFromGraphNode(firstLinkTarget)
			if err != nil {
				return nil, fmt.Errorf("cannot get first link target normal node: %w", err)
			}
			finalLinkSource := hopList[len(hopList)-2]
			finalLinkTarget := hopList[len(hopList)-1]
			finalLinkSourceNormal, err := GetNormalNodeFromGraphNode(finalLinkSource)
			if err != nil {
				return nil, fmt.Errorf("cannot get final link source normal node: %w", err)
			}
			finalLinkTargetNormal, err := GetNormalNodeFromGraphNode(finalLinkTarget)
			if err != nil {
				return nil, fmt.Errorf("cannot get final link target normal node: %w", err)
			}
			// ------------------------------------------

			// 根据节点名称进行链路的获取, 从而得到网关和目的网段
			// ------------------------------------------
			var gateway string
			var destinationNetworkSegment iplib.Net4
			firstLink := (*linksMap)[firstLinkSourceNormal.ContainerName][firstLinkTargetNormal.ContainerName]
			if firstLink == nil {
				firstLink = (*linksMap)[firstLinkTargetNormal.ContainerName][firstLinkSourceNormal.ContainerName]
				gateway = firstLink.SourceInterface.SourceIpv4Addr
			} else {
				gateway = firstLink.TargetInterface.SourceIpv4Addr
			}
			finalLink := (*linksMap)[finalLinkSourceNormal.ContainerName][finalLinkTargetNormal.ContainerName]
			if finalLink == nil {
				finalLink = (*linksMap)[finalLinkTargetNormal.ContainerName][finalLinkSourceNormal.ContainerName]
			}
			destinationNetworkSegment = finalLink.NetworkSegmentIPv4

			allStaticRoutes = append(allStaticRoutes, StaticRoute{
				DestinationNetworkSegment: destinationNetworkSegment.String(),
				Gateway:                   gateway,
			})
		}
	}
	return allStaticRoutes, nil
}
