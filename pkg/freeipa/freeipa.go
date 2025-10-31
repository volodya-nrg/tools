package freeipa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/volodya-nrg/tools/pkg/funcs"
)

const (
	limitDefault                = 20
	timeLayout                  = "20060102150405Z"
	apiVersion                  = "2.254"
	keyOptGivenName             = "givenname"
	keyOptSN                    = "sn"
	keyOptUser                  = "user"
	keyOptUserPassword          = "userpassword"
	keyOptKRBPasswordExpiration = "krbpasswordexpiration"
	keyOptMail                  = "mail"
	keyOptNSAccountLock         = "nsaccountlock"
	keyOptRandom                = "random"
	keyOptVersion               = "version"
	keyOptDescription           = "description"
)

type FreeIPA struct {
	scheme     string
	host       string
	client     *http.Client
	apiVersion string
}

func (f *FreeIPA) Close() error {
	f.client.CloseIdleConnections()
	return nil
}

// Login специальный/отдельный запрос на аутентификацию (не jsonRPC)
func (f *FreeIPA) Login(ctx context.Context, userID, password string) error {
	values := url.Values{
		"user":     []string{userID},
		"password": []string{password},
	}
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/login_password",
	}
	headers := map[string]string{
		"Referer": fmt.Sprintf("%s://%s/ipa", f.scheme, f.host),
	}

	if err := f.httpRequest(ctx, f.client, http.MethodPost, u, []byte(values.Encode()), headers, nil); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}

	return nil
}

func (f *FreeIPA) Logout(ctx context.Context) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}

	req, err := f.rpcReq("session_logout", "", nil, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to logout: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

// users

// GetUsers получение пользователей
func (f *FreeIPA) GetUsers(ctx context.Context, limit, offset int32) ([]User, uint32, error) {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		"pkey_only": true,
	}

	req, err := f.rpcReq("user_find", "", opts, true)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create jsonrpc-request (user_find): %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("failed to get users (pkey_only): code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, 0, errors.New("response result is nil")
	}

	users := make([]User, 0)
	total := resp.Result.Count

	if usersList, ok := resp.Result.Result.([]any); ok {
		users = make([]User, 0, len(usersList))
		for _, user := range usersList {
			if userTmp, ok := user.(map[string]any); ok {
				users = append(users, mapUserToDTOUser(userTmp))
			}
		}
	}

	targetUsers := getRangeFromSlice(users, limit, offset, limitDefault)
	methods := make([]string, len(targetUsers))
	opts = map[string]any{
		"all":        true, // получить полную информацию о пользователях
		"no_members": true, // исключить информацию о группах
	}

	for i, user := range targetUsers {
		method, err := f.rpcReq("user_show", fmt.Sprintf(`["%s"]`, user.UID), opts, false)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create jsonrpc-request (user_show): %w", err)
		}

		methods[i] = string(method)
	}

	req, err = f.rpcReq("batch", fmt.Sprintf(`[%s]`, strings.Join(methods, ",")), nil, true)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create jsonrpc-request (batch): %w", err)
	}

	resp = responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("failed to get users: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, 0, errors.New("response result is nil")
	}

	users = make([]User, 0, len(resp.Result.Results))

	for _, result := range resp.Result.Results {
		if userTmp, ok := result.Result.(map[string]any); ok {
			users = append(users, mapUserToDTOUser(userTmp))
		}
	}

	return users, total, nil
}

func (f *FreeIPA) GetUser(ctx context.Context, userID string) (*User, error) {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		"all": true,
	}

	req, err := f.rpcReq("user_show", fmt.Sprintf(`["%s"]`, userID), opts, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, errors.New("response result is nil")
	}

	userTmp, ok := resp.Result.Result.(map[string]any)
	if !ok {
		return nil, errors.New("failed to parse response")
	}

	user := mapUserToDTOUser(userTmp)

	return &user, nil
}

func (f *FreeIPA) CreateUser(ctx context.Context, reqUser RequestUser) (*User, error) {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		keyOptGivenName: reqUser.GivenName,
		keyOptSN:        reqUser.SN,
	}

	if reqUser.UserPassword != nil {
		opts[keyOptUserPassword] = *reqUser.UserPassword
	} else {
		opts[keyOptRandom] = true
	}
	if reqUser.KRBPasswordExpiration != nil {
		opts[keyOptKRBPasswordExpiration] = reqUser.KRBPasswordExpiration.Format(timeLayout) // спец. формат
	}
	if reqUser.Mail != nil {
		opts[keyOptMail] = *reqUser.Mail
	}
	if reqUser.NsAccountLock != nil {
		opts[keyOptNSAccountLock] = *reqUser.NsAccountLock
	}

	req, err := f.rpcReq("user_add", fmt.Sprintf(`["%s"]`, reqUser.UID), opts, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("failed to create item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, errors.New("response result is nil")
	}

	userTmp, ok := resp.Result.Result.(map[string]any)
	if !ok {
		return nil, errors.New("failed to parse response")
	}

	user := mapUserToDTOUser(userTmp)

	return &user, nil
}

// UpdateUser тут лучше пользователя обратно не отдавать, т.к. он имеет не полные данные
func (f *FreeIPA) UpdateUser(ctx context.Context, reqUser RequestUser) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{}

	if reqUser.GivenName != "" {
		opts[keyOptGivenName] = reqUser.GivenName
	}
	if reqUser.SN != "" {
		opts[keyOptSN] = reqUser.SN
	}
	if reqUser.UserPassword != nil { // если пароль изменят, то время (KRBPasswordExpiration) не учитывается
		opts[keyOptUserPassword] = *reqUser.UserPassword
	}
	if reqUser.KRBPasswordExpiration != nil {
		opts[keyOptKRBPasswordExpiration] = reqUser.KRBPasswordExpiration.Format(timeLayout) // спец. формат
	}
	if reqUser.Mail != nil {
		opts[keyOptMail] = *reqUser.Mail
	}
	if reqUser.NsAccountLock != nil {
		opts[keyOptNSAccountLock] = *reqUser.NsAccountLock
	}

	req, err := f.rpcReq("user_mod", fmt.Sprintf(`["%s"]`, reqUser.UID), opts, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to update item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (f *FreeIPA) DeleteUser(ctx context.Context, userID string) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}

	req, err := f.rpcReq("user_del", fmt.Sprintf(`["%s"]`, userID), nil, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to delete item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

// roles

func (f *FreeIPA) GetRoles(ctx context.Context, limit, offset int32) ([]Role, uint32, error) {
	// TODO тут надо получить кол-во пользователей данной роли

	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		"pkey_only": true,
	}

	req, err := f.rpcReq("role_find", "", opts, true)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("failed to get roles: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, 0, errors.New("response result is nil")
	}

	rolesTmp, ok := resp.Result.Result.([]any)
	if !ok {
		return nil, 0, errors.New("failed to parse roles response")
	}

	roles := make([]Role, 0, len(rolesTmp))
	total := resp.Result.Count

	for _, v := range rolesTmp {
		v2, ok := v.(map[string]any)
		if !ok {
			return nil, 0, errors.New("failed to parse role response")
		}

		roles = append(roles, mapUserToDTORole(v2))
	}

	targetRoles := getRangeFromSlice(roles, limit, offset, limitDefault)
	methods := make([]string, len(targetRoles))
	opts = map[string]any{
		"all":        true, // получить полную информацию о роли
		"no_members": true, // исключить информацию о группах
	}

	for i, role := range targetRoles {
		method, err := f.rpcReq("role_show", fmt.Sprintf(`["%s"]`, role.CN), opts, false)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create jsonrpc-request (role_show): %w", err)
		}

		methods[i] = string(method)
	}

	req, err = f.rpcReq("batch", fmt.Sprintf(`[%s]`, strings.Join(methods, ",")), nil, true)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create jsonrpc-request (batch): %w", err)
	}

	resp = responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("failed to get roles: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, 0, errors.New("response result is nil")
	}

	roles = make([]Role, 0, len(resp.Result.Results))

	for _, result := range resp.Result.Results {
		if roleTmp, ok := result.Result.(map[string]any); ok {
			roles = append(roles, mapUserToDTORole(roleTmp))
		}
	}

	return roles, total, nil
}

func (f *FreeIPA) GetRole(ctx context.Context, name string) (*Role, error) {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		"all": true,
	}

	req, err := f.rpcReq("role_show", fmt.Sprintf(`["%s"]`, name), opts, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return nil, fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, errors.New("response result is nil")
	}

	roleTmp, ok := resp.Result.Result.(map[string]any)
	if !ok {
		return nil, errors.New("failed to parse response")
	}

	role := mapUserToDTORole(roleTmp)

	return &role, nil
}

func (f *FreeIPA) CreateRole(ctx context.Context, name string, desc *string) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{}

	if desc != nil {
		opts[keyOptDescription] = *desc
	}

	req, err := f.rpcReq("role_add", fmt.Sprintf(`["%s"]`, name), opts, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to create item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (f *FreeIPA) UpdateRole(ctx context.Context, name, desc string) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		keyOptDescription: desc,
	}

	req, err := f.rpcReq("role_mod", fmt.Sprintf(`["%s"]`, name), opts, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to update item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (f *FreeIPA) DeleteRole(ctx context.Context, name string) error {
	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}

	req, err := f.rpcReq("role_del", fmt.Sprintf(`["%s"]`, name), nil, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to delete item: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (f *FreeIPA) ToggleRoleForUser(ctx context.Context, roleName, userID string) error {
	user, err := f.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if slices.Contains(user.MemberOfRole, roleName) {
		return f.editRoleForUser(ctx, roleName, userID, true)
	} else {
		return f.editRoleForUser(ctx, roleName, userID, false)
	}
}

func (f *FreeIPA) editRoleForUser(ctx context.Context, roleName, userID string, isRemove bool) error {
	method := "role_add_member"

	if isRemove {
		method = "role_remove_member"
	}

	u := url.URL{
		Scheme: f.scheme,
		Host:   f.host,
		Path:   "ipa/session/json",
	}
	opts := map[string]any{
		keyOptUser: userID,
	}

	req, err := f.rpcReq(method, fmt.Sprintf(`["%s"]`, roleName), opts, true)
	if err != nil {
		return fmt.Errorf("failed to create jsonrpc-request: %w", err)
	}

	resp := responseBasic{}

	if err = f.httpRequest(ctx, f.client, http.MethodPost, u, req, f.headers(), &resp); err != nil {
		return fmt.Errorf("failed to http-request: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("failed to add role to user: code (%d), msg (%s)", resp.Error.Code, resp.Error.Message)
	}

	return nil
}

func (f *FreeIPA) headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"Referer":      fmt.Sprintf("%s://%s/ipa", f.scheme, f.host),
	}
}

// rpcReq формирует rpc-запрос.
// Если запрос делать через объект, после маршалить, то args обретают не ту структуру,
// поэтому формируем явно через строчку/байты. Маршалим только опции.
// ApiVersion в опциях необходим, чтоб в response не прилетало лишнего текста.
func (f *FreeIPA) rpcReq(method, argsSrc string, optsSrc map[string]any, isFull bool) ([]byte, error) {
	args := "[]"
	if argsSrc != "" {
		args = argsSrc
	}

	opts := map[string]any{}
	for k, v := range optsSrc {
		opts[k] = v
	}

	opts[keyOptVersion] = f.apiVersion
	dop := ""

	if isFull {
		dop = fmt.Sprintf(`"version":"2.0", "id":"%s", `, uuid.NewString())
	}

	optsBytes, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	result := fmt.Sprintf(`{%s"method":"%s", "params":[%s, %s]}`, dop, method, args, optsBytes)

	return []byte(result), nil
}

func (f *FreeIPA) httpRequest(
	ctx context.Context,
	client *http.Client,
	method string, //nolint:unparam
	u url.URL,
	body []byte,
	headers map[string]string,
	receiver any,
) error {
	return funcs.HTTPRequest(ctx, client, method, u, body, headers, receiver)
}

func NewFreeIPA(scheme, host string, timeout time.Duration) (*FreeIPA, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie-jar: %w", err)
	}

	return &FreeIPA{
		scheme: scheme,
		host:   host,
		client: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
		apiVersion: apiVersion,
	}, nil
}
