package custom

import "google.golang.org/grpc/codes"

type CustomError struct {
	err      error
	grpcCode codes.Code
	newMsg   string
}

func (m CustomError) Error() string {
	var result string

	if m.newMsg != "" {
		result = m.newMsg
	} else if m.err != nil {
		result = m.err.Error()
	}

	if result == "" {
		result = "undefined error"
	}

	return result
}

func (m CustomError) GetCode() codes.Code {
	return m.grpcCode
}

func (m CustomError) Unwrap() error {
	return m.err
}

// NewCustomError err может быть nil
func NewCustomError(err error, grpcCode codes.Code, newMsg string) *CustomError {
	return &CustomError{
		err:      err,
		grpcCode: grpcCode,
		newMsg:   newMsg,
	}
}
