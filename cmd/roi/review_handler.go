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
		return roi.BadRequest("no shows in roi")
	}
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	ctg := r.FormValue("category")
	if ctg == "" {
		ctg = "shot"
	}
	if show == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			show = shows[0].Show
		}
		http.Redirect(w, r, "/review?show="+show+"&category="+ctg, http.StatusSeeOther)
		return nil
	}
	ts := make([]*roi.Task, 0)
	ts, err = roi.TasksNeedReview(DB, show, ctg)
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
		LoggedInUser string
		Shows        []*roi.Show
		Show         string
		Category     string
		ByDue        map[time.Time][]*roi.Task
	}{
		LoggedInUser: env.User.ID,
		Shows:        shows,
		Show:         show,
		Category:     ctg,
		ByDue:        tsd,
	}
	return executeTemplate(w, "review", recipe)
}
