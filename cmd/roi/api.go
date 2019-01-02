package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/studio2l/roi"
)

// addShotApiHander는 사용자가 api를 통해 샷을 생성할수 있도록 한다.
// 응답은 json 형식이고, 샷이 잘 생성되었다면 .msg,
// 샷 생성에 문제가 있었다면 .err에 메시지가 담긴다.
func addShotApiHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Msg string `json:"msg"`
		Err string `json:"err"`
	}

	w.Header().Set("Content-Type", "application/json")
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/roi?sslmode=disable")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("there was an internal error, sorry!")})
		w.Write(resp)
		return
	}

	prj := r.PostFormValue("project")
	if prj == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("'project' not specified")})
		w.Write(resp)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print("project selection error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("internal error during project selection, sorry!")})
		w.Write(resp)
		return
	}
	if !exist {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("project '%s' not exists", prj)})
		w.Write(resp)
		return
	}

	id := r.PostFormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("'id' not specified")})
		w.Write(resp)
		return
	}
	if !roi.IsValidShotID(id) {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("shot id is not valid: %s", id)})
		w.Write(resp)
		return
	}
	exist, err = roi.ShotExist(db, prj, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("internal error during shot check, sorry!")})
		w.Write(resp)
		return
	}
	if exist {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("shot '%s' already exists", id)})
		w.Write(resp)
		return
	}

	status := "waiting"
	v := r.PostFormValue("status")
	if v != "" {
		status = v
	}
	editOrder := 0
	v = r.PostFormValue("edit_order")
	if v != "" {
		e, err := strconv.Atoi(v)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp, _ := json.Marshal(response{Err: fmt.Sprintf("could not convert edit_order to int: %s", r.PostFormValue("edit_order"))})
			w.Write(resp)
			return
		}
		editOrder = e
	}
	duration := 0
	v = r.PostFormValue("duration")
	if v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp, _ := json.Marshal(response{Err: fmt.Sprintf("could not convert duration to int: %s", r.PostFormValue("duration"))})
			w.Write(resp)
			return
		}
		duration = d
	}
	s := roi.Shot{
		ID:            id,
		Status:        status,
		EditOrder:     editOrder,
		Description:   r.PostFormValue("description"),
		CGDescription: r.PostFormValue("cg_description"),
		TimecodeIn:    r.PostFormValue("timecode_in"),
		TimecodeOut:   r.PostFormValue("timecode_out"),
		Duration:      duration,
		Tags:          strings.Split(r.PostFormValue("tags"), ","),
	}
	err = roi.AddShot(db, prj, s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("%s", err)})
		w.Write(resp)
		return
	}
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(response{Msg: fmt.Sprintf("successfully add a shot: '%s'", id)})
	w.Write(resp)
}
