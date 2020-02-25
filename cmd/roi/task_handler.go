package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/studio2l/roi"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateTaskPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	t, err := roi.GetTask(DB, id)
	if err != nil {
		return err
	}
	vers, err := roi.TaskVersions(DB, id)
	if err != nil {
		return err
	}
	us, err := roi.Users(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Task          *roi.Task
		AllTaskStatus []roi.TaskStatus
		Versions      []*roi.Version
		Users         []*roi.User
	}{
		LoggedInUser:  env.User.ID,
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
		Users:         us,
	}
	return executeTemplate(w, "update-task.html", recipe)
}

func updateTaskPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	tforms, err := parseTimeForms(r.Form, "due_date")
	if err != nil {
		return err
	}
	assignee := r.FormValue("assignee")
	if assignee != "" {
		_, err = roi.GetUser(DB, assignee)
		if err != nil {
			return err
		}
	}
	t, err := roi.GetTask(DB, id)
	if err != nil {
		return err
	}
	t.Status = roi.TaskStatus(r.FormValue("status"))
	t.Assignee = assignee
	t.DueDate = tforms["due_date"]
	t.PublishVersion = r.FormValue("publish_version")
	t.WorkingVersion = r.FormValue("working_version")

	err = roi.UpdateTask(DB, id, t)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateMultiTasksHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.FormValue("post") != "" {
		// 많은 샷 선택시 URL이 너무 길어져 잘릴 염려 때문에 GET을 사용하지 않아,
		// POST와 GET을 구분할 다른 방법이 필요했다. 더 나은 방법을 생각해 볼 것.
		return updateMultiTasksPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	ids := r.Form["id"]
	id := ids[0]
	show, ctg, _, err := roi.SplitUnitID(id)
	if err != nil {
		return err
	}
	site, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	tasks := site.ShotTasks
	if ctg == "asset" {
		tasks = site.AssetTasks
	}
	recipe := struct {
		LoggedInUser  string
		Show          string
		IDs           []string
		Tasks         []string
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  env.User.ID,
		Show:          show,
		IDs:           ids,
		Tasks:         tasks,
		AllTaskStatus: roi.AllTaskStatus,
	}
	return executeTemplate(w, "update-multi-tasks.html", recipe)
}

func updateMultiTasksPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	// 샷 아이디
	ids := r.Form["id"]
	tforms, err := parseTimeForms(r.Form, "due_date")
	if err != nil {
		return err
	}
	task := r.FormValue("task")
	dueDate := tforms["due_date"]
	status := r.FormValue("status")
	assignee := r.FormValue("assignee")
	for _, id := range ids {
		s, err := roi.GetTask(DB, id+"/"+task)
		if err != nil {
			if errors.As(err, &roi.NotFoundError{}) {
				// 여러 샷의 태스크를 한꺼번에 처리할 때는 어떤 샷에는
				// 해당 태스크가 없을수도 있다.
				// 이럴 경우 에러를 내지 않기로 한다.
				continue
			}
			return err
		}
		if !dueDate.IsZero() {
			s.DueDate = dueDate
		}
		if status != "" {
			s.Status = roi.TaskStatus(status)
		}
		if assignee != "" {
			s.Assignee = assignee
		}
		roi.UpdateTask(DB, id+"/"+task, s)
	}
	q := ""
	for i, id := range ids {
		if i != 0 {
			q += " "
		}
		shot := strings.Split(id, "/")[1]
		q += shot
	}
	// 여러 샷 수정 페이지 전인 shots 페이지로 돌아간다.
	return executeTemplate(w, "history-go.html", -2)
}
