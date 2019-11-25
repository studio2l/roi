package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

	if r.Method == "POST" {
		err = r.ParseMultipartForm(1 << 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		version := r.Form.Get("version")
		if version == "" {
			http.Error(w, "need 'version'", http.StatusBadRequest)
			return
		}

		files := fields(r.Form.Get("output_files"))
		work_file := r.Form.Get("work_file")
		created, err := timeFromString(r.Form.Get("created"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid 'create' field: %v", r.Form.Get("created")), http.StatusBadRequest)
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
		src, mov, err := r.FormFile("mov")
		if err != nil {
			if err != http.ErrMissingFile {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			defer src.Close()
			if mov.Size > (32 << 20) {
				http.Error(w, fmt.Sprintf("mov: file size too big (got %dMB, maximum 32MB)", mov.Size>>20), http.StatusBadRequest)
				return
			}
			data, err := ioutil.ReadAll(src)
			if err != nil {
				http.Error(w, "could not get mov file data", http.StatusBadRequest)
				return
			}
			err = os.MkdirAll(fmt.Sprintf("data/show/%s/%s/%s/%s", show, shot, task, version), 0755)
			if err != nil {
				log.Printf("could not create mov directory: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			dst := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
			err = ioutil.WriteFile(dst, data, 0755)
			if err != nil {
				log.Printf("could not save mov file data: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
		v := &roi.Version{
			Show:        show,
			Shot:        shot,
			Task:        task,
			Version:     version,
			OutputFiles: files,
			WorkFile:    work_file,
			Created:     created,
		}
		err = roi.AddVersion(db, show, shot, task, v)
		if err != nil {
			log.Printf("could not add version to task '%s': %v", taskID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/update-version?show=%s&shot=%s&task=%s&version=%s", show, shot, task, version), http.StatusSeeOther)
		return
	}
	now := time.Now()
	v := &roi.Version{
		Show:    show,
		Shot:    shot,
		Task:    task,
		Created: now,
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "add",
		LoggedInUser: session["userid"],
		Version:      v,
	}
	err = executeTemplate(w, "update-version.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
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
	version := r.Form.Get("version")
	if version == "" {
		http.Error(w, "need 'version'", http.StatusBadRequest)
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
		err = r.ParseMultipartForm(1 << 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
		src, mov, err := r.FormFile("mov")
		if err != nil {
			if err != http.ErrMissingFile {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			defer src.Close()
			if mov.Size > (32 << 20) {
				http.Error(w, fmt.Sprintf("mov: file size too big (got %dMB, maximum 32MB)", mov.Size>>20), http.StatusBadRequest)
				return
			}
			data, err := ioutil.ReadAll(src)
			if err != nil {
				http.Error(w, "could not get mov file data", http.StatusBadRequest)
				return
			}
			err = os.MkdirAll(fmt.Sprintf("data/show/%s/%s/%s/%s", show, shot, task, version), 0755)
			if err != nil {
				log.Printf("could not create mov directory: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			dst := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
			err = ioutil.WriteFile(dst, data, 0755)
			if err != nil {
				log.Printf("could not save mov file data: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
		u := roi.UpdateVersionParam{
			OutputFiles: fields(r.Form.Get("output_files")),
			Images:      fields(r.Form.Get("images")),
			WorkFile:    r.Form.Get("work_file"),
			Created:     timeForms["created"],
		}
		roi.UpdateVersion(db, show, shot, task, version, u)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	v, err := roi.GetVersion(db, show, shot, task, version)
	if err != nil {
		log.Printf("could not get version '%s': %v", versionID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if v == nil {
		http.Error(w, fmt.Sprintf("version '%s' not exist", versionID), http.StatusBadRequest)
		return
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "update",
		LoggedInUser: session["userid"],
		Version:      v,
	}
	err = executeTemplate(w, "update-version.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
	return
}
