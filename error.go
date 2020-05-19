package roi

import (
	"fmt"
)

// NotFoundError는 로이에서 특정 항목을 검색했지만 해당 항목이 없음을 의미하는 에러이다.
type NotFoundError struct {
	err error
}

// NotFound는 NotFoundError를 반환한다.
func NotFound(msg string, vals ...interface{}) NotFoundError {
	return NotFoundError{err: fmt.Errorf(msg, vals...)}
}

func (e NotFoundError) Error() string {
	return e.err.Error()
}

func (e NotFoundError) Unwrap() error {
	return e.err
}

// BadRequestError는 로이의 함수를 호출했지만 그와 관련된 정보가 잘못되었음을 의미하는 에러이다.
type BadRequestError struct {
	err error
}

// BadRequest는 BadRequestError를 반환한다.
func BadRequest(msg string, vals ...interface{}) BadRequestError {
	return BadRequestError{err: fmt.Errorf(msg, vals...)}
}

func (e BadRequestError) Error() string {
	return e.err.Error()
}

func (e BadRequestError) Unwrap() error {
	return e.err
}

// AuthError는 특정 사용자가 허락되지 않은 행동을 요청했음을 의미하는 에러이다.
type AuthError struct {
	err error
}

// Auth는 AuthError를 반환한다.
func Auth(msg string, vals ...interface{}) AuthError {
	return AuthError{fmt.Errorf(msg, vals...)}
}

func (e AuthError) Error() string {
	return e.err.Error()
}

func (e AuthError) Unwrap() error {
	return e.err
}
