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
	show, grp, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	vers, err := roi.TaskVersions(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	us, err := roi.Users(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		Env           *Env
		Task          *roi.Task
		AllTaskStatus []roi.Status
		Versions      []*roi.Version
		Users         []*roi.User
	}{
		Env:           env,
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
		Users:         us,
	}
	return executeTemplate(w, "update-task", recipe)
}

func updateTaskPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, task, err := roi.SplitTaskID(id)
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
	t, err := roi.GetTask(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	t.Status = roi.Status(r.FormValue("status"))
	t.Assignee = assignee
	t.DueDate = tforms["due_date"]
	t.PublishVersion = r.FormValue("publish_version")
	t.ApprovedVersion = r.FormValue("approved_version")
	t.ReviewVersion = r.FormValue("review_version")
	t.WorkingVersion = r.FormValue("working_version")

	err = roi.UpdateTask(DB, t)
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
	show, _, _, err := roi.SplitUnitID(id)
	if err != nil {
		return err
	}
	site, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		Env           *Env
		Show          string
		IDs           []string
		Tasks         []string
		AllTaskStatus []roi.Status
	}{
		Env:           env,
		Show:          show,
		IDs:           ids,
		Tasks:         site.Tasks,
		AllTaskStatus: roi.AllTaskStatus,
	}
	return executeTemplate(w, "update-multi-tasks", recipe)
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
		show, grp, unit, err := roi.SplitUnitID(id)
		if err != nil {
			return err
		}
		s, err := roi.GetTask(DB, show, grp, unit, task)
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
			s.Status = roi.Status(status)
		}
		if assignee != "" {
			s.Assignee = assignee
		}
		roi.UpdateTask(DB, s)
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
	return executeTemplate(w, "history-go", -2)
}

func reviewTaskHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return reviewTaskPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	vs, err := roi.TaskVersions(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	showAllVersions := false
	if r.FormValue("all-versions") != "" {
		showAllVersions = true
	}
	if !showAllVersions {
		// 마지막 버전만 보인다.
		if len(vs) != 0 {
			vs = vs[len(vs)-1:]
		}
	}
	reviews := make(map[string][]*roi.Review)
	for _, v := range vs {
		rvs, err := roi.VersionReviews(DB, v.Show, v.Group, v.Unit, v.Task, v.Version)
		if err != nil {
			return err
		}
		reviews[v.ID()] = rvs
	}
	recipe := struct {
		Env             *Env
		Task            *roi.Task
		Versions        []*roi.Version
		Reviews         map[string][]*roi.Review
		ShowAllVersions bool
	}{
		Env:             env,
		Task:            t,
		Versions:        vs,
		Reviews:         reviews,
		ShowAllVersions: showAllVersions,
	}
	return executeTemplate(w, "review-task", recipe)
}

func reviewTaskPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id", "version", "msg", "status")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	ver := r.FormValue("version")
	status := roi.Status(r.FormValue("status"))
	rv := &roi.Review{
		Show:    show,
		Group:   grp,
		Unit:    unit,
		Task:    task,
		Version: ver,
		// 할일: 실제 리뷰어를 전달받도록 할 것.
		Reviewer:  env.User.ID,
		Messenger: env.User.ID,
		Msg:       r.FormValue("msg"),
		Status:    status,
	}
	err = roi.AddReview(DB, rv)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	if status != "" {
		switch status {
		case roi.StatusApproved:
			t.ReviewVersion = ""
			t.ApprovedVersion = ver
		case roi.StatusRetake:
			t.ReviewVersion = ""
		default:
			return roi.BadRequest("invalid review status: %s", status)
		}
		err = roi.UpdateTask(DB, t)
		if err != nil {
			return err
		}
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
