package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/studio2l/roi"
)

type Env struct {
	User *roi.User
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
			User: u,
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
			return roi.BadRequest("form field not found: %s", k)
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
		return nil, roi.NotFound("user not set in session")
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

func saveImageFormFile(r *http.Request, field string, dst string) error {
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
		return roi.BadRequest("mov: file size too big (got %dMB, maximum 32MB)", fi.Size>>20)
	}
	// 어떤 타입의 이미지를 업로드하든 png로 변경한다.
	img, _, err := image.Decode(f)
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return roi.BadRequest(err.Error())
		}
		return fmt.Errorf("could not read image data: %w", err)
	}
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	dstf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstf.Close()
	if err := png.Encode(dstf, img); err != nil {
		return err
	}
	return nil
}

// saveFormFiles는 리퀘스트 멀티파트로 전송되어온 파일을 dstd 디렉토리에 저장한다.
func saveFormFiles(r *http.Request, field string, dstd string) error {
	r.ParseMultipartForm(200000) // 사용하는 최대 메모리 사이즈: 200KB
	fileHeaders := r.MultipartForm.File[field]
	if len(fileHeaders) == 0 {
		return nil
	}
	err := os.MkdirAll(dstd, 0755)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	for _, fh := range fileHeaders {
		err := saveMultipartFile(fh, filepath.Join(dstd, fh.Filename))
		if err != nil {
			return err
		}
	}
	return nil
}

// saveMultipartFile은 멀티파트 파일을 dstf에 저장한다.
func saveMultipartFile(f *multipart.FileHeader, dstf string) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstf)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

// formValues는 http 요청에서 슬라이스 형식의 필드값을 가지고 온다.
// 그 중 비어있는 값은 슬라이스에서 빠진다.
// formValues는 요청의 폼 정보를 자동으로 파싱한다.
func formValues(r *http.Request, field string) []string {
	r.ParseForm()
	vals := make([]string, 0)
	for _, v := range r.Form[field] {
		if v != "" {
			vals = append(vals, v)
		}
	}
	return vals
}
