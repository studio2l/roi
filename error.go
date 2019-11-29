package roi

import "net/http"

// Error는 error 인터페이스를 만족하는 roi의 에러 타입이다.
// roi의 모든 에러는 Error 형이어야 한다.
type Error interface {
	Error() string
	Code() int
	Log() string
}

// NotFound는 로이에서 특정 항목을 검색했지만 해당 항목이 없을 때 반환되는 에러이다.
type NotFound struct {
	Kind string
	ID   string
}

func (e NotFound) Error() string {
	return e.Kind + " not found: " + e.ID
}

func (e NotFound) Code() int {
	return http.StatusNotFound
}

func (e NotFound) Log() string {
	return ""
}

// BadRequest는 로이의 함수를 호출했지만 그와 관련된 정보가 부족하거나 잘못되었을 때 반환되는 에러이다.
type BadRequest struct {
	Msg string
}

func (e BadRequest) Error() string {
	return e.Msg
}

func (e BadRequest) Code() int {
	return http.StatusBadRequest
}

func (e BadRequest) Log() string {
	return ""
}

// Internal은 문제가 서버 밖으로 전달되지 않아야 하는 때 반환되는 에러이다.
// Internal은 errors.Wrapper 인터페이스를 만족한다.
type Internal struct {
	Err error
}

func (e Internal) Error() string {
	return "internal error"
}

func (e Internal) Code() int {
	return http.StatusInternalServerError
}

func (e Internal) Log() string {
	return e.Err.Error()
}

func (e Internal) Unwrap() error {
	return e.Err
}

// Auth는 특정 사용자가 허락되지 않은 행동을 요청했을 때 반환되는 에러이다.
type Auth struct {
	User string
	Op   string
}

func (e Auth) Error() string {
	return e.User + " has no right to do " + e.Op
}

func (e Auth) Code() int {
	return http.StatusUnauthorized
}

func (e Auth) Log() string {
	return e.User + " tried to do " + e.Op
}
