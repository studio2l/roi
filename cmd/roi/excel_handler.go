package main

import (
	"errors"
	"net/http"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/studio2l/roi"
)

func uploadExcelHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return uploadExcelPostHandler(w, r, env)
	}
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
	w.Header().Set("Cache-control", "no-cache")
	recipe := struct {
		Env *Env
	}{
		Env: env,
	}
	return executeTemplate(w, "upload-excel", recipe)
}

func uploadExcelPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	r.ParseMultipartForm(200000) // 사용하는 최대 메모리 사이즈: 200KB
	fileHeaders := r.MultipartForm.File["excel"]
	if len(fileHeaders) == 0 {
		return roi.BadRequest("excel file not uploaded")
	}
	fh := fileHeaders[0]
	f, err := fh.Open()
	if err != nil {
		return err
	}
	xl, err := excelize.OpenReader(f)
	if err != nil {
		return err
	}
	rows := xl.GetRows("Sheet1")
	if len(rows) == 0 {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
	}
	row0 := rows[0]
	title := make(map[int]string)
	for j, cell := range row0 {
		if cell != "" {
			title[j] = cell
		}
	}

	attr := make(map[string]string)

	// pop은 attr에 해당 속성이 있으면 지우고 반환한다.
	pop := func(m map[string]string, key string) string {
		v, ok := m[key]
		if ok {
			delete(m, key)
		}
		return v
	}

	for _, row := range rows[1:] {
		for i := range title {
			attr[title[i]] = row[i]
		}

		show := pop(attr, "show")
		grp := pop(attr, "group")
		unit := pop(attr, "unit")

		if unit == "" {
			// 엑셀의 뒤쪽 열들은 비어있더라도 rows에 추가되는 경우가 있다.
			// unit의 유무로 해당 줄이 비어있는지를 판단한다.
			continue
		}

		add := false
		u, err := roi.GetUnit(DB, show, grp, unit)
		if err != nil {
			if !errors.As(err, &roi.NotFoundError{}) {
				return err
			}
			u = &roi.Unit{
				Show:  show,
				Group: grp,
				Unit:  unit,
			}
			add = true
		}
		status := pop(attr, "status")
		if status != "" {
			u.Status = roi.Status(status)
		}
		desc := pop(attr, "description")
		if desc != "" {
			u.Description = desc
		}
		cg := pop(attr, "cg_description")
		if cg != "" {
			u.CGDescription = cg
		}
		tags := fieldSplit(pop(attr, "tags"))
		if len(tags) != 0 {
			// 태그는 덮어쓰지 않고 추가한다.
			has := make(map[string]bool)
			for _, t := range u.Tags {
				has[t] = true
			}
			for _, t := range tags {
				if !has[t] {
					u.Tags = append(u.Tags, t)
				}
			}
		}
		// 남은 속성들은 커스텀 속성에 업데이트 한다.
		if u.Attrs == nil {
			u.Attrs = make(map[string]string)
		}
		for k, v := range attr {
			u.Attrs[k] = v
		}

		if add {
			err = roi.AddUnit(DB, u)
			if err != nil {
				return err
			}
		} else {
			err = roi.UpdateUnit(DB, u)
			if err != nil {
				return err
			}
		}
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
