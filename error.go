package roi

import "net/http"

// Error는 error 인터페이스를 만족하는 roi의 에러 타입이다.
// roi의 모든 에러는 Error 형이어야 한다.
type Error interface {
	Error() string
	Code() int
	Log() string
}

// NotFoundError는 로이에서 특정 항목을 검색했지만 해당 항목이 없음을 의미하는 에러이다.
type NotFoundError struct {
	kind string
	id   string
}

// NotFound는 NotFoundError를 반환한다.
func NotFound(kind, id string) NotFoundError {
	return NotFoundError{kind: kind, id: id}
}

func (e NotFoundError) Error() string {
	return e.kind + " not found: " + e.id
}

func (e NotFoundError) Code() int {
	return http.StatusNotFound
}

func (e NotFoundError) Log() string {
	return ""
}

// BadRequestError는 로이의 함수를 호출했지만 그와 관련된 정보가 잘못되었음을 의미하는 에러이다.
type BadRequestError struct {
	msg string
}

// BadRequest는 BadRequestError 에러를 반환한다.
func BadRequest(msg string) BadRequestError {
	return BadRequestError{msg: msg}
}

func (e BadRequestError) Error() string {
	return e.msg
}

func (e BadRequestError) Code() int {
	return http.StatusBadRequest
}

func (e BadRequestError) Log() string {
	return ""
}

// InternalError은 문제가 서버 밖으로 전달되지 않아야 하는 때 반환되는 에러이다.
// InternalError은 errors.Wrapper 인터페이스를 만족한다.
type InternalError struct {
	err error
}

// Internal은 InternalError 에러를 반환한다.
func Internal(err error) InternalError {
	return InternalError{err}
}

func (e InternalError) Error() string {
	return "internal error"
}

func (e InternalError) Code() int {
	return http.StatusInternalServerError
}

func (e InternalError) Log() string {
	return e.err.Error()
}

func (e InternalError) Unwrap() error {
	return e.err
}

// AuthError는 특정 사용자가 허락되지 않은 행동을 요청했을 때 반환되는 에러이다.
type AuthError struct {
	user string
	op   string
}

func Auth(user, op string) AuthError {
	return AuthError{user: user, op: op}
}

func (e AuthError) Error() string {
	return e.user + " has no right to do " + e.op
}

func (e AuthError) Code() int {
	return http.StatusUnauthorized
}

func (e AuthError) Log() string {
	return e.user + " tried to do " + e.op
}
