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
	rows, err := db.Query("SELECT code FROM projects where code=$1 LIMIT 1", prj)
	if err != nil {
		log.Print("project selection error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("internal error during project selection, sorry!")})
		w.Write(resp)
		return
	}
	defer rows.Close()
	if !rows.Next() {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("project '%s' not exists", prj)})
		w.Write(resp)
		return
	}
	name := r.PostFormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("'name' not specified")})
		w.Write(resp)
		return
	}
	// 할일: 유효한 샷 이름인지 검사
	stmt := fmt.Sprintf("SELECT shot FROM %s_shots where shot=$1 LIMIT 1", prj)
	rows, err = db.Query(stmt, name)
	if err != nil {
		log.Print("shot selection error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("internal error during shot selection, sorry!")})
		w.Write(resp)
		return
	}
	defer rows.Close()
	if rows.Next() {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("shot '%s' already exists", name)})
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
		Name:          name,
		Scene:         r.PostFormValue("scene"),
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
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(response{Err: fmt.Sprintf("%s", err)})
		w.Write(resp)
		return
	}
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(response{Msg: fmt.Sprintf("successfully add a shot: '%s'", name)})
	w.Write(resp)
}
