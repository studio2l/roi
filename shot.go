package roi

import "strconv"

type ShotStatus int

const (
	ShotWaiting = ShotStatus(iota)
	ShotInProgress
	ShotDone
	ShotHold
	ShotOmit
)

type Shot struct {
	Book   int
	Scene  string
	Name   string
	Status string

	Description   string
	CGDescription string
	TimecodeIn    string
	TimecodeOut   string
}

func ShotFromMap(m map[string]string) Shot {
	book, _ := strconv.Atoi(m["book"])
	return Shot{
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
	"book INT",
	"scene STRING NOT NULL",
	"name STRING NOT NULL CHECK (length(name) > 0)",
	"status STRING",
	"description STRING",
	"cg_description STRING",
	"timecode_in STRING",
	"timecode_out STRING",
	"UNIQUE (scene, name)",
}

func (s Shot) dbKeyValues() []KV {
	kv := []KV{
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
