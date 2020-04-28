package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/studio2l/roi"
)

func addAssetHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addAssetPostHandler(w, r, env)
	}
	w.Header().Set("Cache-control", "no-cache")
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	if id == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		id = cfg.CurrentShow
		if id == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			shows, err := roi.AllShows(DB)
			if err != nil {
				return err
			}
			if len(shows) == 0 {
				return roi.BadRequest("no shows in roi")
			}
			id = shows[0].Show
		}
		http.Redirect(w, r, "/add-asset?id="+id, http.StatusSeeOther)
		return nil
	}
	sw, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	cfg.CurrentShow = id
	err = roi.UpdateUserConfig(DB, env.User.ID, cfg)
	if err != nil {
		return err
	}

	recipe := struct {
		LoggedInUser string
		Show         *roi.Show
	}{
		LoggedInUser: env.User.ID,
		Show:         sw,
	}
	return executeTemplate(w, "add-asset", recipe)
}

func addAssetPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id", "asset")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	asset := r.FormValue("asset")
	sh, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	s := &roi.Asset{
		Show:   id,
		Asset:  asset,
		Status: roi.StatusHold,
		Tasks:  sh.DefaultAssetTasks,
	}
	err = roi.AddAsset(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateAssetHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateAssetPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	s, err := roi.GetAsset(DB, id)
	if err != nil {
		return err
	}
	ts, err := roi.AssetTasks(DB, id)
	if err != nil {
		return err
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Task] = t
	}
	recipe := struct {
		LoggedInUser  string
		Asset         *roi.Asset
		AllUnitStatus []roi.Status
		Tasks         map[string]*roi.Task
		AllTaskStatus []roi.Status
		Thumbnail     string
	}{
		LoggedInUser:  env.User.ID,
		Asset:         s,
		AllUnitStatus: roi.AllUnitStatus,
		Tasks:         tm,
		AllTaskStatus: roi.AllTaskStatus,
		Thumbnail:     "data/show/" + id + "/thumbnail.png",
	}
	return executeTemplate(w, "update-asset", recipe)
}

func updateAssetPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	tasks := fieldSplit(r.FormValue("tasks"))
	tforms, err := parseTimeForms(r.Form, "due_date")
	if err != nil {
		return err
	}
	s, err := roi.GetAsset(DB, id)
	if err != nil {
		return err
	}
	s.Status = roi.Status(r.FormValue("status"))
	s.Description = r.FormValue("description")
	s.CGDescription = r.FormValue("cg_description")
	s.Tags = fieldSplit(r.FormValue("tags"))
	s.Tasks = tasks
	s.DueDate = tforms["due_date"]

	err = roi.UpdateAsset(DB, id, s)
	if err != nil {
		return err
	}
	err = saveImageFormFile(r, "thumbnail", fmt.Sprintf("data/show/%s/thumbnail.png", id))
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateMultiAssetsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.FormValue("post") != "" {
		// 많은 애셋 선택시 URL이 너무 길어져 잘릴 염려 때문에 GET을 사용하지 않아,
		// POST와 GET을 구분할 다른 방법이 필요했다. 더 나은 방법을 생각해 볼 것.
		return updateMultiAssetsPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	ids := r.Form["id"]
	id := ids[0]
	show, _, err := roi.SplitAssetID(id)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Show          string
		IDs           []string
		AllUnitStatus []roi.Status
	}{
		LoggedInUser:  env.User.ID,
		Show:          show,
		IDs:           ids,
		AllUnitStatus: roi.AllUnitStatus,
	}
	return executeTemplate(w, "update-multi-assets", recipe)

}

func updateMultiAssetsPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	ids := r.Form["id"]
	tforms, err := parseTimeForms(r.Form, "due_date")
	if err != nil {
		return err
	}
	dueDate := tforms["due_date"]
	status := r.FormValue("status")
	tags := make([]string, 0)
	for _, tag := range strings.Split(r.FormValue("tags"), ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if tag[0] != '+' && tag[0] != '-' {
			roi.BadRequest(fmt.Sprintf("tag must be started with +/- got %s", tag))
		}
		tags = append(tags, tag)
	}
	workingTasks := make([]string, 0)
	for _, task := range strings.Split(r.FormValue("tasks"), ",") {
		task = strings.TrimSpace(task)
		if task == "" {
			continue
		}
		if task[0] != '+' && task[0] != '-' {
			return roi.BadRequest(fmt.Sprintf("task must be started with +/- got %s", task))
		}
		workingTasks = append(workingTasks, task)
	}
	for _, id := range ids {
		s, err := roi.GetAsset(DB, id)
		if err != nil {
			return err
		}
		if !dueDate.IsZero() {
			s.DueDate = dueDate
		}
		if status != "" {
			s.Status = roi.Status(status)
		}
		for _, tag := range tags {
			prefix := tag[0]
			tag = tag[1:]
			if prefix == '+' {
				s.Tags = appendIfNotExist(s.Tags, tag)
			} else if prefix == '-' {
				s.Tags = removeIfExist(s.Tags, tag)
			}
		}
		for _, task := range workingTasks {
			prefix := task[0]
			task = task[1:]
			if prefix == '+' {
				s.Tasks = appendIfNotExist(s.Tasks, task)
			} else if prefix == '-' {
				s.Tasks = removeIfExist(s.Tasks, task)
			}
		}
		err = roi.UpdateAsset(DB, id, s)
		if err != nil {
			return err
		}
	}
	q := ""
	for i, id := range ids {
		if i != 0 {
			q += " "
		}
		asset := strings.Split(id, "/")[1]
		q += asset
	}
	// 여러 애셋 수정 페이지 전인 assets 페이지로 돌아간다.
	return executeTemplate(w, "history-go", -2)
}
