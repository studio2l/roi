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
	level := r.FormValue("level")
	if level == "" {
		level = "unit"
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
		http.Redirect(w, r, "/review?show="+show+"&category="+ctg+"&level="+level+"&target="+target, http.StatusSeeOther)
		return nil
	}
	rts := make([]*roi.ReviewTarget, 0)
	if target == "having-due" {
		rts, err = roi.ReviewTargetsHavingDue(DB, show, ctg, level)
		if err != nil {
			return err
		}
	} else if target == "need-review" {
		rts, err = roi.ReviewTargetsNeedReview(DB, show, ctg, level)
		if err != nil {
			return err
		}
	} else {
		return roi.BadRequest(fmt.Sprintf("invalid review target: %s", target))
	}
	rtsd := make(map[string][]*roi.ReviewTarget)
	for _, rt := range rts {
		due := stringFromDate(rt.DueDate)
		if rtsd[due] == nil {
			rtsd[due] = make([]*roi.ReviewTarget, 0)
		}
		rtsd[due] = append(rtsd[due], rt)
	}
	recipe := struct {
		LoggedInUser string
		Shows        []*roi.Show
		Show         string
		Category     string
		Level        string
		Target       string
		ByDue        map[string][]*roi.ReviewTarget
	}{
		LoggedInUser: env.User.ID,
		Shows:        shows,
		Show:         show,
		Category:     ctg,
		Level:        level,
		Target:       target,
		ByDue:        rtsd,
	}
	return executeTemplate(w, "review.bml", recipe)
}
