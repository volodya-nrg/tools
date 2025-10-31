package freeipa

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volodya-nrg/tools/pkg/funcs"
)

func TestFreeIPA(t *testing.T) { //nolint:tparallel
	t.Parallel()

	const (
		password1     = "password1"
		password2     = "password2"
		scheme        = "https"
		host          = "ipa-dev-nms.fraxis.ru"
		adminLogin    = "admin"
		adminPassword = "lk239d81llcSlk932UoPRds"
		timeout       = 5 * time.Second
	)

	// создадим клиента
	cl, err := NewFreeIPA(scheme, host, timeout)
	require.NoError(t, err)
	require.NotNil(t, cl)

	t.Run("check users", func(t *testing.T) { //nolint:paralleltest
		// зайдем под админом
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPassword))

		newUserID := funcs.RandStr()

		// создадим пользователя
		reqUser := RequestUser{
			UID:                   newUserID,
			GivenName:             newUserID + "-firstname",
			SN:                    newUserID + "-lastname",
			Mail:                  funcs.Pointer(newUserID + "@example.com"),
			UserPassword:          funcs.Pointer(password1),
			KRBPasswordExpiration: funcs.Pointer(time.Now().AddDate(0, 3, 0)),
			NsAccountLock:         funcs.Pointer(false),
		}
		userExpected, err := cl.CreateUser(t.Context(), reqUser)
		require.NoError(t, err)

		// получим пользователя
		userActual, err := cl.GetUser(t.Context(), userExpected.UID)
		require.NoError(t, err)
		require.Equal(t, userExpected, userActual)

		// получим дефолтный диапазон (20) пользователей
		users, total, err := cl.GetUsers(t.Context(), -1, -1) // limit=default, offset=0
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(users), 2) // admin и новый пользователь
		require.Len(t, users, int(total))

		// выйдем из админа
		require.NoError(t, cl.Logout(t.Context()))

		// проверим можно ли получить пользователя из под гостя
		_, err = cl.GetUser(t.Context(), newUserID)
		require.Error(t, err)

		// зайдем под новым пользователем, и выйдем
		require.NoError(t, cl.Login(t.Context(), newUserID, password1))
		require.NoError(t, cl.Logout(t.Context()))

		// зайдем под админом, чтоб обновить все поля, т.к. у него привилегий больше
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPassword))

		// обновим пользователя
		newPassExp := time.Now().AddDate(0, 3, 0)
		reqUser = RequestUser{
			UID:                   newUserID,
			GivenName:             newUserID + "-firstname2",
			SN:                    newUserID + "-lastname2",
			Mail:                  funcs.Pointer(newUserID + "@example.ru"),
			UserPassword:          funcs.Pointer(password2), //nolint:lll // если сбросим пароль, то время жизни тоже сбрасывается до текущего
			KRBPasswordExpiration: funcs.Pointer(newPassExp),
			NsAccountLock:         funcs.Pointer(true),
		}
		require.NoError(t, cl.UpdateUser(t.Context(), reqUser))

		// проверим измененные данные
		userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, reqUser.UID, userActual.UID)
		require.Equal(t, reqUser.GivenName, userActual.GivenName)
		require.Equal(t, reqUser.SN, userActual.SN)
		require.Equal(t, *reqUser.Mail, userActual.Mail)
		require.True(t, newPassExp.After(userActual.KRBPasswordExpiration))
		require.Equal(t, *reqUser.NsAccountLock, userActual.NsAccountLock)

		// создадим роль и назначим ее пользователю
		roleName := funcs.RandStr()
		require.NoError(t, cl.CreateRole(t.Context(), roleName, nil))
		require.NoError(t, cl.ToggleRoleForUser(t.Context(), roleName, newUserID))

		// проверим что роль присутствует у пользователя
		userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Contains(t, userActual.MemberOfRole, roleName)

		// привяжем данную роль и к админу
		require.NoError(t, cl.ToggleRoleForUser(t.Context(), roleName, adminLogin))

		// запросим роль и проверим сколько у скольких человек она привязана, а так же кто привязан
		roleActual, err := cl.GetRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Len(t, roleActual.MemberUser, 2)
		require.Contains(t, roleActual.MemberUser, newUserID)
		require.Contains(t, roleActual.MemberUser, adminLogin)

		// запросим список ролей и посмотрим в нужной привязанных людей
		roles, total, err := cl.GetRoles(t.Context(), -1, -1)
		require.NoError(t, err)
		require.Positive(t, total)
		isHasNeedUsers := false
		for _, v := range roles {
			if v.CN == roleName {
				require.Len(t, v.MemberUser, 2)
				require.Contains(t, v.MemberUser, newUserID)
				require.Contains(t, v.MemberUser, adminLogin)
				isHasNeedUsers = true
				break
			}
		}
		require.True(t, isHasNeedUsers)

		// удалим привязку роли от пользователя
		require.NoError(t, cl.ToggleRoleForUser(t.Context(), roleName, newUserID))

		// проверим что у пользователя нет привязки к этой роли
		userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.NotContains(t, userActual.MemberOfRole, roleName)

		// удалим роль и пользователя
		require.NoError(t, cl.DeleteRole(t.Context(), roleName))
		require.NoError(t, cl.DeleteUser(t.Context(), newUserID))

		// проверим что пользователя нет
		_, err = cl.GetUser(t.Context(), newUserID)
		require.Error(t, err)

		// выйдем из админа
		require.NoError(t, cl.Logout(t.Context()))
	})
	t.Run("check roles", func(t *testing.T) { //nolint:paralleltest
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPassword))

		roleName := funcs.RandStr()
		roleDesc := funcs.RandStr()

		// создадим роль
		require.NoError(t, cl.CreateRole(t.Context(), roleName, funcs.Pointer(roleDesc)))

		// получим роль
		role, err := cl.GetRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, roleName, role.CN)
		require.Equal(t, roleDesc, role.Description)

		// изменим роль
		roleDesc2 := funcs.RandStr()
		require.NoError(t, cl.UpdateRole(t.Context(), role.CN, roleDesc2))

		// получим роль
		role, err = cl.GetRole(t.Context(), role.CN)
		require.NoError(t, err)
		require.Equal(t, roleDesc2, role.Description)

		// проверим что данная роль появилась
		roles, total, err := cl.GetRoles(t.Context(), -1, -1)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, uint32(1))
		require.True(t, slices.ContainsFunc(roles, func(r Role) bool {
			return r.CN == roleName
		}))

		// проверим список v1
		roles, total, err = cl.GetRoles(t.Context(), 1, 0)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, uint32(1))
		require.Len(t, roles, 1)

		// проверим список v2, должно быть <= 20
		roles, total, err = cl.GetRoles(t.Context(), math.MaxInt32, -1)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, uint32(1))
		require.LessOrEqual(t, len(roles), limitDefault)

		// проверим список v3
		roles, total, err = cl.GetRoles(t.Context(), -1, math.MaxInt32)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, uint32(1))
		require.LessOrEqual(t, len(roles), limitDefault)

		// удалим роль
		require.NoError(t, cl.DeleteRole(t.Context(), roleName))

		// получим роль
		_, err = cl.GetRole(t.Context(), roleName)
		require.Error(t, err)

		// выйдем из под админа
		require.NoError(t, cl.Logout(t.Context()))
	})
}
