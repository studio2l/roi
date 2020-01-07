package main

import (
	"fmt"
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
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
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
		LoggedInUser:  env.SessionUser.ID,
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
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
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

	err = roi.UpdateTask(DB, id, t)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateTaskWorkingVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	err = roi.UpdateTaskWorkingVersion(DB, id, version)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateTaskPublishVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	err = roi.UpdateTaskPublishVersion(DB, id, version)
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
	for _, id := range ids {
		err = roi.VerifyShotID(id)
		if err != nil {
			return err
		}
	}
	id := ids[0]
	show, _, err := roi.SplitShotID(id)
	if err != nil {
		return err
	}
	site, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Show          string
		IDs           []string
		Tasks         []string
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  env.SessionUser.ID,
		Show:          show,
		IDs:           ids,
		Tasks:         site.Tasks,
		AllTaskStatus: roi.AllTaskStatus,
	}
	return executeTemplate(w, "update-multi-tasks.html", recipe)
}

func updateMultiTasksPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
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
	show := strings.Split(ids[0], "/")[0]
	http.Redirect(w, r, fmt.Sprintf("/shots?show=%s&q=%s", show, q), http.StatusSeeOther)
	return nil
}
