package chainmaker_api

import (
	"path/filepath"
	"zhanghefan123/security_topology/configs"
)

type ClientConfiguration struct {
	SdkConfigPath       string
	RuntimeType         string
	ContractName        string
	Version             string
	ByteCodePath        string
	ChainId             string
	OrgId               string
	UserTlsCrtFilePath  string
	UserTlsKeyFilePath  string
	UserSignCrtFilePath string
	UserSignKeyFilePath string
	AdminKeyFilePaths   string
	AdminCertFilePaths  string
	PayerKeyFilePath    string
	PayerCrtFilePath    string
	PayerOrgId          string
	AdminOrgIds         string
	Params              string
	GasLimit            uint64
	Timeout             int64
	SyncResult          bool
}

func NewClientConfiguration(contractName string) *ClientConfiguration {
	sdkConfigFilePath := filepath.Join(configs.TopConfiguration.PathConfig.ResourcesPath, "client_config/sdk_config.yml")
	return &ClientConfiguration{
		SdkConfigPath:       sdkConfigFilePath,
		RuntimeType:         "WASMER",
		ContractName:        contractName,
		Version:             "1.0",
		ByteCodePath:        "./testdata/demos/claim-wasm-demo/rust-fact-2.0.0.wasm",
		ChainId:             "",
		OrgId:               "",
		UserTlsCrtFilePath:  "",
		UserTlsKeyFilePath:  "",
		UserSignCrtFilePath: "",
		UserSignKeyFilePath: "",
		AdminKeyFilePaths:   "./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.key",
		AdminCertFilePaths:  "./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.crt",
		PayerKeyFilePath:    "",
		AdminOrgIds:         "",
		Params:              "{}",
		GasLimit:            0,
		Timeout:             10,
		SyncResult:          true,
	}
}
