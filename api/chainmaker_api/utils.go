package chainmaker_api

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"encoding/json"
	"fmt"
	"github.com/hokaccha/go-prettyjson"
)

func ConvertParameters(pars map[string]interface{}) []*common.KeyValuePair {
	var kvp []*common.KeyValuePair
	for k, v := range pars {
		var value string
		switch v := v.(type) {
		case string:
			value = v
		default:
			bz, err := json.Marshal(v)
			if err != nil {
				return nil
			}
			value = string(bz)
		}
		kvp = append(kvp, &common.KeyValuePair{
			Key:   k,
			Value: []byte(value),
		})
	}
	return kvp
}

func MakeEndorsement(adminKeys, adminCrts, adminOrgs []string, client *sdk.ChainClient, payload *common.Payload) (
	[]*common.EndorsementEntry, error) {
	endorsementEntrys := make([]*common.EndorsementEntry, len(adminKeys))
	for i := range adminKeys {
		if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
			e, err := sdkutils.MakeEndorserWithPath(adminKeys[i], adminCrts[i], payload)
			if err != nil {
				return nil, err
			}
			endorsementEntrys[i] = e
		} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
			e, err := sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				client.GetHashType(),
				adminOrgs[i],
				payload,
			)
			if err != nil {
				return nil, err
			}

			endorsementEntrys[i] = e
		} else {
			e, err := sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				client.GetHashType(),
				"",
				payload,
			)
			if err != nil {
				return nil, err
			}

			endorsementEntrys[i] = e
		}
	}
	return endorsementEntrys, nil
}

// PrintPrettyJson print pretty json of data
func PrintPrettyJson(data interface{}) {
	output, err := prettyjson.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
}
