package chainmaker_api

import (
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hokaccha/go-prettyjson"
)

const (
	OrgId1 = "wx-org1.chainmaker.org"
	OrgId2 = "wx-org2.chainmaker.org"
	OrgId3 = "wx-org3.chainmaker.org"
	OrgId4 = "wx-org4.chainmaker.org"
	OrgId5 = "wx-org5.chainmaker.org"

	UserNameOrg1Client1 = "org1client1"
	UserNameOrg2Client1 = "org2client1"

	UserNameOrg1Admin1 = "org1admin1"
	UserNameOrg2Admin1 = "org2admin1"
	UserNameOrg3Admin1 = "org3admin1"
	UserNameOrg4Admin1 = "org4admin1"
	UserNameOrg5Admin1 = "org5admin1"

	Version        = "1.0.0"
	UpgradeVersion = "2.0.0"
)

var users = map[string]*User{
	"org1client1": {
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.tls.key",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.tls.crt",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.key",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt",
	},
	"org2client1": {
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/client1/client1.tls.key",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/client1/client1.tls.crt",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/client1/client1.sign.key",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/client1/client1.sign.crt",
	},
	"org1admin1": {
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.tls.key",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.tls.crt",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key",
		"../../testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt",
	},
	"org2admin1": {
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.tls.key",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.tls.crt",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.key",
		"../../testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.crt",
	},
	"org3admin1": {
		"../../testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.tls.key",
		"../../testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.tls.crt",
		"../../testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.key",
		"../../testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.crt",
	},
	"org4admin1": {
		"../../testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.tls.key",
		"../../testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.tls.crt",
		"../../testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.sign.key",
		"../../testdata/crypto-config/wx-org4.chainmaker.org/user/admin1/admin1.sign.crt",
	},
	"org5admin1": {
		"../../testdata/crypto-config/wx-org5.chainmaker.org/user/admin1/admin1.tls.key",
		"../../testdata/crypto-config/wx-org5.chainmaker.org/user/admin1/admin1.tls.crt",
		"../../testdata/crypto-config/wx-org5.chainmaker.org/user/admin1/admin1.sign.key",
		"../../testdata/crypto-config/wx-org5.chainmaker.org/user/admin1/admin1.sign.crt",
	},
}
var permissionedPkUsers = map[string]*PermissionedPkUsers{
	"org1client1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org1/user/client1/client1.key",
		OrgId1,
	},
	"org2client1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org2/user/client1/client1.key",
		OrgId2,
	},
	"org1admin1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org1/user/admin1/admin1.key",
		OrgId1,
	},
	"org2admin1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org2/user/admin1/admin1.key",
		OrgId2,
	},
	"org3admin1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org3/user/admin1/admin1.key",
		OrgId3,
	},
	"org4admin1": {
		"../../testdata/crypto-config-pk/permissioned-with-key/wx-org4/user/admin1/admin1.key",
		OrgId4,
	},
}

var pkUsers = map[string]*PkUsers{
	"org1client1": {
		"../../testdata/crypto-config-pk/public/user/user1/user1.key",
	},
	"org2client1": {
		"../../testdata/crypto-config-pk/public/user/user2/user2.key",
	},
	"org1admin1": {
		"../../testdata/crypto-config-pk/public/admin/admin1/admin1.key",
	},
	"org2admin1": {
		"../../testdata/crypto-config-pk/public/admin/admin2/admin2.key",
	},
	"org3admin1": {
		"../../testdata/crypto-config-pk/public/admin/admin3/admin3.key",
	},
	"org4admin1": {
		"../../testdata/crypto-config-pk/public/admin/admin4/admin4.key",
	},
}

type PkUsers struct {
	SignKeyPath string
}

type PermissionedPkUsers struct {
	SignKeyPath string
	OrgId       string
}

type User struct {
	TlsKeyPath, TlsCrtPath   string
	SignKeyPath, SignCrtPath string
}

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

func GetEndorsersWithAuthType(hashType crypto.HashType, authType sdk.AuthType, payload *common.Payload, usernames ...string) ([]*common.EndorsementEntry, error) {
	var endorsers []*common.EndorsementEntry

	for _, name := range usernames {
		var entry *common.EndorsementEntry
		var err error
		switch authType {
		case sdk.PermissionedWithCert:
			u, ok := users[name]
			if !ok {
				return nil, errors.New("user not found")
			}
			entry, err = sdkutils.MakeEndorserWithPath(u.SignKeyPath, u.SignCrtPath, payload)
			if err != nil {
				return nil, err
			}

		case sdk.PermissionedWithKey:
			u, ok := permissionedPkUsers[name]
			if !ok {
				return nil, errors.New("user not found")
			}
			entry, err = sdkutils.MakePkEndorserWithPath(u.SignKeyPath, hashType, u.OrgId, payload)
			if err != nil {
				return nil, err
			}

		case sdk.Public:
			u, ok := pkUsers[name]
			if !ok {
				return nil, errors.New("user not found")
			}
			entry, err = sdkutils.MakePkEndorserWithPath(u.SignKeyPath, hashType, "", payload)
			if err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("invalid authType")
		}
		endorsers = append(endorsers, entry)
	}

	return endorsers, nil
}
