package freeipa

import "time"

type config struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type responseBasic struct {
	Result    *responseResult `json:"result"`
	Error     *responseError  `json:"error"`
	ID        string          `json:"id"`
	Principal string          `json:"principal"`
	Version   string          `json:"version"`
}

type responseError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
	Name    string            `json:"name"`
}

type responseResult struct {
	Result    any               `json:"result"`
	Results   []responseItem    `json:"results"`
	Messages  []responseMessage `json:"messages"`
	Count     uint32            `json:"count"`
	Truncated bool              `json:"truncated"` // пусть будет на всякий случай
	Summary   string            `json:"summary"`   // пусть будет на всякий случай
}

type responseMessage struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    struct {
		ServerVersion string `json:"server_version"`
	} `json:"data"`
}

type responseItem struct {
	Result  any     `json:"result"`
	Value   string  `json:"value"` // uid
	Summary *string `json:"summary"`
	// Error   *responseError `json:"error"`
	Error     string `json:"error"`
	ErrorCode int32  `json:"error_code"`
	ErrorName string `json:"error_name"`
	ErrorKw   struct {
		Reason string `json:"reason"`
	} `json:"error_kw"`
}

type Role struct {
	CN          string
	DN          string
	Description string
	ObjectClass []string
	MemberUser  []string
}

type User struct {
	UID                   string
	GivenName             string // имя
	SN                    string // фамилия
	DN                    string
	MemberOfGroup         []string
	MemberOfRole          []string
	Mail                  string
	NsAccountLock         bool
	KRBPasswordExpiration time.Time
	CN                    string // ФИО
	TelephoneNumber       string // рабочий телефон
	Mobile                string // мобильный телефон
	Title                 string // должность
	Organization          string // компания
	OrgUnit               string // отдел в компании
	JPEGPhoto             string // аватарка
}

type RequestUser struct {
	UID                   string     // id
	GivenName             string     // имя
	SN                    string     // фамилия
	Mail                  *string    // е-мэйл
	UserPassword          *string    // новый пароль
	KRBPasswordExpiration *time.Time // время действия пароля
	NsAccountLock         *bool      // заблокирован ли аккаунт
	CN                    *string    // fullname, FIO
	TelephoneNumber       *string    // рабочий телефон
	Mobile                *string    // мобильный телефон
	Title                 *string    // должность
	OU                    *string    // отдел (orgunit)
	AddAttr               []string   // доп. аттрибуты (компания, аватарка)
}
