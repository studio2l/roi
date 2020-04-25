package main

import (
	"fmt"
	"net/http"

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
	target := r.FormValue("target")
	if target == "" {
		target = "having-due"
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
		http.Redirect(w, r, "/review?show="+show+"&category="+ctg+"&target="+target, http.StatusSeeOther)
		return nil
	}
	ts := make([]*roi.Task, 0)
	if target == "having-due" {
		ts, err = roi.TasksHavingDue(DB, show, ctg)
		if err != nil {
			return err
		}
	} else if target == "need-review" {
		ts, err = roi.TasksNeedReview(DB, show, ctg)
		if err != nil {
			return err
		}
	} else {
		return roi.BadRequest(fmt.Sprintf("invalid review target: %s", target))
	}
	tsd := make(map[string][]*roi.Task)
	for _, t := range ts {
		due := stringFromDate(t.DueDate)
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
		Target       string
		ByDue        map[string][]*roi.Task
	}{
		LoggedInUser: env.User.ID,
		Shows:        shows,
		Show:         show,
		Category:     ctg,
		Target:       target,
		ByDue:        tsd,
	}
	return executeTemplate(w, "review.bml", recipe)
}
