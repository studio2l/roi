package main

import (
	"strconv"
	"time"
)

type Project struct {
	Name string
	Code string

	Status string

	Client        string
	Director      string
	Producer      string
	VFXSupervisor string
	VFXManager    string

	CrankIn     time.Time
	CrankUp     time.Time
	StartDate   time.Time
	ReleaseDate time.Time
	VFXDueDate  time.Time

	OutputSize [2]int
	LutFile    string
}

type Scene struct {
	Project string
	Name    string
}

type ShotStatus int

const (
	ShotWaiting = ShotStatus(iota)
	ShotInProgress
	ShotDone
	ShotHold
	ShotOmit
)

type Shot struct {
	Project string
	Book    int
	Scene   string
	Name    string
	Status  string

	Description   string
	CGDescription string
	TimecodeIn    string
	TimecodeOut   string
}

func ShotFromMap(m map[string]string) Shot {
	book, _ := strconv.Atoi(m["book"])
	return Shot{
		Project:       m["project"],
		Book:          book,
		Scene:         m["scene"],
		Name:          m["name"],
		Status:        m["status"],
		Description:   m["description"],
		CGDescription: m["cg_description"],
		TimecodeIn:    m["timecode_in"],
		TimecodeOut:   m["timecode_out"],
	}
}

func (s Shot) dbKeyTypeValues() []KTV {
	ktv := []KTV{
		{"project", "STRING", q(s.Project)},
		{"book", "INT", strconv.Itoa(s.Book)},
		{"scene", "STRING", q(s.Scene)},
		{"name", "STRING", q(s.Name)},
		{"status", "STRING", q(s.Status)},
		{"description", "STRING", q(s.Description)},
		{"cg_description", "STRING", q(s.CGDescription)},
		{"timecode_in", "STRING", q(s.TimecodeIn)},
		{"timecode_out", "STRING", q(s.TimecodeOut)},
	}
	return ktv
}

type TaskStatus int

const (
	TaskWaiting = TaskStatus(iota)
	TaskAssigned
	TaskInProgress
	TaskPending
	TaskRetake
	TaskDone
	TaskHold
	TaskOmit
)

type Task struct {
	Project string
	Scene   string
	Shot    string
	Type    string
	Status  string

	Assignee string
}
