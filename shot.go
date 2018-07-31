package roi

import (
	"strconv"
)

type ShotStatus int

const (
	ShotWaiting = ShotStatus(iota)
	ShotInProgress
	ShotDone
	ShotHold
	ShotOmit
)

type Shot struct {
	Book          int
	Scene         string
	Name          string
	Status        string
	EditOrder     int
	Description   string
	CGDescription string
	TimecodeIn    string
	TimecodeOut   string
	Duration      int
	Tags          string // 콤마로 분리
}

func ShotFromMap(m map[string]string) Shot {
	return Shot{
		Book:          toInt(m["book"]),
		Scene:         m["scene"],
		Name:          m["shot"],
		Status:        m["status"],
		EditOrder:     toInt(m["edit_order"]),
		Description:   m["description"],
		CGDescription: m["cg_description"],
		TimecodeIn:    m["timecode_in"],
		TimecodeOut:   m["timecode_out"],
		Duration:      toInt(m["duration"]),
		Tags:          m["tags"],
	}
}

var ShotTableFields = []string{
	"book INT",
	"scene STRING NOT NULL CHECK (scene NOT LIKE '% %')",
	"shot STRING UNIQUE NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %')",
	"status STRING NOT NULL CHECK (length(status) > 0)  CHECK (status NOT LIKE '% %')",
	"edit_order INT NOT NULL",
	"description STRING",
	"cg_description STRING",
	"timecode_in STRING",
	"timecode_out STRING",
	"duration INT NOT NULL",
	"tags STRING",
	// 할일: 샷과 소스에 대해서 서로 어떤 역할을 가지는지 확실히 이해한 뒤 추가.
	// "base_source STRING",
	// "other_sources STRING[]",
}

func (s Shot) dbKeyValues() []KV {
	kv := []KV{
		{"book", strconv.Itoa(s.Book)},
		{"scene", q(s.Scene)},
		{"shot", q(s.Name)},
		{"status", q(s.Status)},
		{"edit_order", strconv.Itoa(s.EditOrder)},
		{"description", q(s.Description)},
		{"cg_description", q(s.CGDescription)},
		{"timecode_in", q(s.TimecodeIn)},
		{"timecode_out", q(s.TimecodeOut)},
		{"duration", strconv.Itoa(s.Duration)},
		{"tags", q(s.Tags)},
	}
	return kv
}
