package chainmaker_api

import (
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"fmt"
)

// CreateChainMakerClient 创建 chainmakerClient
func CreateChainMakerClient(clientConfiguration *ClientConfiguration) (*sdk.ChainClient, error) {
	chainMakerClient, err := sdk.NewChainClient(
		sdk.WithConfPath(clientConfiguration.SdkConfigPath),
		sdk.WithChainClientChainId(""),
		sdk.WithChainClientOrgId(""),
		sdk.WithUserCrtFilePath(""),
		sdk.WithUserCrtFilePath(""),
		sdk.WithUserKeyFilePath(""),
		sdk.WithUserSignCrtFilePath(""),
		sdk.WithUserSignKeyFilePath(""),
		sdk.WithEnableTxResultDispatcher(true),
		sdk.WithRetryLimit(20),
		sdk.WithRetryInterval(2000),
	)
	if err != nil {
		return nil, fmt.Errorf("create chain maker client error: %w", err)
	}
	return chainMakerClient, nil
}
