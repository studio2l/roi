package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// rootHandler는 루트 페이지로 사용자가 접근했을때 그 사용자에게 필요한 정보를 맞춤식으로 제공한다.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	if session == nil || session["userid"] == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user := session["userid"]
	tasks, err := roi.UserTasks(DB, user)
	if err != nil {
		log.Printf("could not get user tasks: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// 태스크를 미리 아이디 기준으로 정렬해 두면 아래에서 사용되는
	// tasksOfDay 또한 아이디 기준으로 정렬된다.
	sort.Slice(tasks, func(i, j int) bool {
		ti := tasks[i]
		idi := ti.Show + "." + ti.Shot + "." + ti.Task
		tj := tasks[j]
		idj := tj.Show + "." + tj.Shot + "." + tj.Task
		return strings.Compare(idi, idj) <= 0
	})
	taskFromID := make(map[string]*roi.Task)
	for _, t := range tasks {
		tid := t.Show + "." + t.Shot + "." + t.Task
		taskFromID[tid] = t
	}
	tasksOfDay := make(map[string][]string, 28)
	for _, t := range tasks {
		due := stringFromDate(t.DueDate)
		if tasksOfDay[due] == nil {
			tasksOfDay[due] = make([]string, 0)
		}
		tid := t.Show + "." + t.Shot + "." + t.Task
		tasksOfDay[due] = append(tasksOfDay[due], tid)
	}
	// 앞으로 4주에 대한 태스크 정보를 보인다.
	// 총 기간이나 단위는 추후 설정할 수 있도록 할 것.
	timeline := make([]string, 28)
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	for i := range timeline {
		timeline[i] = stringFromDate(today.Add(time.Duration(i) * 24 * time.Hour))
	}
	numTasks := make(map[string]map[roi.TaskStatus]int)
	for _, t := range tasks {
		if numTasks[t.Show] == nil {
			numTasks[t.Show] = make(map[roi.TaskStatus]int)
		}
		numTasks[t.Show][t.Status] += 1
	}
	recipe := struct {
		LoggedInUser  string
		User          string
		Timeline      []string
		NumTasks      map[string]map[roi.TaskStatus]int
		TaskFromID    map[string]*roi.Task
		TasksOfDay    map[string][]string
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  session["userid"],
		User:          session["userid"],
		Timeline:      timeline,
		NumTasks:      numTasks,
		TaskFromID:    taskFromID,
		TasksOfDay:    tasksOfDay,
		AllTaskStatus: roi.AllTaskStatus,
	}
	err = executeTemplate(w, "user.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
