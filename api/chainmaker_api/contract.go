package chainmaker_api

import (
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"zhanghefan123/security_topology/api/chainmaker_api/types"
)

type createUpgradeContractOp int

const (
	CreateContractOp createUpgradeContractOp = iota + 1
	UpgradeContractOp
)

// 示例创建合约代码
// ./cmc client contract user create \
//--contract-name=fact \
//--runtime-type=WASMER \
//--byte-code-path=./testdata/claim-wasm-demo/rust-fact-2.0.0.wasm \
//--version=1.0 \
//--sdk-conf-path=./testdata/sdk_config.yml \
//--admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.key \
//--admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.crt \
//--sync-result=true \
//--params="{}"

// CreateUpgradeUserContract 创建或升级合约
func CreateUpgradeUserContract(client *sdk.ChainClient, clientConfiguration *ClientConfiguration, op createUpgradeContractOp) error {
	adminKeys, adminCerts, adminOrgs, err := MakeAdminInfo(client,
		clientConfiguration.AdminKeyFilePaths,
		clientConfiguration.AdminCertFilePaths,
		clientConfiguration.AdminOrgIds)
	if err != nil {
		return fmt.Errorf("cannot create admin info: %w", err)
	}

	rt, ok := common.RuntimeType_value[clientConfiguration.RuntimeType]
	if !ok {
		return fmt.Errorf("unknown runtime type: %s", clientConfiguration.RuntimeType)
	}

	var kvs []*common.KeyValuePair

	if clientConfiguration.RuntimeType != "EVM" {
		if clientConfiguration.Params != "" {
			kvsMap := make(map[string]interface{})
			err = json.Unmarshal([]byte(clientConfiguration.Params), &kvsMap)
			if err != nil {
				return err
			}
			kvs = ConvertParameters(kvsMap)
		}
	}

	var payload *common.Payload
	switch op {
	case CreateContractOp:
		payload, err = client.CreateContractCreatePayload(clientConfiguration.ContractName, clientConfiguration.Version,
			clientConfiguration.ByteCodePath, common.RuntimeType(rt), kvs)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown operation")
	}

	if clientConfiguration.GasLimit > 0 {
		var limit = &common.Limit{GasLimit: clientConfiguration.GasLimit}
		payload = client.AttachGasLimit(payload, limit)
	}

	endorsementEntrys, err := MakeEndorsement(adminKeys, adminCerts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	var payer []*common.EndorsementEntry
	if len(clientConfiguration.PayerKeyFilePath) > 0 {
		payer, err = MakeEndorsement([]string{clientConfiguration.PayerKeyFilePath}, []string{clientConfiguration.PayerCrtFilePath}, []string{clientConfiguration.PayerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return err
		}
	}

	var resp *common.TxResponse
	if len(payer) == 0 {
		resp, err = client.SendContractManageRequest(payload, endorsementEntrys, clientConfiguration.Timeout, clientConfiguration.SyncResult)
	} else {
		resp, err = client.SendContractManageRequestWithPayer(payload, endorsementEntrys, payer[0], clientConfiguration.Timeout, clientConfiguration.SyncResult)
	}
	if err != nil {
		return err
	}
	err = CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}
	return createUpgradeUserContractOutput(resp)
}

// createUpgradeUserContractOutput 用来进行合约的升级
func createUpgradeUserContractOutput(resp *common.TxResponse) error {
	if resp.ContractResult != nil && resp.ContractResult.Result != nil {
		var contract common.Contract
		err := contract.Unmarshal(resp.ContractResult.Result)
		if err != nil {
			return err
		}
		PrintPrettyJson(types.CreateUpgradeContractTxResponse{
			TxResponse: resp,
			ContractResult: &types.CreateUpgradeContractContractResult{
				ContractResult: resp.ContractResult,
				Result:         &contract,
			},
		})
	} else {
		PrintPrettyJson(resp)
	}
	return nil
}

// invokeContract 是用来进行合约的调用
func invokeContract(contractName string, client *sdk.ChainClient, method string, withSyncResult bool) (*common.TxResponse, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	fileHash := uuid.GetUUID()
	kvs := []*common.KeyValuePair{
		{
			Key:   "time",
			Value: []byte(curTime),
		},
		{
			Key:   "file_hash",
			Value: []byte(fileHash),
		},
		{
			Key:   "file_name",
			Value: []byte(fmt.Sprintf("file_%s", curTime)),
		},
	}
	return client.InvokeContract(contractName, method, "", kvs, -1, withSyncResult) // 进行合约的调用
}
