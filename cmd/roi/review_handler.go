package main

import (
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

// reviewHandler는 /shots/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func reviewHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	if len(shows) == 0 {
		recipe := struct {
			Env *Env
		}{
			Env: env,
		}
		return executeTemplate(w, "no-shows", recipe)
	}
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	if show == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			show = shows[0].Show
		}
		http.Redirect(w, r, "/review?show="+show, http.StatusSeeOther)
		return nil
	}
	ts := make([]*roi.Task, 0)
	ts, err = roi.TasksNeedReview(DB, show)
	if err != nil {
		return err
	}
	tsd := make(map[time.Time][]*roi.Task)
	for _, t := range ts {
		due := t.DueDate
		if tsd[due] == nil {
			tsd[due] = make([]*roi.Task, 0)
		}
		tsd[due] = append(tsd[due], t)
	}
	recipe := struct {
		Env   *Env
		Shows []*roi.Show
		Show  string
		ByDue map[time.Time][]*roi.Task
	}{
		Env:   env,
		Shows: shows,
		Show:  show,
		ByDue: tsd,
	}
	return executeTemplate(w, "review", recipe)
}
