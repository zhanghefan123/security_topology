package chainmaker_api

import (
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"fmt"
	"strings"
)

const (
	// AdminOrgidKeyCertLengthNotEqualFormat define AdminOrgidKeyCertLengthNotEqualFormat error fmt
	AdminOrgidKeyCertLengthNotEqualFormat = "admin orgId & key & cert list length not equal, [keys len: %d]/[certs len:%d]"
	// AdminOrgidKeyLengthNotEqualFormat define AdminOrgidKeyLengthNotEqualFormat error fmt
	AdminOrgidKeyLengthNotEqualFormat = "admin orgId & key list length not equal, [keys len: %d]/[org-ids len:%d]"
)

func MakeAdminInfo(client *sdk.ChainClient, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds string) (
	adminKeys, adminCrts, adminOrgs []string, err error) {
	if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminCrtFilePaths != "" {
			adminCrts = strings.Split(adminCrtFilePaths, ",")
		}
		if len(adminKeys) != len(adminCrts) {
			err = fmt.Errorf(AdminOrgidKeyCertLengthNotEqualFormat, len(adminKeys), len(adminCrts))
		}
	} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminOrgIds != "" {
			adminOrgs = strings.Split(adminOrgIds, ",")
		}
		if len(adminKeys) != len(adminOrgs) {
			err = fmt.Errorf(AdminOrgidKeyLengthNotEqualFormat, len(adminKeys), len(adminOrgs))
		}
	} else {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
	}
	return
}
