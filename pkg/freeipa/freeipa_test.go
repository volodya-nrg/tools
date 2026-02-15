package freeipa

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volodya-nrg/tools/pkg/funcs"
)

func TestFreeIPA(t *testing.T) { //nolint:tparallel
	t.Parallel()

	const (
		password1 = "password1"
		password2 = "password2"
		timeout   = 5 * time.Second
	)

	configBytes, err := os.ReadFile("config.json")
	require.NoError(t, err)

	configLoc := &config{}
	require.NoError(t, json.Unmarshal(configBytes, configLoc))

	var (
		adminLogin = configLoc.Login
		adminPass  = configLoc.Password
		cl         = NewFreeIPA(configLoc.Scheme, configLoc.Host, &http.Transport{}, timeout)
	)

	t.Run("check users", func(t *testing.T) { //nolint:paralleltest
		// зайдем под админом
		statusCode, err := cl.Login(t.Context(), adminLogin, adminPass)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

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
			CN:                    funcs.Pointer(newUserID + "-fullname"),
			TelephoneNumber:       funcs.Pointer(newUserID + "-telephone-number"),
			Mobile:                funcs.Pointer(newUserID + "-mobile"),
			Title:                 funcs.Pointer(newUserID + "-title"),
			OU:                    funcs.Pointer(newUserID + "-ou"),
			AddAttr:               []string{"o=MyCompany", "jpegphoto=path/to/photo.jpg"},
		}
		statusCode, userExpected, err := cl.CreateUser(t.Context(), reqUser)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// получим пользователя
		statusCode, userActual, err := cl.GetUser(t.Context(), userExpected.UID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Equal(t, userExpected, userActual)

		// получим дефолтный диапазон (20) пользователей
		statusCode, users, total, err := cl.GetUsers(t.Context(), -1, -1) // limit=default, offset=0
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.GreaterOrEqual(t, len(users), 2) // admin и новый пользователь
		require.GreaterOrEqual(t, int(total), len(users))

		// выйдем из админа
		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// проверим еще раз выход
		statusCode, err = cl.Logout(t.Context())
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, statusCode)

		// проверим можно ли получить пользователя из под гостя
		statusCode, _, err = cl.GetUser(t.Context(), newUserID)
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, statusCode)

		// зайдем под новым пользователем, и выйдем
		statusCode, err = cl.Login(t.Context(), newUserID, password1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// зайдем под админом, чтоб обновить все поля, т.к. у него привилегий больше
		statusCode, err = cl.Login(t.Context(), adminLogin, adminPass)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

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
		statusCode, err = cl.UpdateUser(t.Context(), reqUser)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// проверим измененные данные
		statusCode, userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Equal(t, reqUser.UID, userActual.UID)
		require.Equal(t, reqUser.GivenName, userActual.GivenName)
		require.Equal(t, reqUser.SN, userActual.SN)
		require.Equal(t, *reqUser.Mail, userActual.Mail)
		require.True(t, newPassExp.After(userActual.KRBPasswordExpiration))
		require.Equal(t, *reqUser.NsAccountLock, userActual.NsAccountLock)

		// создадим роль и назначим ее пользователю
		roleName := funcs.RandStr()
		statusCode, err = cl.CreateRole(t.Context(), roleName, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		statusCode, err = cl.ToggleRoleForUser(t.Context(), roleName, newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// проверим что роль присутствует у пользователя
		statusCode, userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Contains(t, userActual.MemberOfRole, roleName)

		// привяжем данную роль и к админу
		statusCode, err = cl.ToggleRoleForUser(t.Context(), roleName, adminLogin)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// запросим роль и проверим сколько у скольких человек она привязана, а так же кто привязан
		statusCode, roleActual, err := cl.GetRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Len(t, roleActual.MemberUser, 2)
		require.Contains(t, roleActual.MemberUser, newUserID)
		require.Contains(t, roleActual.MemberUser, adminLogin)

		// запросим список ролей и посмотрим в нужной привязанных людей
		statusCode, roles, total, err := cl.GetRoles(t.Context(), -1, -1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
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
		statusCode, err = cl.ToggleRoleForUser(t.Context(), roleName, newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// проверим что у пользователя нет привязки к этой роли
		statusCode, userActual, err = cl.GetUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.NotContains(t, userActual.MemberOfRole, roleName)

		// удалим роль
		statusCode, err = cl.DeleteRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// удалим пользователя
		statusCode, err = cl.DeleteUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// проверим что пользователя нет
		statusCode, _, err = cl.GetUser(t.Context(), newUserID)
		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, statusCode)

		// выйдем из админа
		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
	})
	t.Run("check roles", func(t *testing.T) { //nolint:paralleltest
		statusCode, err := cl.Login(t.Context(), adminLogin, adminPass)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		roleName := funcs.RandStr()
		roleDesc := funcs.RandStr()

		// создадим роль
		statusCode, err = cl.CreateRole(t.Context(), roleName, funcs.Pointer(roleDesc))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// получим роль
		statusCode, role, err := cl.GetRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Equal(t, roleName, role.CN)
		require.Equal(t, roleDesc, role.Description)

		// проверим явно что она есть
		statusCode, isHas, err := cl.HasRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.True(t, isHas)

		// изменим роль
		roleDesc2 := funcs.RandStr()
		statusCode, err = cl.UpdateRole(t.Context(), role.CN, roleDesc2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// получим роль
		statusCode, role, err = cl.GetRole(t.Context(), role.CN)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Equal(t, roleDesc2, role.Description)

		// проверим что данная роль появилась
		statusCode, roles, total, err := cl.GetRoles(t.Context(), -1, -1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.GreaterOrEqual(t, total, uint32(1))
		require.True(t, slices.ContainsFunc(roles, func(r Role) bool {
			return r.CN == roleName
		}))

		roleNames := make([]string, len(roles))
		for i, roleLoc := range roles {
			roleNames[i] = roleLoc.CN
		}

		// проверим список v1
		statusCode, roles, total, err = cl.GetRoles(t.Context(), 1, 0)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.GreaterOrEqual(t, total, uint32(1))
		require.Len(t, roles, 1)

		// проверим список v2, должно быть <= 20
		statusCode, roles, total, err = cl.GetRoles(t.Context(), math.MaxInt32, -1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.GreaterOrEqual(t, total, uint32(1))
		require.LessOrEqual(t, len(roles), limitDefault)

		// проверим список v3
		statusCode, roles, total, err = cl.GetRoles(t.Context(), -1, math.MaxInt32)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.GreaterOrEqual(t, total, uint32(1))
		require.LessOrEqual(t, len(roles), limitDefault)

		// проверим список по именам
		statusCode, roles, err = cl.GetRolesByName(t.Context(), roleNames)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Len(t, roles, len(roleNames))

		// получим по именам, но запросим левую роль, нужно чтоб response отработал корректно
		statusCode, _, err = cl.GetRolesByName(t.Context(), []string{funcs.RandStr()})
		require.Error(t, err)
		require.Equal(t, 0, statusCode)

		// удалим роль
		statusCode, err = cl.DeleteRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// получим роль
		statusCode, _, err = cl.GetRole(t.Context(), roleName)
		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, statusCode)

		// проверим через другой метод
		statusCode, isHas, err = cl.HasRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.False(t, isHas)

		// выйдем из под админа
		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
	})
	t.Run("check pwd policy", func(t *testing.T) {
		// считаем максимальный строк действия пароля из под гостя
		statusCode, pwdMaxLife, err := cl.GetKrbMaxPWDLife(t.Context())
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, statusCode)
		require.Zero(t, pwdMaxLife)

		statusCode, err = cl.Login(t.Context(), adminLogin, adminPass)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// считаем максимальный строк действия пароля
		statusCode, pwdMaxLife, err = cl.GetKrbMaxPWDLife(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
		require.Positive(t, pwdMaxLife)

		// создадим пользователя
		newUserID := funcs.RandStr()
		reqUser := RequestUser{
			UID:                   newUserID,
			GivenName:             newUserID + "-firstname",
			SN:                    newUserID + "-lastname",
			UserPassword:          funcs.Pointer(password1),
			KRBPasswordExpiration: funcs.Pointer(time.Now().AddDate(0, 3, 0)),
		}
		statusCode, _, err = cl.CreateUser(t.Context(), reqUser)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// выйдем из под админа
		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// зайдем под новым пользователем
		statusCode, err = cl.Login(t.Context(), newUserID, password1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// считаем максимальный строк действия пароля из пользователя
		statusCode, pwdMaxLife, err = cl.GetKrbMaxPWDLife(t.Context())
		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, statusCode) // password policy not found
		require.Zero(t, pwdMaxLife)

		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// зайдем под админом, удалим пользователя и выйдем
		statusCode, err = cl.Login(t.Context(), adminLogin, adminPass)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// удалим пользователя
		statusCode, err = cl.DeleteUser(t.Context(), newUserID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		statusCode, err = cl.Logout(t.Context())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)
	})
}
