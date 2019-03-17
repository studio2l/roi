package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// shotHandler는 /shot/ 하위 페이지로 사용자가 접속했을때 샷 정보가 담긴 페이지를 반환한다.
func shotHandler(w http.ResponseWriter, r *http.Request) {
	pth := r.URL.Path[len("/shot/"):]
	pths := strings.Split(pth, "/")
	if len(pths) != 2 {
		http.NotFound(w, r)
		return
	}
	prj := pths[0]
	shot := pths[1]
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	s, err := roi.GetShot(db, prj, shot)
	if err != nil {
		log.Fatal(err)
	}
	if s.ID == "" {
		http.NotFound(w, r)
		return
	}
	tasks, err := roi.AllTasks(db, prj, s.ID)
	if err != nil {
		log.Printf("could not get all tasks of shot '%s'", s.ID)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	recipt := struct {
		LoggedInUser string
		Project      string
		Shot         *roi.Shot
		Tasks        []*roi.Task
	}{
		LoggedInUser: session["userid"],
		Project:      prj,
		Shot:         s,
		Tasks:        tasks,
	}
	err = executeTemplate(w, "shot.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

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
	prj := r.Form.Get("project_id")
	if prj == "" {
		// 할일: 현재 GUI 디자인으로는 프로젝트를 선택하기 어렵기 때문에
		// 일단 첫번째 프로젝트로 이동한다. 나중에는 에러가 나야 한다.
		// 관련 이슈: #143
		prjRows, err := db.Query("SELECT id FROM projects")
		if err != nil {
			log.Print("could not select the first project:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer prjRows.Close()
		if !prjRows.Next() {
			fmt.Fprintf(w, "no projects in roi yet")
			return
		}
		if err := prjRows.Scan(&prj); err != nil {
			log.Printf("could not scan a row of project '%s': %v", prj, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/add-shot/?project_id="+prj, http.StatusSeeOther)
		return
	}
	p, err := roi.GetProject(db, prj)
	if err != nil {
		log.Printf("could not get project '%s': %v", prj, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if p == nil {
		msg := fmt.Sprintf("project '%s' not exist", prj)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		shot := r.Form.Get("id")
		if shot == "" {
			http.Error(w, "need shot 'id'", http.StatusBadRequest)
			return
		}
		exist, err := roi.ShotExist(db, prj, shot)
		if err != nil {
			log.Printf("could not check shot '%s' exist", shot)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, "shot '%s' already exist", http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"), ",")
		s := &roi.Shot{
			ID:            shot,
			ProjectID:     prj,
			Status:        roi.ShotWaiting,
			EditOrder:     atoi(r.Form.Get("edit_order")),
			Description:   r.Form.Get("description"),
			CGDescription: r.Form.Get("cg_description"),
			TimecodeIn:    r.Form.Get("timecode_in"),
			TimecodeOut:   r.Form.Get("timecode_out"),
			Duration:      atoi(r.Form.Get("duration")),
			Tags:          fields(r.Form.Get("tags"), ","),
			WorkingTasks:  tasks,
		}
		err = roi.AddShot(db, prj, s)
		if err != nil {
			log.Printf("could not add shot '%s': %v", prj+"."+shot, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		for _, task := range tasks {
			t := &roi.Task{
				ProjectID: prj,
				ShotID:    shot,
				Name:      task,
			}
			roi.AddTask(db, prj, shot, t)
		}
		http.Redirect(w, r, fmt.Sprintf("/shot/%s/%s", prj, shot), http.StatusSeeOther)
		return
	}
	recipt := struct {
		LoggedInUser string
		Project      *roi.Project
	}{
		LoggedInUser: session["userid"],
		Project:      p,
	}
	err = executeTemplate(w, "add-shot.html", recipt)
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
	prj := r.Form.Get("project_id")
	if prj == "" {
		http.Error(w, "need 'project_id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", prj), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("id")
	if shot == "" {
		http.Error(w, "need 'id'", http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		exist, err = roi.ShotExist(db, prj, shot)
		if err != nil {
			log.Print(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"), ",")
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
			Tags:          fields(r.Form.Get("tags"), ","),
			WorkingTasks:  tasks,
			DueDate:       tforms["due_date"],
		}
		err = roi.UpdateShot(db, prj, shot, upd)
		if err != nil {
			log.Print(err)
			http.Error(w, fmt.Sprintf("could not update shot '%s'", shot), http.StatusInternalServerError)
			return
		}
		// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
		for _, task := range tasks {
			t := &roi.Task{
				ProjectID: prj,
				ShotID:    shot,
				Name:      task,
				Status:    roi.TaskNotSet,
				DueDate:   time.Time{},
			}
			tid := prj + "." + shot + "." + task
			exist, err := roi.TaskExist(db, prj, shot, task)
			if err != nil {
				log.Printf("could not check task '%s' exist: %v", tid, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !exist {
				err := roi.AddTask(db, prj, shot, t)
				if err != nil {
					log.Printf("could not add task '%s': %v", tid, err)
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/shot/%s/%s", prj, shot), http.StatusSeeOther)
		return
	}
	s, err := roi.GetShot(db, prj, shot)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
		return
	}
	ts, err := roi.AllTasks(db, prj, shot)
	if err != nil {
		log.Printf("could not get all tasks of shot '%s': %v", prj+"."+shot, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Name] = t
	}
	recipt := struct {
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
	err = executeTemplate(w, "update-shot.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}
