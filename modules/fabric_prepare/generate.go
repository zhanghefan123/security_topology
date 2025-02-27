package fabric_prepare

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/configs"
)

const (
	InitializePathMap             = "InitializePathMap"             // 第一步: 进行 path mapping 的初始化
	GenerateGenesisBlockConfigYml = "GenerateGenesisBlockConfigYml" // 第二步: 创建创世区块 yml 文件
	GenerateOrdererOrgCryptoYml   = "GenerateOrdererOrgCryptoYml"   // 第三步: 创建 fabric 排序节点 yml 文件
	GeneratePeerOrgCryptoYml      = "GeneratePeerOrgCryptoYml"      // 第四步: 创建 fabric peer 节点 yml 文件
	InvokeCryptogenTool           = "InvokeCryptogenTool"           // 第五步: 调用加密工具
)

const (
	FabricBin           = "FabricBin"
	FabricBinCryptogen  = "FabricBinCryptogen"
	GenesisBlockBasic   = "GenesisBlockBasic"
	GenesisBlockNew     = "GenesisBlockNew"
	OrdererOrgCryptoNew = "OrdererOrgCryptoNew"
	PeerOrgCryptoNew    = "PeerOrgCryptoNew"
	organizationsPath   = "organizationsPath"
)

type GenerateFunction func() error

func (p *FabricPrepare) Generate() error {
	generateSteps := []map[string]GenerateFunction{
		{InitializePathMap: p.InitializePathMap},
		{GenerateGenesisBlockConfigYml: p.GenerateGenesisBlockConfigYml},
		{GenerateOrdererOrgCryptoYml: p.GenerateOrdererOrgCryptoYml},
		{GeneratePeerOrgCryptoYml: p.GeneratePeerOrgCryptoYml},
		{InvokeCryptogenTool: p.InvokeCryptogenTool},
	}
	err := p.generatePrepareSteps(generateSteps)
	if err != nil {
		return fmt.Errorf("generate prepare failed %w", err)
	}
	return nil
}

// InitializePathMap 进行 path mapping 的初始化
func (p *FabricPrepare) InitializePathMap() error {
	if _, ok := p.generateSteps[InitializePathMap]; ok {
		fabricPrepareWorkLogger.Infof("already initialize path mapping")
		return nil
	}

	fabricConfig := configs.TopConfiguration.FabricConfig
	fabricSamplesPath := fabricConfig.FabricSamplesPath
	fabricNetworkPath := fabricConfig.FabricNetworkPath

	p.pathMapping[FabricBin] = path.Join(fabricSamplesPath, "bin")
	p.pathMapping[FabricBinCryptogen] = path.Join(fabricSamplesPath, "bin/cryptogen")
	p.pathMapping[GenesisBlockBasic] = path.Join(fabricNetworkPath, "bft_config/configtx_basic.yaml")
	p.pathMapping[GenesisBlockNew] = path.Join(fabricNetworkPath, "bft_config/configtx.yaml")
	p.pathMapping[OrdererOrgCryptoNew] = path.Join(fabricNetworkPath, "organizations/cryptogen/crypto-config-orderer.yaml")
	p.pathMapping[PeerOrgCryptoNew] = path.Join(fabricNetworkPath, "organizations/cryptogen")
	p.pathMapping[organizationsPath] = path.Join(fabricNetworkPath, "organizations")

	fabricPrepareWorkLogger.Infof("successfully initialize path mapping")

	p.generateSteps[InitializePathMap] = struct{}{}
	return nil
}

// generateSteps 按步骤进行初始化
func (p *FabricPrepare) generatePrepareSteps(generateSteps []map[string]GenerateFunction) (err error) {
	fmt.Println()
	moduleNum := len(generateSteps)
	for idx, initStep := range generateSteps {
		for name, generateFunc := range initStep {
			if err = generateFunc(); err != nil {
				return fmt.Errorf("generate step [%s] failed, %w", name, err)
			}
			fabricPrepareWorkLogger.Infof("Generate STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
		}
	}
	fmt.Println()
	return
}

// GenerateGenesisBlockConfigYml  创建创世区块 yml 文件
func (p *FabricPrepare) GenerateGenesisBlockConfigYml() (err error) {
	if _, ok := p.generateSteps[GenerateGenesisBlockConfigYml]; ok {
		fabricPrepareWorkLogger.Errorf("already generate genesis block config file")
		return nil
	}

	inputFile := p.pathMapping[GenesisBlockBasic] // 模版文件路径
	outputFile := p.pathMapping[GenesisBlockNew]  // 输出文件路径
	var file, newFile *os.File                    // 模版文件和输出文件句柄
	ordererNum := p.fabricOrderNodeCount
	peerNum := p.fabricPeerNodeCount
	orderStartPort := configs.TopConfiguration.FabricConfig.OrderStartPort

	// 打开输入文件
	// -------------------------------------------------------
	file, err = os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		closeFileErr := file.Close()
		if err != nil {
			err = closeFileErr
		}
	}(file)
	// -------------------------------------------------------

	// 创建输出文件
	// -------------------------------------------------------
	newFile, err = os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func(file *os.File) {
		closeFileErr := file.Close()
		if err != nil {
			err = closeFileErr
		}
	}(newFile)
	err = os.Chmod(outputFile, 0777)
	if err != nil {
		return fmt.Errorf("failed to open Permission: %v", err)
	}
	// -------------------------------------------------------

	// 使用 bufio.Scanner 逐行读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是 "OrdererEndpoints:" 行
		if strings.Contains(line, "need_to_fill_1") {
			for i := 1; i <= ordererNum; i++ {
				_, _ = fmt.Fprintf(newFile, "      - orderer%d.example.com:%d\n", i, orderStartPort+i)
			}
		} else if strings.Contains(line, "need_to_fill_2") {
			for i := 1; i <= peerNum; i++ {
				_, _ = fmt.Fprintf(newFile, "  - &Org%d\n", i)
				_, _ = fmt.Fprintf(newFile, "    Name: Org%dMSP\n", i)
				_, _ = fmt.Fprintf(newFile, "    ID: Org%dMSP\n", i)
				_, _ = fmt.Fprintf(newFile, "    MSPDir: ../organizations/peerOrganizations/org%d.example.com/msp\n", i)
				_, _ = fmt.Fprintf(newFile, "    Policies:\n")
				_, _ = fmt.Fprintf(newFile, "      Readers:\n")
				_, _ = fmt.Fprintf(newFile, "        Type: Signature\n")
				_, _ = fmt.Fprintf(newFile, "        Rule: \"OR('Org%dMSP.admin', 'Org%dMSP.peer', 'Org%dMSP.client')\"\n", i, i, i)
				_, _ = fmt.Fprintf(newFile, "      Writers:\n")
				_, _ = fmt.Fprintf(newFile, "        Type: Signature\n")
				_, _ = fmt.Fprintf(newFile, "        Rule: \"OR('Org%dMSP.admin', 'Org%dMSP.client')\"\n", i, i)
				_, _ = fmt.Fprintf(newFile, "      Admins:\n")
				_, _ = fmt.Fprintf(newFile, "        Type: Signature\n")
				_, _ = fmt.Fprintf(newFile, "        Rule: \"OR('Org%dMSP.admin')\"\n", i)
				_, _ = fmt.Fprintf(newFile, "      Endorsement:\n")
				_, _ = fmt.Fprintf(newFile, "        Type: Signature\n")
				_, _ = fmt.Fprintf(newFile, "        Rule: \"OR('Org%dMSP.peer')\"\n", i)
			}
		} else if strings.Contains(line, "need_to_fill_3") {
			for i := 1; i <= ordererNum; i++ {
				_, _ = fmt.Fprintf(newFile, "        - ID: %d\n", i)
				_, _ = fmt.Fprintf(newFile, "          Host: orderer%d.example.com\n", i)
				_, _ = fmt.Fprintf(newFile, "          Port: %d\n", orderStartPort+i)
				_, _ = fmt.Fprintf(newFile, "          MSPID: OrdererMSP\n")
				_, _ = fmt.Fprintf(newFile, "          Identity: ../organizations/ordererOrganizations/example.com/orderers/orderer%d.example.com/msp/signcerts/orderer%d.example.com-cert.pem\n", i, i)
				_, _ = fmt.Fprintf(newFile, "          ClientTLSCert: ../organizations/ordererOrganizations/example.com/orderers/orderer%d.example.com/tls/server.crt\n", i)
				_, _ = fmt.Fprintf(newFile, "          ServerTLSCert: ../organizations/ordererOrganizations/example.com/orderers/orderer%d.example.com/tls/server.crt\n", i)
			}
		} else if strings.Contains(line, "need_to_fill_4") {
			for i := 1; i <= peerNum; i++ {
				_, _ = fmt.Fprintf(newFile, "        - *Org%d\n", i)
			}
		} else {
			_, err = fmt.Fprintln(newFile, line)
			if err != nil {
				return fmt.Errorf("failed to write line to new file: %v", err)
			}
		}
	}

	// 检查扫描时是否出错
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 提示成功
	fabricPrepareWorkLogger.Infof("Successfully updated the YAML and saved to %s\n", outputFile)
	return nil
}

// GenerateOrdererOrgCryptoYml 为排序节点进行 yml 的生成
func (p *FabricPrepare) GenerateOrdererOrgCryptoYml() (err error) {
	if _, ok := p.generateSteps[GenerateOrdererOrgCryptoYml]; ok {
		fabricPrepareWorkLogger.Errorf("already generate orderer organization cryptogen file")
		return nil
	}

	outputFile := p.pathMapping[OrdererOrgCryptoNew]
	var file *os.File
	ordererNum := p.fabricOrderNodeCount
	if ordererNum != 0 {
		// 进行文件的创建和权限修改
		// --------------------------------------------------------
		file, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer func(file *os.File) {
			fileCloseErr := file.Close()
			if err != nil {
				err = fileCloseErr
			}
		}(file)
		err = os.Chmod(outputFile, 0777)
		if err != nil {
			return fmt.Errorf("failed to open Permission: %v", err)
		}
		// --------------------------------------------------------
		_, _ = fmt.Fprintf(file, "OrdererOrgs:\n")
		_, _ = fmt.Fprintf(file, "  - Name: Orderer\n")
		_, _ = fmt.Fprintf(file, "    Domain: example.com\n")
		_, _ = fmt.Fprintf(file, "    EnableNodeOUs: true\n")
		_, _ = fmt.Fprintf(file, "    Specs:\n")
		for i := 1; i <= ordererNum; i++ {
			_, _ = fmt.Fprintf(file, "      - Hostname: orderer%d\n", i)
			_, _ = fmt.Fprintf(file, "        SANS:\n")
			_, _ = fmt.Fprintf(file, "          - localhost\n")
			// _,_ = fmt.Fprintf(file, "          - %s\n", p.ipv4AddressesOrderer[i-1])
		}
		fabricPrepareWorkLogger.Infof("Successfully updated the YAML and saved to %s\n", outputFile)
	} else {
		fabricPrepareWorkLogger.Infof("No YAML need to be generated for fabric order nodes")
	}
	return nil
}

// GeneratePeerOrgCryptoYml 为对等节点进行 yml 的生成
func (p *FabricPrepare) GeneratePeerOrgCryptoYml() (err error) {
	if _, ok := p.generateSteps[GeneratePeerOrgCryptoYml]; ok {
		fabricPrepareWorkLogger.Errorf("already generate peer organization cryptogen file")
		return nil
	}

	peerNum := p.fabricPeerNodeCount
	for i := 1; i <= peerNum; i++ {
		// 文件创建
		// --------------------------------------------------------
		var file *os.File
		outputFile := path.Join(p.pathMapping[PeerOrgCryptoNew], fmt.Sprintf("crypto-config-org%d.yaml", i))
		file, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		err = os.Chmod(outputFile, 0777)
		if err != nil {
			return fmt.Errorf("failed to open Permission: %v", err)
		}
		// --------------------------------------------------------
		_, _ = fmt.Fprintf(file, "PeerOrgs:\n")
		_, _ = fmt.Fprintf(file, "  - Name: Org%d\n", i)
		_, _ = fmt.Fprintf(file, "    Domain: org%d.example.com\n", i)
		_, _ = fmt.Fprintf(file, "    EnableNodeOUs: true\n")
		_, _ = fmt.Fprintf(file, "    Template:\n")
		_, _ = fmt.Fprintf(file, "      Count: 1\n")
		_, _ = fmt.Fprintf(file, "      SANS:\n")
		_, _ = fmt.Fprintf(file, "        - localhost\n")
		// _,_ = fmt.Fprintf(file, "        - %s\n", p.ipv4AddressesPeer[i-1])
		_, _ = fmt.Fprintf(file, "    Users:\n")
		_, _ = fmt.Fprintf(file, "      Count: 1\n")
		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
		fabricPrepareWorkLogger.Infof("Successfully updated the YAML and saved to %s\n", outputFile)
	}

	return nil
}

// InvokeCryptogenTool 调用 cryptogen 工具
func (p *FabricPrepare) InvokeCryptogenTool() error {
	if _, ok := p.generateSteps[InvokeCryptogenTool]; ok {
		fabricPrepareWorkLogger.Errorf("already invoke crytogen tool")
		return nil
	}

	ordererNum := p.fabricOrderNodeCount
	peerNum := p.fabricPeerNodeCount
	configFile := p.pathMapping[OrdererOrgCryptoNew]
	if ordererNum != 0 {
		cmd := exec.Command(p.pathMapping[FabricBinCryptogen], "generate", "--config", configFile, "--output", p.pathMapping[organizationsPath])
		_, err := cmd.CombinedOutput()
		if err != nil {
			fabricPrepareWorkLogger.Errorf("can not generate orderer msp")
		}
	}
	fabricPrepareWorkLogger.Infof("Successfully generate orderer msp")

	for i := 1; i <= peerNum; i++ {
		configFile = path.Join(p.pathMapping[PeerOrgCryptoNew], fmt.Sprintf("crypto-config-org%d.yaml", i))
		cmd := exec.Command(p.pathMapping[FabricBinCryptogen], "generate", "--config", configFile, "--output", p.pathMapping[organizationsPath])
		_, err := cmd.CombinedOutput()
		if err != nil {
			fabricPrepareWorkLogger.Errorf("can not generate peer msp")
		}
	}

	fmt.Printf("p.pathMapping[organizationsPath]:%s", p.pathMapping[organizationsPath])

	err := setPermissions(p.pathMapping[organizationsPath], 0777)
	if err != nil {
		fabricPrepareWorkLogger.Errorf("Error changing permissions: %v", err)
	}
	return nil
}

// setPermissions 进行权限的修改
func setPermissions(path string, mode os.FileMode) error {
	err := os.Chmod(path, mode)
	if err != nil {
		return err
	}

	var fi os.FileInfo

	// 如果是目录，递归修改目录下的所有文件和子目录
	if fi, err = os.Stat(path); err == nil && fi.IsDir() {
		err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if err = os.Chmod(path, mode); err != nil {
				return err
			}
			return nil
		})
		return err
	}
	return nil
}
