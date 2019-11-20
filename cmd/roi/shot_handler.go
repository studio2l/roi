package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

func addShotHandler(w http.ResponseWriter, r *http.Request) {
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
	// 어떤 프로젝트에 샷을 생성해야 하는지 체크.
	show := r.Form.Get("show")
	if show == "" {
		// 할일: 현재 GUI 디자인으로는 프로젝트를 선택하기 어렵기 때문에
		// 일단 첫번째 프로젝트로 이동한다. 나중에는 에러가 나야 한다.
		// 관련 이슈: #143
		showRows, err := db.Query("SELECT show FROM shows")
		if err != nil {
			log.Print("could not select the first show:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer showRows.Close()
		if !showRows.Next() {
			fmt.Fprintf(w, "no shows in roi yet")
			return
		}
		if err := showRows.Scan(&show); err != nil {
			log.Printf("could not scan a row of show '%s': %v", show, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/add-shot?show="+show, http.StatusSeeOther)
		return
	}
	sw, err := roi.GetShow(db, show)
	if err != nil {
		log.Printf("could not get show '%s': %v", show, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if sw == nil {
		msg := fmt.Sprintf("show '%s' not exist", show)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		shot := r.Form.Get("shot")
		if shot == "" {
			http.Error(w, "need 'shot'", http.StatusBadRequest)
			return
		}
		exist, err := roi.ShotExist(db, show, shot)
		if err != nil {
			log.Printf("could not check shot '%s' exist", shot)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, "shot '%s' already exist", http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"))
		s := &roi.Shot{
			Shot:          shot,
			Show:          show,
			Status:        roi.ShotWaiting,
			EditOrder:     atoi(r.Form.Get("edit_order")),
			Description:   r.Form.Get("description"),
			CGDescription: r.Form.Get("cg_description"),
			TimecodeIn:    r.Form.Get("timecode_in"),
			TimecodeOut:   r.Form.Get("timecode_out"),
			Duration:      atoi(r.Form.Get("duration")),
			Tags:          fields(r.Form.Get("tags")),
			WorkingTasks:  tasks,
		}
		err = roi.AddShot(db, show, s)
		if err != nil {
			log.Printf("could not add shot '%s': %v", show+"."+shot, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		for _, task := range tasks {
			t := &roi.Task{
				Show:    show,
				Shot:    shot,
				Task:    task,
				Status:  roi.TaskNotSet,
				DueDate: time.Time{},
			}
			roi.AddTask(db, show, shot, t)
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	recipe := struct {
		LoggedInUser string
		Show         *roi.Show
	}{
		LoggedInUser: session["userid"],
		Show:         sw,
	}
	err = executeTemplate(w, "add-shot.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}

func updateShotHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		exist, err = roi.ShotExist(db, show, shot)
		if err != nil {
			log.Print(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"))
		tforms, err := parseTimeForms(r.Form, "due_date")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		upd := roi.UpdateShotParam{
			Status:        roi.ShotStatus(r.Form.Get("status")),
			EditOrder:     atoi(r.Form.Get("edit_order")),
			Description:   r.Form.Get("description"),
			CGDescription: r.Form.Get("cg_description"),
			TimecodeIn:    r.Form.Get("timecode_in"),
			TimecodeOut:   r.Form.Get("timecode_out"),
			Duration:      atoi(r.Form.Get("duration")),
			Tags:          fields(r.Form.Get("tags")),
			WorkingTasks:  tasks,
			DueDate:       tforms["due_date"],
		}
		err = roi.UpdateShot(db, show, shot, upd)
		if err != nil {
			log.Print(err)
			http.Error(w, fmt.Sprintf("could not update shot '%s'", shot), http.StatusInternalServerError)
			return
		}
		// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
		for _, task := range tasks {
			t := &roi.Task{
				Show:    show,
				Shot:    shot,
				Task:    task,
				Status:  roi.TaskNotSet,
				DueDate: time.Time{},
			}
			tid := show + "." + shot + "." + task
			exist, err := roi.TaskExist(db, show, shot, task)
			if err != nil {
				log.Printf("could not check task '%s' exist: %v", tid, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !exist {
				err := roi.AddTask(db, show, shot, t)
				if err != nil {
					log.Printf("could not add task '%s': %v", tid, err)
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
			}
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	s, err := roi.GetShot(db, show, shot)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
		return
	}
	ts, err := roi.ShotTasks(db, show, shot)
	if err != nil {
		log.Printf("could not get all tasks of shot '%s': %v", show+"."+shot, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Task] = t
	}
	recipe := struct {
		LoggedInUser  string
		Shot          *roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  session["userid"],
		Shot:          s,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tm,
		AllTaskStatus: roi.AllTaskStatus,
	}
	err = executeTemplate(w, "update-shot.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
