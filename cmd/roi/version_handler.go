package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/studio2l/roi"
)

func addVersionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	show := r.Form.Get("show")
	if show == "" {
		http.Error(w, "need 'show'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ShowExist(db, show)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("show '%s' not exist", show), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("shot")
	if shot == "" {
		http.Error(w, "need 'shot'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("task")
	if task == "" {
		http.Error(w, "need 'task'", http.StatusBadRequest)
		return
	}
	// addVersion은 새 버전을 추가하는 역할만 하고 값을 넣는 역할은 하지 않는다.
	// 만일 인수를 받았다면 에러를 낼 것.
	version := r.Form.Get("version")
	if version != "" {
		// 버전은 db에 기록된 마지막 버전을 기준으로 하지 여기서 받아들이지 않는다.
		http.Error(w, "'version' should not be specified", http.StatusBadRequest)
		return
	}
	if r.Form.Get("files") != "" {
		http.Error(w, "does not accept 'files'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("mov") != "" {
		http.Error(w, "does not accept 'mov'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("work_file") != "" {
		http.Error(w, "does not accept 'work_file'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("created") != "" {
		http.Error(w, "does not accept 'created'", http.StatusBadRequest)
		return
	}
	taskID := fmt.Sprintf("%s.%s.%s", show, shot, task)
	t, err := roi.GetTask(db, show, shot, task)
	if err != nil {
		log.Printf("could not get task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
		return
	}
	o := &roi.Version{
		Show: show,
		Shot: shot,
		Task: task,
	}
	err = roi.AddVersion(db, show, shot, task, o)
	if err != nil {
		log.Printf("could not add version to task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	show := r.Form.Get("show")
	if show == "" {
		http.Error(w, "need 'show'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ShowExist(db, show)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("show '%s' not exist", show), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("shot")
	if shot == "" {
		http.Error(w, "need 'shot'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("task")
	if task == "" {
		http.Error(w, "need 'task'", http.StatusBadRequest)
		return
	}
	v := r.Form.Get("version")
	if v == "" {
		http.Error(w, "need 'version'", http.StatusBadRequest)
		return
	}
	version, err := strconv.Atoi(v)
	if err != nil {
		http.Error(w, "'version' is not a number", http.StatusBadRequest)
		return
	}
	taskID := fmt.Sprintf("%s.%s.%s", show, shot, task)
	t, err := roi.GetTask(db, show, shot, task)
	if err != nil {
		log.Printf("could not get task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
		return
	}
	versionID := fmt.Sprintf("%s.%s.%s.v%v03d", show, shot, task, version)
	if r.Method == "POST" {
		exist, err := roi.VersionExist(db, show, shot, task, version)
		if err != nil {
			log.Printf("could not check version '%s' exist: %v", versionID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, "version '%s' not exist", http.StatusBadRequest)
			return
		}
		timeForms, err := parseTimeForms(r.Form,
			"created",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		u := roi.UpdateVersionParam{
			OutputFiles: fields(r.Form.Get("output_files"), ","),
			Images:      fields(r.Form.Get("images"), ","),
			Mov:         r.Form.Get("mov"),
			WorkFile:    r.Form.Get("work_file"),
			Created:     timeForms["created"],
		}
		roi.UpdateVersion(db, show, shot, task, version, u)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	o, err := roi.GetVersion(db, show, shot, task, version)
	if err != nil {
		log.Printf("could not get version '%s': %v", versionID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if o == nil {
		http.Error(w, fmt.Sprintf("version '%s' not exist", versionID), http.StatusBadRequest)
		return
	}
	recipt := struct {
		LoggedInUser string
		Version      *roi.Version
	}{
		LoggedInUser: session["userid"],
		Version:      o,
	}
	err = executeTemplate(w, "update-version.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
	return
}
