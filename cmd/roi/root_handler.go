package main

import (
	"net/http"
)

// rootHandler는 루트 페이지로 사용자가 접근했을때 그 사용자에게 필요한 정보를 맞춤식으로 제공한다.
func rootHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.URL.Path != "/" {
		// 정의되지 않은 페이지로의 이동을 차단
		http.Error(w, "page not found", http.StatusNotFound)
		return nil
	}
	http.Redirect(w, r, "/user/"+env.User.ID, http.StatusSeeOther)
	return nil
}
