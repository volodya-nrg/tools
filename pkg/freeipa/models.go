package freeipa

import "time"

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
	GivenName             string
	SN                    string
	DN                    string
	MemberOfGroup         []string
	MemberOfRole          []string
	Mail                  string
	NsAccountLock         bool
	KRBPasswordExpiration time.Time
}

type RequestUser struct {
	UID                   string
	GivenName             string
	SN                    string
	Mail                  *string
	UserPassword          *string
	KRBPasswordExpiration *time.Time
	NsAccountLock         *bool
}
