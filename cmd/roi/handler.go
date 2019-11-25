package main

import (
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
			return httpError{msg: fmt.Sprintf("form field not found: %s", k), code: http.StatusBadRequest}
		}
	}
	return nil
}

func getSessionUser(r *http.Request) (*roi.User, error) {
	session, err := getSession(r)
	if err != nil {
		return nil, httpError{msg: "could not get session", code: http.StatusUnauthorized}
	}
	user := session["userid"]
	u, err := roi.GetUser(DB, user)
	if err != nil {
		return nil, httpError{msg: "could not get user information", code: http.StatusInternalServerError}
	}
	if u == nil {
		return nil, httpError{msg: fmt.Sprintf("user not exist: %s", user), code: http.StatusBadRequest}
	}
	return u, nil
}

func saveFormFile(r *http.Request, field string, dst string) error {
	f, fi, err := r.FormFile(field)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil
		}
		return err
	}
	defer f.Close()
	if fi.Size > (32 << 20) {
		return httpError{msg: fmt.Sprintf("mov: file size too big (got %dMB, maximum 32MB)", fi.Size>>20), code: http.StatusBadRequest}
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return httpError{msg: "could not read file data", code: http.StatusInternalServerError}
	}
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return httpError{msg: fmt.Sprintf("could not create directory: %v", err), code: http.StatusInternalServerError}
	}
	err = ioutil.WriteFile(dst, data, 0755)
	if err != nil {
		return httpError{msg: fmt.Sprintf("could not save file: %v", err), code: http.StatusInternalServerError}
	}
	return nil
}
