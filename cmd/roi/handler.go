package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/studio2l/roi"
)

type Env struct {
	SessionUser *roi.User
}

// HandlerFunc는 이 패키지에서 사용하는 핸들 함수이다.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, env *Env) error

// handle은 이 패키지에서 사용하는 핸들 함수를 http.HandleFunc로 변경한다.
func handle(serve HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *roi.User
		if !(r.URL.Path == "/login" || r.URL.Path == "/login/" || r.URL.Path == "/signup") {
			var err error
			u, err = sessionUser(r)
			if err != nil {
				if errors.As(err, &roi.NotFoundError{}) {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				clearSession(w)
				handleError(w, err)
				return
			}
		}
		env := &Env{
			SessionUser: u,
		}
		err := serve(w, r, env)
		if err != nil {
			handleError(w, err)
		}
	}
}

// handleError는 handle에서 요청을 처리하던 도중 에러가 났을 때 에러 메시지를 답신한다.
func handleError(w http.ResponseWriter, err error) {
	var e roi.Error
	if errors.As(err, &e) {
		if e.Log() != "" {
			log.Print(e.Log())
		}
		http.Error(w, err.Error(), e.Code())
		return
	}
	log.Print(err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

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
		return nil, fmt.Errorf("could not get session: %w", err)
	}
	user := session["userid"]
	if user == "" {
		return nil, roi.NotFound("user", "")
	}
	u, err := roi.GetUser(DB, user)
	if err != nil {
		if errors.As(err, &roi.NotFoundError{}) {
			// 일반적으로 db에 사용자가 없는 것은 NotFound 에러를 내지만,
			// 존재하지 않는 사용자가 세션 유저로 등록되어 있는 것은 해킹일 가능성이 높다.
			// 로그에 남도록 Internal 에러를 내고 %v 포매팅을 사용해 NotFound 타입정보는 지운다.
			return nil, fmt.Errorf("warn: invalid session user (malicious attack?): %s: %v", user, err)
		}
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
		return fmt.Errorf("could not read file data: %w", err)
	}
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	err = ioutil.WriteFile(dst, data, 0755)
	if err != nil {
		return fmt.Errorf("could not save file: %w", err)
	}
	return nil
}
