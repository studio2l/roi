package main

import (
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
	kind := r.FormValue("kind")
	if kind == "" {
		kind = "unit"
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
		http.Redirect(w, r, "/review?show="+show+"&category="+ctg+"&kind="+kind, http.StatusSeeOther)
		return nil
	}
	rts, err := roi.ReviewTargetsHavingDue(DB, show, ctg, kind)
	if err != nil {
		return err
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
		Kind         string
		ByDue        map[string][]*roi.ReviewTarget
	}{
		LoggedInUser: env.User.ID,
		Shows:        shows,
		Show:         show,
		Category:     ctg,
		Kind:         kind,
		ByDue:        rtsd,
	}
	return executeTemplate(w, "review.html", recipe)
}
