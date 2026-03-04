package test

import (
	"testing"
	"zhanghefan123/security_topology/modules/entities/real_entities/graph/entities"
)

func TestJudge(t *testing.T) {
	segment := &entities.Segment{
		Path: make([]*entities.Node, 0),
	}
	segment.Path = append(segment.Path, &entities.Node{NodeName: "4"})
	segment.Path = append(segment.Path, &entities.Node{NodeName: "5"})
	segment.Path = append(segment.Path, &entities.Node{NodeName: "6"})
	path := &entities.Path{
		NodeList: make([]*entities.Node, 0),
	}
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "1"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "2"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "4"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "3"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "5"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "6"})
	path.NodeList = append(path.NodeList, &entities.Node{NodeName: "7"})
}

func TestFilterPathsAccordingToDirectLink(t *testing.T) {
	allPaths := make([]*entities.Path, 0)
	path1 := &entities.Path{
		NodeList: make([]*entities.Node, 0),
	}
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "1"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "2"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "3"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "5"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "4"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "6"})
	path1.NodeList = append(path1.NodeList, &entities.Node{NodeName: "7"})

	path2 := &entities.Path{
		NodeList: make([]*entities.Node, 0),
	}
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "1"})
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "2"})
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "4"})
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "5"})
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "6"})
	path2.NodeList = append(path2.NodeList, &entities.Node{NodeName: "7"})

	path3 := &entities.Path{
		NodeList: make([]*entities.Node, 0),
	}
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "1"})
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "2"})
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "3"})
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "5"})
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "6"})
	path3.NodeList = append(path3.NodeList, &entities.Node{NodeName: "7"})

	path4 := &entities.Path{
		NodeList: make([]*entities.Node, 0),
	}
	path4.NodeList = append(path4.NodeList, &entities.Node{NodeName: "1"})
	path4.NodeList = append(path4.NodeList, &entities.Node{NodeName: "2"})
	path4.NodeList = append(path4.NodeList, &entities.Node{NodeName: "4"})
	path4.NodeList = append(path4.NodeList, &entities.Node{NodeName: "6"})
	path4.NodeList = append(path4.NodeList, &entities.Node{NodeName: "7"})

	allPaths = append(allPaths, path1)
	allPaths = append(allPaths, path2)
	allPaths = append(allPaths, path3)
	allPaths = append(allPaths, path4)
}
