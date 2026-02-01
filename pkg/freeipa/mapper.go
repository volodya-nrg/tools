package freeipa

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"
)

/*
mapUserToDTOUser маппер пришедших данных в целевую dto

	"ipantsecurityidentifier": []"S-1-5-21-468653659-1072500696-132053521-500",
	"nsaccountlock": false,
	"gidnumber": []"371000000",
	"loginshell": []"/bin/bash",
	"krbprincipalname": []"admin@NMS.FRAXIS.RU","root@NMS.FRAXIS.RU",
	"krbcanonicalname": []"admin@NMS.FRAXIS.RU",
	"uid": []"admin",
	"preserved": false,
	"ipauniqueid": []"579d3fec-add5-11f0-99cb-7a02fc2ca798",
	"krblastadminunlock": map["__datetime__"]"20251020165845Z",
	"krbextradata": map["__base64__"]"AAJFavZocm9vdC9hZG1pbkBOTVMuRlJBWElTLlJVAA==",
	"memberof_group": []"trust admins","admins",
	"memberof_role": []"roleName",
	"dn": "uid=admin,cn=users,cn=accounts,dc=nms,dc=fraxis,dc=ru",
	"objectclass": []"top","person","posixaccount","krbprincipalaux","krbticketpolicyaux",...
	"gecos": []"Administrator",
	"krbloginfailedcount": []"0",
	"homedirectory": []"/home/admin",
	"uidnumber": []"371000000",
	"cn": []"Administrator",
	"krblastpwdchange": map["__datetime__"]"20251020165845Z",
	"krbpasswordexpiration": map["__datetime__"]"20251020165845Z",
	"krblastfailedauth": map["__datetime__"]"20251020165845Z",
	"sn": []"Administrator",
	"has_password": false,
	"has_keytab": true,
*/
func mapUserToDTOUser(m map[string]any) User {
	user := User{}

	if v, ok := m[keyOptNSAccountLock]; ok && isBool(v) {
		user.NsAccountLock = v.(bool) //nolint:forcetypeassert
	}
	if v, ok := m[keyOptUID]; ok && isNotEmptySlice(v) {
		user.UID = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptDN]; ok {
		user.DN = fmt.Sprintf("%v", v)
	}
	if v, ok := m[keyOptMemberofGroup]; ok && isNotEmptySlice(v) {
		user.MemberOfGroup = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	if v, ok := m[keyOptMemberofRole]; ok && isNotEmptySlice(v) {
		user.MemberOfRole = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	// это не нужно, где-то оно прилетает, где-то нет
	// if v, ok := m["has_password"]; ok && isBool(v) {
	//	user.HasPassword = v.(bool) //nolint:forcetypeassert
	// }
	if v, ok := m[keyOptSN]; ok && isNotEmptySlice(v) {
		user.SN = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	// данного св-ва может и не быть если пароль уже просрочен
	if v, ok := m[keyOptKRBPasswordExpiration]; ok && isNotEmptySlice(v) {
		if m2, ok2 := v.([]any)[0].(map[string]any); ok2 {
			if v3, ok3 := m2["__datetime__"]; ok3 {
				if tLoc, err := time.Parse(timeLayout, fmt.Sprintf("%v", v3)); err != nil {
					slog.Error("failed to parse time", slog.String("err", err.Error())) //nolint:noctx
				} else {
					user.KRBPasswordExpiration = tLoc
				}
			}
		}
	}
	if v, ok := m[keyOptMail]; ok && isNotEmptySlice(v) {
		user.Mail = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptGivenName]; ok && isNotEmptySlice(v) {
		user.GivenName = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptCN]; ok && isNotEmptySlice(v) {
		user.CN = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptTelephoneNumber]; ok && isNotEmptySlice(v) {
		user.TelephoneNumber = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptMobile]; ok && isNotEmptySlice(v) {
		user.Mobile = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptTitle]; ok && isNotEmptySlice(v) {
		user.Title = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptOU]; ok && isNotEmptySlice(v) {
		user.OrgUnit = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptO]; ok && isNotEmptySlice(v) {
		user.Organization = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptJPEGPhoto]; ok && isNotEmptySlice(v) {
		if m2, ok2 := v.([]any)[0].(map[string]any); ok2 {
			if v3, ok3 := m2["__base64__"]; ok3 {
				if decoded, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", v3)); err == nil {
					user.JPEGPhoto = string(decoded)
				}
			}
		}
	}

	return user
}

func mapUserToDTORole(m map[string]any) Role {
	role := Role{}

	if v, ok := m[keyOptDescription]; ok && isNotEmptySlice(v) {
		role.Description = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptCN]; ok && isNotEmptySlice(v) {
		role.CN = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m[keyOptObjectClass]; ok && isNotEmptySlice(v) {
		role.ObjectClass = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	if v, ok := m[keyOptDN]; ok {
		role.DN = fmt.Sprintf("%v", v)
	}
	if v, ok := m[keyOptMemberUser]; ok && isNotEmptySlice(v) {
		role.MemberUser = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}

	return role
}
