package chainmaker_api

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"errors"
	"fmt"
)

func CheckProposalRequestResp(resp *common.TxResponse, needContractResult bool) error {
	if resp.Code != common.TxStatusCode_SUCCESS {
		if resp.Message == "" {
			if resp.ContractResult != nil && resp.ContractResult.Code != 0 && resp.ContractResult.Message != "" {
				return errors.New(resp.ContractResult.Message)
			}
			return errors.New(resp.Code.String())
		}
		return errors.New(resp.Message)
	}

	if needContractResult && resp.ContractResult == nil {
		return fmt.Errorf("contract result is nil")
	}

	if resp.ContractResult != nil && resp.ContractResult.Code != 0 {
		return errors.New(resp.ContractResult.Message)
	}

	return nil
}
