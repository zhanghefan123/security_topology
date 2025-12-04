package fisco_bcos_prepare

import (
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/otiai10/copy"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/utils/dir"
	"zhanghefan123/security_topology/utils/execute"
	"zhanghefan123/security_topology/utils/file"
)

const (
	CopyBuildChainShTemplate = "CopyBuildChainShTemplate"
	ModifyBuildChainSh       = "ModifyBuildChainSh"
	ExecuteBuildChainSh      = "ExecuteBuidlChainSh"
	ModifyNodesJson          = "ModifyNodesJson"
	ConsolePrepareWork       = "ConsolePrepareWork"
)

type GenerateFunction func() error

func (p *FiscoBcosPrepare) Generate() error {
	generateSteps := []map[string]GenerateFunction{
		{CopyBuildChainShTemplate: p.CopyBuildChainShTemplate},
		{ModifyBuildChainSh: p.ModifyBuildChainSh},
		{ExecuteBuildChainSh: p.ExecuteBuildChainSh},
		{ModifyNodesJson: p.ModifyNodesJson},
		{ConsolePrepareWork: p.ConsolePrepareWork},
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

func (p *FiscoBcosPrepare) CopyBuildChainShTemplate() error {
	if _, ok := p.generateSteps[CopyBuildChainShTemplate]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already copy build chain sh template")
		return nil
	}

	sourceFile := filepath.Join(configs.TopConfiguration.PathConfig.ResourcesPath, "FISCO_BCOS_EXAMPLE/build_chain.sh")
	targetFile := filepath.Join(configs.TopConfiguration.FiscoBcosConfig.ExamplePath, "build_chain.sh")
	_, err := fileutils.CopyFile(sourceFile, targetFile)
	if err != nil {
		return fmt.Errorf("error copy build chain sh template file: %v", err)
	}

	p.generateSteps[CopyBuildChainShTemplate] = struct{}{}
	return nil
}

func (p *FiscoBcosPrepare) ModifyBuildChainSh() error {
	if _, ok := p.generateSteps[ModifyBuildChainSh]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already execute modify build chain sh")
		return nil
	}

	replaceMap := map[string]string{
		"{log_level}":            configs.TopConfiguration.FiscoBcosConfig.LogLevel,
		"{block_tx_count_limit}": fmt.Sprintf("%d", configs.TopConfiguration.FiscoBcosConfig.BlockTxCountLimit),
	}

	fmt.Printf("block_tx_count_limit == %d\n", configs.TopConfiguration.FiscoBcosConfig.BlockTxCountLimit)

	sourceFile := filepath.Join(configs.TopConfiguration.FiscoBcosConfig.ExamplePath, "build_chain.sh")
	targetFile := sourceFile
	err := file.CopyAndReplaceTemplate(sourceFile, targetFile, replaceMap)
	if err != nil {
		return fmt.Errorf("cannot copy and replace file due to: %v", err)
	}

	p.generateSteps[ModifyBuildChainSh] = struct{}{}
	return nil
}

// 被 ModifyBuildChainSh 替代了
// ------------------------------------------------------------------------------------------------------
/*
func (p *FiscoBcosPrepare) ChangeLogLevel() error {
	if _, ok := p.generateSteps[ChangeLogLevel]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already execute change log level")
		return nil
	}

	// buildSh 文件
	buildShFile := filepath.Join(configs.TopConfiguration.FiscoBcosConfig.ExamplePath, "build_chain.sh")

	var finalString = ""
	// 获取文件句柄
	{
		var findFirstLogLevel = false
		fileHandle, err := os.Open(buildShFile)
		if err != nil {
			return fmt.Errorf("error read buildShFile due to %v", err)
		}
		// 循环进行文件读取
		scanner := bufio.NewScanner(fileHandle)
		for scanner.Scan() {
			line := scanner.Text()
			if !findFirstLogLevel && strings.Contains(line, "log_level=") {
				line = fmt.Sprintf(`log_level="%s"`, configs.TopConfiguration.FiscoBcosConfig.LogLevel)
				findFirstLogLevel = true
			}
			finalString += fmt.Sprintf("%s\n", line)
		}
		err = fileHandle.Close()
		if err != nil {
			return fmt.Errorf("close file handle error %v", err)
		}
	}

	err := file.WriteStringIntoFile(buildShFile, finalString)
	if err != nil {
		return fmt.Errorf("write string into file error: %v", err)
	}

	p.generateSteps[ChangeLogLevel] = struct{}{}
	return nil
}
*/
// ------------------------------------------------------------------------------------------------------

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

func (p *FiscoBcosPrepare) ConsolePrepareWork() error {
	if _, ok := p.generateSteps[ConsolePrepareWork]; ok {
		fiscoBcosPrepareWorkLogger.Infof("already prepare console")
		return nil
	}

	examplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath
	consolePath := configs.TopConfiguration.FiscoBcosConfig.ConsolePath
	configExampleFile := filepath.Join(consolePath, "conf/config-example.toml")
	configDestinationFile := filepath.Join(consolePath, "conf/config.toml")
	err := copy.Copy(configExampleFile, configDestinationFile)
	if err != nil {
		return fmt.Errorf("copy config-example.toml --> config.toml failed, due to err: %v", err)
	}

	sdkCertsPath := filepath.Join(examplePath, "nodes/127.0.0.1/sdk/")
	sdkDestinationPath := filepath.Join(consolePath, "conf/")
	entries, err := os.ReadDir(sdkCertsPath)
	if err != nil {
		return fmt.Errorf("could not read sdkCertsPath")
	}
	for _, entry := range entries {
		fileSrcPath := filepath.Join(sdkCertsPath, entry.Name())
		fileDestPath := filepath.Join(sdkDestinationPath, entry.Name())
		err = copy.Copy(fileSrcPath, fileDestPath)
		if err != nil {
			return fmt.Errorf("copy sdkCertsPath --> console/conf failed, due to err: %v", err)
		}
	}

	p.generateSteps[ConsolePrepareWork] = struct{}{}
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
