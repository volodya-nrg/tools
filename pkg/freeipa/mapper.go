package freeipa

import (
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

	if v, ok := m["nsaccountlock"]; ok && isBool(v) {
		user.NsAccountLock = v.(bool) //nolint:forcetypeassert
	}
	if v, ok := m["uid"]; ok && isNotEmptySlice(v) {
		user.UID = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m["dn"]; ok {
		user.DN = fmt.Sprintf("%v", v)
	}
	if v, ok := m["memberof_group"]; ok && isNotEmptySlice(v) {
		user.MemberOfGroup = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	if v, ok := m["memberof_role"]; ok && isNotEmptySlice(v) {
		user.MemberOfRole = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	// это не нужно, где-то оно прилетает, где-то нет
	// if v, ok := m["has_password"]; ok && isBool(v) {
	//	user.HasPassword = v.(bool) //nolint:forcetypeassert
	// }
	if v, ok := m["sn"]; ok && isNotEmptySlice(v) {
		user.SN = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	// данного св-ва может и не быть если пароль уже просрочен
	if v, ok := m["krbpasswordexpiration"]; ok && isNotEmptySlice(v) {
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
	if v, ok := m["mail"]; ok && isNotEmptySlice(v) {
		user.Mail = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m["givenname"]; ok && isNotEmptySlice(v) {
		user.GivenName = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}

	return user
}

func mapUserToDTORole(m map[string]any) Role {
	role := Role{}

	if v, ok := m["description"]; ok && isNotEmptySlice(v) {
		role.Description = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m["cn"]; ok && isNotEmptySlice(v) {
		role.CN = convertSliceAnyToSliceStr(v.([]any))[0] //nolint:forcetypeassert
	}
	if v, ok := m["objectclass"]; ok && isNotEmptySlice(v) {
		role.ObjectClass = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}
	if v, ok := m["dn"]; ok {
		role.DN = fmt.Sprintf("%v", v)
	}
	if v, ok := m["member_user"]; ok && isNotEmptySlice(v) {
		role.MemberUser = convertSliceAnyToSliceStr(v.([]any)) //nolint:forcetypeassert
	}

	return role
}
