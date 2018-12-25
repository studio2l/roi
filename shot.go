package roi

import (
	"regexp"

	"github.com/lib/pq"
)

var reValidShotName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+$`)

// IsValidShotName은 해당 이름이 샷 이름으로 적절한지 여부를 반환한다.
func IsValidShotName(name string) bool {
	return reValidShotName.MatchString(name)
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
	// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	"id UUID PRIMARY KEY DEFAULT gen_random_uuid()",
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

func (s Shot) toOrdMap() *ordMap {
	o := newOrdMap()
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
