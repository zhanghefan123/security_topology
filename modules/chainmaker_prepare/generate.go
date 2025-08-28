package chainmaker_prepare

import (
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/file"
)

const (
	monitorPort             = 14321
	pprofPort               = 24321
	trustedPort             = 13301
	enableVmGo              = "false"
	vmGoContainerNamePrefix = "chainmaker-vm-go-container"
)

const (
	InitializePathMap     = "InitializePathMap"
	GenerateCertFiles     = "GenerateCertFiles"
	ResolvePeerIds        = "ResolvePeerIds"
	GenerateLogYml        = "GenerateLogYml"
	GenerateBcTemplate    = "GenerateBcTemplate"
	GenerateBcYml         = "GenerateBcYml"
	ModifyBcYml           = "ModifyBcYml"
	GenerateChainMakerYml = "GenerateChainMakerYml"
	CopyPrepareFiles      = "CopyPrepareFiles"
)

const (
	Resources               = "Resources"
	ChainMakerBuild         = "ChainMakerBuild"
	Templates               = "Templates"
	BuildConfig             = "BuildConfig"
	BuildCryptoConfig       = "BuildCryptoConfig"
	TestData                = "TestData"
	MultiNode               = "MultiNode"
	ChainMakerCryptoGenBin  = "ChainMakerCryptoGenBin"
	CryptoConfigOldFileName = "CryptoConfigOldFileName"
	CryptoConfigNewFileName = "CryptoConfigNewFileName"
	PKCS11OldFileName       = "PKCS11OldFileName"
	PKCS11NewFileName       = "PKCS11NewFileName"
	FrontPartOfBCFile       = "FrontPartOfBCFile"
	BackPartOfBCFile        = "BackPartOfBCFile"
	CmdTestData             = "CmdTestData"
)

type GenerateFunction func() error

func (p *ChainMakerPrepare) Generate() error {
	generateSteps := []map[string]GenerateFunction{
		{InitializePathMap: p.InitializePathMapping},
		{GenerateCertFiles: p.GenerateCertsFiles},
		{ResolvePeerIds: p.ResolvePeerIds},
		{GenerateLogYml: p.GenerateLogYml},
		{GenerateBcTemplate: p.GenerateBcTemplate},
		{GenerateBcYml: p.GenerateBcYml},
		{ModifyBcYml: p.ModifyBcYml},
		{GenerateChainMakerYml: p.GenerateChainMakerYml},
		{CopyPrepareFiles: p.CopyPrepareFiles},
	}
	err := p.generatePrepareSteps(generateSteps)
	if err != nil {
		return fmt.Errorf("generate prepare failed %w", err)
	}
	return nil
}

func (p *ChainMakerPrepare) InitializePathMapping() error {
	if _, ok := p.generateSteps[InitializePathMap]; ok {
		chainmakerPrepareWorkLogger.Infof("already initialize path mapping")
		return nil
	}

	resources := configs.TopConfiguration.PathConfig.ResourcesPath
	chainMakerConfig := configs.TopConfiguration.ChainMakerConfig
	chainMakerBuildPath := chainMakerConfig.ChainMakerBuild
	templatesFilePath := chainMakerConfig.TemplatesFilePath
	chainMakerGoProjectPath := chainMakerConfig.ChainMakerGoProjectPath
	chainMakerCryptoGenProjectPath := chainMakerConfig.CryptoGenProjectPath
	p.pathMapping[Resources] = configs.TopConfiguration.PathConfig.ResourcesPath
	p.pathMapping[ChainMakerBuild] = chainMakerBuildPath
	p.pathMapping[Templates] = templatesFilePath
	p.pathMapping[BuildConfig] = filepath.Join(chainMakerBuildPath, "config")
	p.pathMapping[BuildCryptoConfig] = filepath.Join(chainMakerBuildPath, "crypto-config")
	p.pathMapping[TestData] = filepath.Join(chainMakerGoProjectPath, "tools/cmc/testdata")
	p.pathMapping[MultiNode] = filepath.Join(chainMakerGoProjectPath, "scripts/docker/multi_node")
	p.pathMapping[ChainMakerCryptoGenBin] = filepath.Join(chainMakerCryptoGenProjectPath, "bin/chainmaker-cryptogen")
	p.pathMapping[CryptoConfigOldFileName] = "crypto_config_template.yml"
	p.pathMapping[CryptoConfigNewFileName] = "./crypto_config.yml"
	p.pathMapping[PKCS11OldFileName] = "pkcs11_keys.yml"
	p.pathMapping[PKCS11NewFileName] = "./pkcs11_keys.yml"
	p.pathMapping[FrontPartOfBCFile] = filepath.Join(resources, "bc_template_part/front_part_of_bc_file.txt")
	p.pathMapping[BackPartOfBCFile] = filepath.Join(resources, "bc_template_part/back_part_of_bc_file.txt")
	p.pathMapping[CmdTestData] = filepath.Join("../cmd", "testdata/")

	chainmakerPrepareWorkLogger.Infof("successfully initialize path mapping")

	p.generateSteps[InitializePathMap] = struct{}{}
	return nil
}

// generateSteps 按步骤进行初始化
func (p *ChainMakerPrepare) generatePrepareSteps(generateSteps []map[string]GenerateFunction) (err error) {
	fmt.Println()
	moduleNum := len(generateSteps)
	for idx, initStep := range generateSteps {
		for name, generateFunc := range initStep {
			if err = generateFunc(); err != nil {
				return fmt.Errorf("generate step [%s] failed, %w", name, err)
			}
			chainmakerPrepareWorkLogger.Infof("Generate STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
		}
	}
	fmt.Println()
	return
}

func (p *ChainMakerPrepare) GenerateCertsFiles() error {
	if _, ok := p.generateSteps[GenerateCertFiles]; ok {
		chainmakerPrepareWorkLogger.Errorf("already generate certs files")
		return nil
	}

	cryptoGenBinPath := fmt.Sprintf(p.pathMapping[ChainMakerCryptoGenBin])

	// 创建 build 文件夹
	err := os.MkdirAll(p.pathMapping[ChainMakerBuild], os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to Mkdir: %w", err)
	}

	// 进行加密文件的拷贝
	_, err = fileutils.CopyFile(fmt.Sprintf("%s/%s", p.pathMapping[Templates],
		p.pathMapping[CryptoConfigOldFileName]), p.pathMapping[CryptoConfigNewFileName])
	if err != nil {
		return fmt.Errorf("failed to copy crypto configuration file: %w", err)
	}

	// 进行 count: 4 的替换
	err = file.CopyAndReplaceTemplate(p.pathMapping[CryptoConfigNewFileName], p.pathMapping[CryptoConfigNewFileName], map[string]string{
		"count: 4": fmt.Sprintf("count: %d", p.nodeCount),
	})
	if err != nil {
		return fmt.Errorf("failed to copy and replace crypto config.yml %w", err)
	}

	// 进行 pkcs11_keys 文件的拷贝
	_, err = fileutils.CopyFile(fmt.Sprintf("%s/%s", p.pathMapping[Templates], p.pathMapping[PKCS11OldFileName]), p.pathMapping[PKCS11NewFileName])
	if err != nil {
		return fmt.Errorf("failed to copy and replace pkcs11 config.yml: %w", err)
	}

	// Generate crypto materials
	cmd := exec.Command(cryptoGenBinPath, "generate", "-c", p.pathMapping[CryptoConfigNewFileName], "-p", p.pathMapping[PKCS11OldFileName])
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}
	chainmakerPrepareWorkLogger.Infof("Certificates generated successfully.")

	// 将创建完成的 crypto-config 放入
	mvCommand := exec.Command("cp", "-r", "./crypto-config", "./build/crypto-config")
	err = mvCommand.Run()
	if err != nil {
		fmt.Printf("failed to copoy crypto-config: %v \n", err)
		return fmt.Errorf("failed to copy crypto-config: %w", err)
	}

	// 1.根据节点数量创建相应的文件夹
	// 2.将 crypto-config 之中的 wx-org{id}-chainmaker.org/ca/ca.crt 进行拷贝
	for i := 1; i <= p.nodeCount; i++ {
		for j := 1; j <= p.nodeCount; j++ {
			configCertsDir := fmt.Sprintf("%s/node%d/certs/ca/wx-org%d.chainmaker.org/", p.pathMapping[BuildConfig], i, j)
			configCertsName := path.Join(configCertsDir, "ca.crt")
			cryptoCertsName := fmt.Sprintf("%s/wx-org%d.chainmaker.org/ca/ca.crt", p.pathMapping[BuildCryptoConfig], j)

			// 进行文件夹的创建
			err = os.MkdirAll(configCertsDir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to mkdir: %w", err)
			}

			// 将文件拷贝到文件架之中
			_, err = fileutils.CopyFile(cryptoCertsName, configCertsName)
			if err != nil {
				return fmt.Errorf("failed to copy and replace ca: %w", err)
			}
		}
	}

	// 3. 还需要进行 node 文件架和 user 文件夹的拷贝
	for i := 1; i <= p.nodeCount; i++ {
		cryptoConfigNodeDir := fmt.Sprintf("%s/wx-org%d.chainmaker.org/node", p.pathMapping[BuildCryptoConfig], i)
		cryptoConfigUserDir := fmt.Sprintf("%s/wx-org%d.chainmaker.org/user", p.pathMapping[BuildCryptoConfig], i)
		buildConfigNodeCertsDir := fmt.Sprintf("%s/node%d/certs", p.pathMapping[BuildConfig], i)

		cmd = exec.Command("cp", "-r", cryptoConfigNodeDir, buildConfigNodeCertsDir)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to copy and replace node: %w", err)
		}
		cmd = exec.Command("cp", "-r", cryptoConfigUserDir, buildConfigNodeCertsDir)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to copy and replace node: %w", err)
		}
	}

	p.generateSteps[GenerateCertFiles] = struct{}{}
	return nil
}

// ResolvePeerIds 解析 peerId 列表
func (p *ChainMakerPrepare) ResolvePeerIds() error {
	if _, ok := p.generateSteps[ResolvePeerIds]; ok {
		chainmakerPrepareWorkLogger.Infof("already resolve peer ids")
		return nil
	}

	for i := 1; i <= p.nodeCount; i++ {
		peerId, err := p.ResolvePeerId(i)
		if err != nil {
			return fmt.Errorf("failed to resolve peer id: %w", err)
		}
		p.peerIdList = append(p.peerIdList, peerId)
		p.PeerIdToIndexMapping[peerId] = i - 1
	}

	chainmakerPrepareWorkLogger.Infof("successfully resolve peer ids")
	p.generateSteps[GenerateCertFiles] = struct{}{}
	return nil
}

// ResolvePeerId 进行 peerId 的解析
func (p *ChainMakerPrepare) ResolvePeerId(nodeId int) (string, error) {
	buildCryptoConfig := p.pathMapping[BuildCryptoConfig]
	peerIdFilePath := fmt.Sprintf("%s/wx-org%d.chainmaker.org/node/consensus1/consensus1.nodeid",
		buildCryptoConfig, nodeId)
	fileContent, err := os.Open(peerIdFilePath)
	if err != nil {
		return "", fmt.Errorf("resolve peer id failed")
	}
	peerIdBytes, err := io.ReadAll(fileContent)
	return string(peerIdBytes), nil
}

// GenerateLogYml 生成 log 配置文件
func (p *ChainMakerPrepare) GenerateLogYml() error {
	logTemplateFilePath := filepath.Join(p.pathMapping[Templates], "log.tpl")
	logLevel := configs.TopConfiguration.ChainMakerConfig.LogLevel
	for i := 1; i <= p.nodeCount; i++ {
		targetFilePath := fmt.Sprintf("%s/node%d/log.yml", p.pathMapping[BuildConfig], i)
		err := file.CopyAndReplaceTemplate(logTemplateFilePath, targetFilePath, map[string]string{
			"{log_level}": logLevel,
		})
		if err != nil {
			return fmt.Errorf("cannot generate log file %w", err)
		}
	}
	chainmakerPrepareWorkLogger.Infof("successfully generate log yml")
	return nil
}

func (p *ChainMakerPrepare) GenerateBcTemplate() error {
	if _, ok := p.generateSteps[GenerateBcTemplate]; ok {
		chainmakerPrepareWorkLogger.Infof("already generate bc template")
		return nil
	}

	finalOutput := ""
	resourcesPath := p.pathMapping[Resources]
	frontPartOfBcFile := p.pathMapping[FrontPartOfBCFile]
	backPartOfBcFile := p.pathMapping[BackPartOfBCFile]

	frontContent, err := os.ReadFile(frontPartOfBcFile)
	if err != nil {
		return fmt.Errorf("error reading front part of bc file: %w", err)
	}
	finalOutput += string(frontContent)

	consensusPartAndTrustRoots := `
# Consensus settings
consensus:
  # Consensus type
  # 0-SOLO, 1-TBFT, 3-MAXBFT, 4-RAFT, 5-DPOS, 6-ABFT
  type: {consensus_type}

  # Consensus node list
  nodes:
    # Each org has one or more consensus nodes.
    # We use p2p node id to represent nodes here.

`
	// 动态生成节点配置部分
	for i := 1; i <= p.nodeCount; i++ {
		consensusPartAndTrustRoots += fmt.Sprintf("    - org_id: \"{org%d_id}\"\n", i)
		consensusPartAndTrustRoots += fmt.Sprintf("      node_id:\n        - \"{org%d_peerid}\"\n", i)
	}

	consensusPartAndTrustRoots += `
  # We can specify other consensus config here in key-value format.
  ext_config:
    # - key: aa
    #   value: chain01_ext11

# Trust roots is used to specify the organizations' root certificates in permessionedWithCert mode.
# When in permessionedWithKey mode or public mode, it represents the admin users.
trust_roots:
  # org id and root file path list.

`

	// 动态生成 trust_roots 部分
	for i := 1; i <= p.nodeCount; i++ {
		consensusPartAndTrustRoots += fmt.Sprintf("  - org_id: \"{org%d_id}\"\n", i)
		consensusPartAndTrustRoots += fmt.Sprintf("    root:\n")
		consensusPartAndTrustRoots += fmt.Sprintf("      - \"../config/{org_path}/certs/ca/{org%d_id}/ca.crt\"\n", i)

	}

	finalOutput += consensusPartAndTrustRoots

	// -------------- 添加后面的部分的内容 ---------------
	backContent, err := os.ReadFile(backPartOfBcFile)
	if err != nil {
		return fmt.Errorf("error reading back part of bc file: %w", err)
	}
	finalOutput += string(backContent)
	// -------------- 添加后面的部分的内容 ---------------

	// 将生成的内容写入输出文件
	outputFilePath := filepath.Join(resourcesPath, fmt.Sprintf("bc_%d.yml", p.nodeCount))
	err = os.WriteFile(outputFilePath, []byte(finalOutput), 0644)
	if err != nil {
		return fmt.Errorf("error writing bc config file: %w", err)
	}

	chainmakerPrepareWorkLogger.Infof("GenerateBcTemplate successfully!")

	p.generateSteps[GenerateBcTemplate] = struct{}{}
	return nil
}

func (p *ChainMakerPrepare) GenerateBcYml() error {
	if _, ok := p.generateSteps[GenerateBcYml]; ok {
		chainmakerPrepareWorkLogger.Infof("already generate bc yml")
		return nil
	}

	for i := 1; i <= p.nodeCount; i++ {
		nodeChainConfigDir := fmt.Sprintf("%s/node%d/chainconfig", p.pathMapping[BuildConfig], i)
		err := os.MkdirAll(nodeChainConfigDir, os.ModePerm)
		if err != nil {
			chainmakerPrepareWorkLogger.Error("cannot generate dir")
			os.Exit(0)
		}
		sourceFilePath := filepath.Join(p.pathMapping[Resources], fmt.Sprintf("bc_%d.yml", p.nodeCount))
		targetFilePath := fmt.Sprintf("%s/bc1.yml", nodeChainConfigDir)
		_, err = fileutils.CopyFile(sourceFilePath, targetFilePath)
		if err != nil {
			chainmakerPrepareWorkLogger.Error("cannot copy file")
			os.Exit(1)
		}
	}

	chainmakerPrepareWorkLogger.Infof("GenerateBcYml successfully!")

	p.generateSteps[GenerateBcYml] = struct{}{}
	return nil
}

func (p *ChainMakerPrepare) ModifyBcYml() error {
	if _, ok := p.generateSteps[ModifyBcYml]; ok {
		chainmakerPrepareWorkLogger.Infof("already modify bc yml")
		return nil
	}

	replaceMap := map[string]string{
		"{chain_id}":       "chain1",
		"{version}":        "\"2030200\"",
		"{consensus_type}": fmt.Sprintf("%d", p.selectedConsensusType),
	}

	for i := 1; i <= p.nodeCount; i++ {
		replaceMap[fmt.Sprintf("{org%d_id}", i)] = fmt.Sprintf("wx-org%d.chainmaker.org", i)
		replaceMap[fmt.Sprintf("{org%d_peerid}", i)] = fmt.Sprintf("%s", p.peerIdList[i-1])
	}

	for i := 1; i <= p.nodeCount; i++ {
		bcFilePath := fmt.Sprintf("%s/node%d/chainconfig/bc1.yml", p.pathMapping[BuildConfig], i)

		replaceMap["{org_path}"] = fmt.Sprintf("wx-org%d.chainmaker.org", i)

		err := file.CopyAndReplaceTemplate(bcFilePath, bcFilePath, replaceMap)
		if err != nil {
			chainmakerPrepareWorkLogger.Errorf("cannot copy and replace file")
			os.Exit(1)
		}
	}

	chainmakerPrepareWorkLogger.Infof("Modify BC yml successfully!")
	p.generateSteps[ModifyBcYml] = struct{}{}
	return nil
}

func (p *ChainMakerPrepare) GenerateChainMakerYml() error {
	if _, ok := p.generateSteps[GenerateChainMakerYml]; ok {
		chainmakerPrepareWorkLogger.Infof("already generate chainmaker yml")
		return nil
	}

	p2pStartPort := configs.TopConfiguration.ChainMakerConfig.P2pStartPort
	rpcStartPort := configs.TopConfiguration.ChainMakerConfig.RpcStartPort
	vmGoRuntimePort := configs.TopConfiguration.ChainMakerConfig.VmGoRuntimePort
	vmGoEnginePort := configs.TopConfiguration.ChainMakerConfig.VmGoEnginePort
	consensusType := p.selectedConsensusType

	// 创建 peerIdList
	seedsString := "seeds:\n"
	for i := 1; i <= p.nodeCount; i++ {
		peerId := p.peerIdList[i-1]
		seedsString += fmt.Sprintf("    - \"/ip4/%s/tcp/%d/p2p/%s\"\n",
			p.ipv4Addresses[i-1],
			p2pStartPort+i-1,
			peerId)
	}

	// 利用 peerIdList 创建 seeds
	for i := 1; i <= p.nodeCount; i++ {
		templatesFilePath := p.pathMapping[Templates]
		chainMakerTemplateFilePath := filepath.Join(templatesFilePath, "chainmaker.tpl")
		targetFilePath := fmt.Sprintf("./build/config/node%d/chainmaker.yml", i)

		// 因为只有一个 chain 所以为 org_path1 和 bc1.yml
		blockChainStr := "blockchain:\n" +
			fmt.Sprintf("  - chainId: chain1\n") +
			fmt.Sprintf("    genesis: ../config/%s/chainconfig/bc1.yml\n", fmt.Sprintf("wx-org%d.chainmaker.org", i))

		err := file.CopyAndReplaceTemplate(chainMakerTemplateFilePath, targetFilePath, map[string]string{
			"{net_port}":                     fmt.Sprintf("%d", p2pStartPort+i-1),
			"{rpc_port}":                     fmt.Sprintf("%d", rpcStartPort+i-1),
			"{monitor_port}":                 fmt.Sprintf("%d", monitorPort+i-1),
			"{pprof_port}":                   fmt.Sprintf("%d", pprofPort+i-1),
			"{trusted_port}":                 fmt.Sprintf("%d", trustedPort+i-1),
			"{enable_vm_go}":                 enableVmGo,
			"{dockervm_container_name}":      fmt.Sprintf("%s%d", vmGoContainerNamePrefix, i),
			"{vm_go_runtime_port}":           fmt.Sprintf("%d", vmGoRuntimePort+i-1),
			"{vm_go_engine_port}":            fmt.Sprintf("%d", vmGoEnginePort+i-1),
			"{vm_go_log_level}":              "INFO",
			"{vm_go_protocol}":               "tcp",
			"seeds:":                         seedsString,
			"blockchain:":                    blockChainStr,
			"{consensus_type}":               fmt.Sprintf("%d", consensusType),
			"{org_id}":                       fmt.Sprintf("wx-org%d.chainmaker.org", i),
			"{org_path}":                     fmt.Sprintf("wx-org%d.chainmaker.org", i),
			"{org_path1}":                    fmt.Sprintf("wx-org%d.chainmaker.org", i),
			"{node_cert_path}":               "node/consensus1/consensus1.sign",
			"{net_cert_path}":                "node/consensus1/consensus1.tls",
			"{rpc_cert_path}":                "node/consensus1/consensus1.tls",
			"listen_addr: /ip4/0.0.0.0/tcp/": fmt.Sprintf("listen_addr: /ip4/%s/tcp/", p.ipv4Addresses[i-1]), //
		})
		if err != nil {
			return fmt.Errorf("cannot generate log file %w", err)
		}
	}

	chainmakerPrepareWorkLogger.Infof("GenerateChainMakerYml successfully!")

	p.generateSteps[GenerateChainMakerYml] = struct{}{}
	return nil
}

// CopyPrepareFiles 进行生成的文件的拷贝
func (p *ChainMakerPrepare) CopyPrepareFiles() error {
	if _, ok := p.generateSteps[CopyPrepareFiles]; ok {
		chainmakerPrepareWorkLogger.Infof("already copy prepare files")
		return nil
	}

	copyMap := map[string]string{
		p.pathMapping[BuildCryptoConfig]: p.pathMapping[TestData],
		p.pathMapping[BuildCryptoConfig]: p.pathMapping[CmdTestData],
		p.pathMapping[BuildConfig]:       p.pathMapping[MultiNode],
	}

	for source, target := range copyMap {
		cmd := exec.Command("cp", "-r", source, target)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("cp -r %s %s failed", source, target)
		}
	}

	chainmakerPrepareWorkLogger.Infof("copy prepare files successfully!")

	p.generateSteps[CopyPrepareFiles] = struct{}{}
	return nil
}
