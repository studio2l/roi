package main

import (
	"net/http"

	"github.com/gorilla/securecookie"
)

// cookieHandler는 클라이언트 브라우저 세션에 암호화된 쿠키를 저장을 돕는다.
var cookieHandler *securecookie.SecureCookie

// setSession은 클라이언트 브라우저에 세션을 저장한다.
func setSession(w http.ResponseWriter, session map[string]string) error {
	encoded, err := cookieHandler.Encode("session", session)
	if err != nil {
		return err
	}
	c := &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, c)
	return nil
}

// getSession은 클라이언트 브라우저에 저장되어 있던 세션을 불러온다.
func getSession(r *http.Request) (map[string]string, error) {
	c, _ := r.Cookie("session")
	if c == nil {
		return nil, nil
	}
	value := make(map[string]string)
	err := cookieHandler.Decode("session", c.Value, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// clearSession은 클라이언트 브라우저에 저장되어 있던 세션을 지운다.
func clearSession(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
}
