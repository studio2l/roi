package roi

import (
	"strconv"
	"strings"

	"github.com/lib/pq"
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
	Tags          []string
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
	"tags STRING[]",
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
		{"tags", q(strings.Join(s.Tags, ","))},
	}
	return kv
}

func (s Shot) toOrdMap() *ordMap {
	o := newOrdMap()
	o.Set("book", s.Book)
	o.Set("scene", s.Scene)
	o.Set("shot", s.Name)
	o.Set("status", s.Status)
	o.Set("edit_order", s.EditOrder)
	o.Set("description", s.Description)
	o.Set("cg_description", s.CGDescription)
	o.Set("timecode_in", s.TimecodeIn)
	o.Set("timecode_out", s.TimecodeOut)
	o.Set("duration", s.Duration)
	o.Set("tags", pq.Array(s.Tags))
	return o
}
