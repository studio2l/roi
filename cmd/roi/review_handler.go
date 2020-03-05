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
	us, err := roi.UnitsHavingDue(DB, show, ctg)
	if err != nil {
		return err
	}
	usd := make(map[string][]*roi.Unit)
	for _, u := range us {
		due := stringFromDate(u.DueDate)
		if usd[due] == nil {
			usd[due] = make([]*roi.Unit, 0)
		}
		usd[due] = append(usd[due], u)
	}
	recipe := struct {
		LoggedInUser  string
		Shows         []*roi.Show
		Show          string
		Category      string
		ByDue         map[string][]*roi.Unit
		AllUnitStatus []roi.UnitStatus
	}{
		LoggedInUser:  env.User.ID,
		Shows:         shows,
		Show:          show,
		Category:      ctg,
		ByDue:         usd,
		AllUnitStatus: roi.AllUnitStatus,
	}
	return executeTemplate(w, "review.html", recipe)
}
