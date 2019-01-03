package roi

import (
	"regexp"

	"github.com/lib/pq"
)

var reValidShotID = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+$`)

// IsValidShotID은 해당 이름이 샷 이름으로 적절한지 여부를 반환한다.
func IsValidShotID(id string) bool {
	return reValidShotID.MatchString(id)
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
	ID            string
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
	// uniqid는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	"uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"id STRING UNIQUE NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %')",
	"status STRING NOT NULL CHECK (length(status) > 0)  CHECK (status NOT LIKE '% %')",
	"edit_order INT NOT NULL",
	"description STRING NOT NULL",
	"cg_description STRING NOT NULL",
	"timecode_in STRING NOT NULL",
	"timecode_out STRING NOT NULL",
	"duration INT NOT NULL",
	"tags STRING[] NOT NULL",
	// 할일: 샷과 소스에 대해서 서로 어떤 역할을 가지는지 확실히 이해한 뒤 추가.
	// "base_source STRING",
	// "other_sources STRING[]",
}

// ordMapFromShot은 샷 정보를 OrdMap에 담는다.
func ordMapFromShot(s Shot) *ordMap {
	o := newOrdMap()
	o.Set("id", s.ID)
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
