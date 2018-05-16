package roi

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

var ShotTableFields = []string{
	"project STRING NOT NULL CHECK (length(project) > 0)",
	"book INT",
	"scene STRING NOT NULL",
	"name STRING NOT NULL CHECK (length(name) > 0)",
	"status STRING",
	"description STRING",
	"cg_description STRING",
	"timecode_in STRING",
	"timecode_out STRING",
	"UNIQUE (project, scene, name)",
}

func (s Shot) dbKeyValues() []KV {
	kv := []KV{
		{"project", q(s.Project)},
		{"book", strconv.Itoa(s.Book)},
		{"scene", q(s.Scene)},
		{"name", q(s.Name)},
		{"status", q(s.Status)},
		{"description", q(s.Description)},
		{"cg_description", q(s.CGDescription)},
		{"timecode_in", q(s.TimecodeIn)},
		{"timecode_out", q(s.TimecodeOut)},
	}
	return kv
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
