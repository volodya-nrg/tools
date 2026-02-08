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

type config struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

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
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPass))

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
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPass))

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
		require.NoError(t, cl.Login(t.Context(), adminLogin, adminPass))

		roleName := funcs.RandStr()
		roleDesc := funcs.RandStr()

		// создадим роль
		require.NoError(t, cl.CreateRole(t.Context(), roleName, funcs.Pointer(roleDesc)))

		// получим роль
		role, err := cl.GetRole(t.Context(), roleName)
		require.NoError(t, err)
		require.Equal(t, roleName, role.CN)
		require.Equal(t, roleDesc, role.Description)

		// проверим явно что она есть
		isHas, err := cl.HasRole(t.Context(), roleName)
		require.NoError(t, err)
		require.True(t, isHas)

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

		roleNames := make([]string, len(roles))
		for i, roleLoc := range roles {
			roleNames[i] = roleLoc.CN
		}

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

		// проверим список по именам
		roles, err = cl.GetRolesByName(t.Context(), roleNames)
		require.NoError(t, err)
		require.Len(t, roles, len(roleNames))

		// получим по именам, но запросим левую роль, нужно чтоб response отработал корректно
		_, err = cl.GetRolesByName(t.Context(), []string{funcs.RandStr()})
		require.Error(t, err)

		// удалим роль
		require.NoError(t, cl.DeleteRole(t.Context(), roleName))

		// получим роль
		_, err = cl.GetRole(t.Context(), roleName)
		require.Error(t, err)

		// проверим через другой метод
		isHas, err = cl.HasRole(t.Context(), roleName)
		require.NoError(t, err)
		require.False(t, isHas)

		// выйдем из под админа
		require.NoError(t, cl.Logout(t.Context()))
	})
}
