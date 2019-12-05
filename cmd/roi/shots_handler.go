package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// shotsHandler는 /shots/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func shotsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	if len(shows) == 0 {
		return roi.BadRequest("no shows in roi")
	}
	cfg, err := roi.GetUserConfig(DB, env.SessionUser.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	query := r.FormValue("q")
	if show == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			show = shows[0].Show
		}
		http.Redirect(w, r, "/shots?show="+show+"&q="+query, http.StatusSeeOther)
		return nil
	}
	_, err = roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	cfg.CurrentShow = show
	err = roi.UpdateUserConfig(DB, env.SessionUser.ID, cfg)
	if err != nil {
		return err
	}

	f := make(map[string]string)
	for _, v := range strings.Fields(query) {
		kv := strings.Split(v, ":")
		if len(kv) == 1 {
			f["shot"] = v
		} else {
			f[kv[0]] = kv[1]
		}
	}
	// toTime은 문자열을 time.Time으로 변경하되 에러가 나면 버린다.
	toTime := func(s string) time.Time {
		t, _ := timeFromString(s)
		return t
	}
	shots, err := roi.SearchShots(DB, show, f["shot"], f["tag"], f["status"], f["task"], f["assignee"], f["task-status"], toTime(f["task-due"]))
	if err != nil {
		return err
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range shots {
		ts, err := roi.ShotTasks(DB, show, s.Shot)
		if err != nil {
			return err
		}
		tm := make(map[string]*roi.Task)
		for _, t := range ts {
			tm[t.Task] = t
		}
		tasks[s.Shot] = tm
	}
	recipe := struct {
		LoggedInUser  string
		Shows         []*roi.Show
		Show          string
		Shots         []*roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
		Query         string
	}{
		LoggedInUser:  env.SessionUser.ID,
		Shows:         shows,
		Show:          show,
		Shots:         shots,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tasks,
		AllTaskStatus: roi.AllTaskStatus,
		Query:         query,
	}
	return executeTemplate(w, "shots.html", recipe)
}
