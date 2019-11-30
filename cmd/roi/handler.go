package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/studio2l/roi"
)

func mustFields(r *http.Request, keys ...string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			return roi.BadRequest(fmt.Sprintf("form field not found: %s", k))
		}
	}
	return nil
}

// sessionUser는 현재 http 요청의 세션 쿠키에서 userid를 검사하여,
// 그 아이디에 해당하는 유저를 반환한다.
// 세션에 userid 정보가 없다면 nil 유저를 반환한다.
func sessionUser(r *http.Request) (*roi.User, error) {
	session, err := getSession(r)
	if err != nil {
		return nil, roi.Internal(fmt.Errorf("could not get session: %w", err))
	}
	user := session["userid"]
	if user == "" {
		return nil, roi.NotFound("user", "")
	}
	u, err := roi.GetUser(DB, user)
	if errors.As(err, &roi.NotFoundError{}) {
		// 일반적으로 db에 사용자가 없는 것은 NotFound 에러를 내지만,
		// 존재하지 않는 사용자가 세션 유저로 등록되어 있는 것은 해킹일 가능성이 높다.
		// 로그에 남도록 Internal 에러를 내고 %v 포매팅을 사용해 NotFound 타입정보는 지운다.
		return nil, roi.Internal(fmt.Errorf("warn: invalid session user (malicious attack?): %s: %v", user, err))
	}
	return u, nil
}

func saveFormFile(r *http.Request, field string, dst string) error {
	f, fi, err := r.FormFile(field)
	if err != nil {
		if err == http.ErrMissingFile {
			// 사용자가 파일을 업로드 하지 않은 것은 에러가 아니다.
			return nil
		}
		return err
	}
	defer f.Close()
	if fi.Size > (32 << 20) {
		return roi.BadRequest(fmt.Sprintf("mov: file size too big (got %dMB, maximum 32MB)", fi.Size>>20))
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return roi.Internal(fmt.Errorf("could not read file data: %w", err))
	}
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return roi.Internal(fmt.Errorf("could not create directory: %w", err))
	}
	err = ioutil.WriteFile(dst, data, 0755)
	if err != nil {
		return roi.Internal(fmt.Errorf("could not save file: %w", err))
	}
	return nil
}
