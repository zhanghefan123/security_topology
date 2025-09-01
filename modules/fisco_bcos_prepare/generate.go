package fisco_bcos_prepare

import (
	"fmt"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
	"zhanghefan123/security_topology/modules/utils/file"
)

const (
	ExecuteBuildChainSh = "ExecuteBuidlChainSh"
	ModifyNodesJson     = "ModifyNodesJson"
)

type GenerateFunction func() error

func (p *FiscoBcosPrepare) Generate() error {
	generateSteps := []map[string]GenerateFunction{
		{ExecuteBuildChainSh: p.ExecuteBuildChainSh},
		{ModifyNodesJson: p.ModifyNodesJson},
	}
	err := p.generatePrepareSteps(generateSteps)
	if err != nil {
		return fmt.Errorf("generate prepare failed: %w", err)
	} else {
		return nil
	}
}

// generateSteps 按步骤进行初始化
func (p *FiscoBcosPrepare) generatePrepareSteps(generateSteps []map[string]GenerateFunction) (err error) {
	fmt.Println()
	moduleNum := len(generateSteps)
	for idx, initStep := range generateSteps {
		for name, generateFunc := range initStep {
			if err = generateFunc(); err != nil {
				return fmt.Errorf("generate step [%s] failed, %w", name, err)
			}
			fiscoBcosPrepareWorkLogger.Infof("Generate STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
		}
	}
	fmt.Println()
	return
}

func (p *FiscoBcosPrepare) ExecuteBuildChainSh() error {
	if _, ok := p.generateSteps[ExecuteBuildChainSh]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already execute build_chain.sh")
		return nil
	}

	fiscoExamplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath
	p2pStartPort := configs.TopConfiguration.FiscoBcosConfig.P2pStartPort
	rpcStartPort := configs.TopConfiguration.FiscoBcosConfig.RpcStartPort

	err := dir.WithContextManager(fiscoExamplePath, func() error {
		commandArgs := []string{
			"build_chain.sh",
			"-D",
			"-l",
			fmt.Sprintf("127.0.0.1:%d", p.fiscoBcosNodeCount),
			"-p",
			fmt.Sprintf("%d,%d", p2pStartPort, rpcStartPort),
		}
		err := execute.Command("bash", commandArgs)
		if err != nil {
			return fmt.Errorf("failed to execute build_chain.sh")
		} else {
			return nil
		}
	})

	if err != nil {
		return fmt.Errorf("fail to execute build_chain.sh")
	}

	p.generateSteps[ExecuteBuildChainSh] = struct{}{}
	return nil
}

/*
ModifyNodesJson
生成案例 {"nodes":["127.0.0.1:30300","127.0.0.1:30301","127.0.0.1:30302","127.0.0.1:30303"]}
生成案例 {"nodes":["192.168.0.1/30:30300","192.168.0.6/30:30301","192.168.0.10/30:30302","192.168.0.14/30:30303"]}
*/
func (p *FiscoBcosPrepare) ModifyNodesJson() error {
	if _, ok := p.generateSteps[ModifyNodesJson]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already modify nodes json")
		return nil
	}

	// 生成每个 nodes.json 的内容
	// --------------------------------------------------------------------------------------
	finalString := ""
	finalString += "{\"nodes\":["
	for index, ipAddress := range p.firstIpAddresses {
		if index != len(p.firstIpAddresses)-1 {
			finalString += fmt.Sprintf("\"%s:%d\",", ipAddress[:len(ipAddress)-3], p.p2pStartPort+index)
		} else {
			finalString += fmt.Sprintf("\"%s:%d\"", ipAddress[:len(ipAddress)-3], p.p2pStartPort+index)
		}
	}
	finalString += "]}"
	// --------------------------------------------------------------------------------------

	fiscoExamplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath

	for index := range p.fiscoBcosNodeCount {
		nodesJsonFilePath := filepath.Join(fiscoExamplePath, fmt.Sprintf("nodes/127.0.0.1/node%d/nodes.json", index))
		err := file.WriteStringIntoFile(nodesJsonFilePath, finalString)
		if err != nil {
			return fmt.Errorf("modify nodes.json failed")
		}
	}

	p.generateSteps[ModifyNodesJson] = struct{}{}
	return nil
}

// 我们的实现的 netstat 情况
/*

root@aa15d0662797:/data# netstat -antp
Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 0.0.0.0:30300           0.0.0.0:*               LISTEN      1/fisco-bcos
tcp        0      0 0.0.0.0:20200           0.0.0.0:*               LISTEN      1/fisco-bcos
tcp        0      0 0.0.0.0:2623            0.0.0.0:*               LISTEN      40/mgmtd
tcp        0      0 127.0.0.1:2616          0.0.0.0:*               LISTEN      58/staticd
tcp        0      0 127.0.0.1:2608          0.0.0.0:*               LISTEN      55/isisd
tcp        0      0 127.0.0.1:2601          0.0.0.0:*               LISTEN      35/zebra
tcp        0      0 127.0.0.1:2604          0.0.0.0:*               LISTEN      49/ospfd
tcp        0      0 127.0.0.1:2605          0.0.0.0:*               LISTEN      42/bgpd
tcp        0      0 192.168.0.1:30300       192.168.0.10:39118      ESTABLISHED 1/fisco-bcos
tcp        0      0 192.168.0.1:37368       192.168.0.6:30301       ESTABLISHED 1/fisco-bcos
tcp        0      0 192.168.0.1:37448       192.168.0.14:30303      ESTABLISHED 1/fisco-bcos
tcp6       0      0 ::1:2606                :::*                    LISTEN      52/ospf6d
tcp6       0      0 :::2623                 :::*                    LISTEN      40/mgmtd

*/

// 官方实现的 netstat 情况
/*

root@zhf-virtual-machine:/data# netstat -atnp
Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 0.0.0.0:20202           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:20203           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:20200           0.0.0.0:*               LISTEN      1/fisco-bcos
tcp        0      0 0.0.0.0:20201           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:30302           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:30303           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:30300           0.0.0.0:*               LISTEN      1/fisco-bcos
tcp        0      0 0.0.0.0:30301           0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.53:53           0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.1:6012          0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.1:6010          0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.1:6011          0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.1:631           0.0.0.0:*               LISTEN      -
tcp        0      0 10.134.86.192:60302     34.107.243.93:443       ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:51096      ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:50767      ESTABLISHED -
tcp        0      0 127.0.0.1:56754         127.0.0.1:30300         ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:50245      ESTABLISHED -
tcp        0      0 127.0.0.1:44296         127.0.0.1:30302         ESTABLISHED -
tcp        0      0 10.134.86.192:40232     185.125.190.81:80       TIME_WAIT   -
tcp        0      0 10.134.86.192:22        10.134.86.15:50849      ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:50848      ESTABLISHED -
tcp        0      0 127.0.0.1:56744         127.0.0.1:30300         ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:49693      ESTABLISHED -
tcp        0      0 127.0.0.1:30302         127.0.0.1:44308         ESTABLISHED -
tcp        0      0 127.0.0.1:30300         127.0.0.1:56744         ESTABLISHED 1/fisco-bcos
tcp        0      0 127.0.0.1:30301         127.0.0.1:44016         ESTABLISHED -
tcp        0      0 10.134.86.192:22        10.134.86.15:49692      ESTABLISHED -
tcp        0      0 127.0.0.1:30302         127.0.0.1:44296         ESTABLISHED -
tcp        0      0 127.0.0.1:30300         127.0.0.1:56762         ESTABLISHED 1/fisco-bcos
tcp        0     96 10.134.86.192:22        10.134.86.15:50766      ESTABLISHED -
tcp        0      0 127.0.0.1:44308         127.0.0.1:30302         ESTABLISHED -
tcp        0      0 127.0.0.1:30300         127.0.0.1:56754         ESTABLISHED 1/fisco-bcos
tcp        0      0 127.0.0.1:56762         127.0.0.1:30300         ESTABLISHED -
tcp        0      0 127.0.0.1:44016         127.0.0.1:30301         ESTABLISHED -
tcp6       0      0 :::8080                 :::*                    LISTEN      -
tcp6       0      0 :::2375                 :::*                    LISTEN      -
tcp6       0      0 :::22                   :::*                    LISTEN      -
tcp6       0      0 ::1:6011                :::*                    LISTEN      -
tcp6       0      0 ::1:6010                :::*                    LISTEN      -
tcp6       0      0 ::1:6012                :::*                    LISTEN      -
tcp6       0      0 ::1:631                 :::*                    LISTEN      -

*/
