package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/studio2l/roi"
)

func addUnitHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addUnitPostHandler(w, r, env)
	}
	w.Header().Set("Cache-control", "no-cache")
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	// 할일: id를 show로 변경할 것
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
				recipe := struct {
					LoggedInUser string
				}{
					LoggedInUser: env.User.ID,
				}
				return executeTemplate(w, "no-shows", recipe)
			}
			id = shows[0].Show
		}
		http.Redirect(w, r, "/add-unit?id="+id, http.StatusSeeOther)
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
		Env  *Env
		Show *roi.Show
	}{
		Env:  env,
		Show: sw,
	}
	return executeTemplate(w, "add-unit", recipe)
}

func addUnitPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id", "group", "unit")
	if err != nil {
		return err
	}
	// 할일: id를 show로 변경할 것
	id := r.FormValue("id")
	grp := r.FormValue("group")
	unit := r.FormValue("unit")
	s := &roi.Unit{
		Show:   id,
		Group:  grp,
		Unit:   unit,
		Status: roi.StatusInProgress,
	}
	err = roi.AddUnit(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateUnitHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateUnitPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, err := roi.SplitUnitID(id)
	if err != nil {
		return err
	}
	s, err := roi.GetUnit(DB, show, grp, unit)
	if err != nil {
		return err
	}
	ts, err := roi.UnitTasks(DB, show, grp, unit)
	if err != nil {
		return err
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Task] = t
	}
	recipe := struct {
		Env           *Env
		Unit          *roi.Unit
		AllUnitStatus []roi.Status
		Tasks         map[string]*roi.Task
		AllTaskStatus []roi.Status
		Thumbnail     string
	}{
		Env:           env,
		Unit:          s,
		AllUnitStatus: roi.AllUnitStatus,
		Tasks:         tm,
		AllTaskStatus: roi.AllTaskStatus,
		Thumbnail:     "data/show/" + id + "/thumbnail.png",
	}
	return executeTemplate(w, "update-unit", recipe)
}

func updateUnitPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, err := roi.SplitUnitID(id)
	if err != nil {
		return err
	}
	tforms, err := parseTimeForms(r.Form, "due_date")
	if err != nil {
		return err
	}
	s, err := roi.GetUnit(DB, show, grp, unit)
	if err != nil {
		return err
	}
	s.Status = roi.Status(r.FormValue("status"))
	s.EditOrder = atoi(r.FormValue("edit_order"))
	s.Description = r.FormValue("description")
	s.CGDescription = r.FormValue("cg_description")
	s.Tags = fieldSplit(r.FormValue("tags"))
	s.Assets = fieldSplit(r.FormValue("assets"))
	s.Tasks = fieldSplit(r.FormValue("tasks"))
	s.DueDate = tforms["due_date"]
	s.Attrs = make(roi.DBStringMap)

	for _, ln := range strings.Split(r.FormValue("attrs"), "\n") {
		kv := strings.SplitN(ln, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			continue
		}
		s.Attrs[k] = v
	}

	err = roi.UpdateUnit(DB, s)
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

func updateMultiUnitsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.FormValue("post") != "" {
		// 많은 샷 선택시 URL이 너무 길어져 잘릴 염려 때문에 GET을 사용하지 않아,
		// POST와 GET을 구분할 다른 방법이 필요했다. 더 나은 방법을 생각해 볼 것.
		return updateMultiUnitsPostHandler(w, r, env)
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
	recipe := struct {
		Env           *Env
		Show          string
		IDs           []string
		AllUnitStatus []roi.Status
	}{
		Env:           env,
		Show:          show,
		IDs:           ids,
		AllUnitStatus: roi.AllUnitStatus,
	}
	return executeTemplate(w, "update-multi-units", recipe)
}

func updateMultiUnitsPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
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
			roi.BadRequest("tag must be started with +/- got %s", tag)
		}
		tags = append(tags, tag)
	}
	assets := make([]string, 0)
	for _, asset := range strings.Split(r.FormValue("assets"), ",") {
		asset = strings.TrimSpace(asset)
		if asset == "" {
			continue
		}
		if asset[0] != '+' && asset[0] != '-' {
			return roi.BadRequest("asset must be started with +/- got %s", asset)
		}
		assets = append(assets, asset)
	}
	workingTasks := make([]string, 0)
	for _, task := range strings.Split(r.FormValue("tasks"), ",") {
		task = strings.TrimSpace(task)
		if task == "" {
			continue
		}
		if task[0] != '+' && task[0] != '-' {
			return roi.BadRequest("task must be started with +/- got %s", task)
		}
		workingTasks = append(workingTasks, task)
	}
	for _, id := range ids {
		show, grp, unit, err := roi.SplitUnitID(id)
		if err != nil {
			return err
		}
		s, err := roi.GetUnit(DB, show, grp, unit)
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
		for _, asset := range assets {
			prefix := asset[0]
			asset = asset[1:]
			if prefix == '+' {
				s.Assets = appendIfNotExist(s.Assets, asset)
			} else if prefix == '-' {
				s.Assets = removeIfExist(s.Assets, asset)
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
		err = roi.UpdateUnit(DB, s)
		if err != nil {
			return err
		}
	}
	// 할일: 아래코드는 어디서 왔는가? 이상하다...
	q := ""
	for i, id := range ids {
		if i != 0 {
			q += " "
		}
		unit := strings.Split(id, "/")[1]
		q += unit
	}
	// 여러 샷 수정 페이지 전인 units 페이지로 돌아간다.
	return executeTemplate(w, "history-go", -2)
}
