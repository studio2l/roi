package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/lib/pq"
)

var reValidShotID = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+$`)

// IsValidShotID은 해당 이름이 샷 이름으로 적절한지 여부를 반환한다.
func IsValidShotID(id string) bool {
	return reValidShotID.MatchString(id)
}

type ShotStatus string

const (
	ShotWaiting    ShotStatus = "waiting"
	ShotInProgress            = "in-progress"
	ShotDone                  = "done"
	ShotHold                  = "hold"
	ShotOmit                  = "omit"
)

type Shot struct {
	// 샷 아이디. 프로젝트 내에서 고유해야 한다.
	// 영문자와 숫자, 언더바(_) 만 사용할 것.
	// 예) CG_0010, EP01_SC01_0010
	ID string

	// 관련 아이디
	ProjectID string

	// 샷 정보
	Status        ShotStatus
	EditOrder     int
	Description   string
	CGDescription string
	TimecodeIn    string
	TimecodeOut   string
	Duration      int
	Tags          []string
}

func (s *Shot) dbValues() []interface{} {
	if s == nil {
		s = &Shot{}
	}
	return []interface{}{
		s.ID,
		s.ProjectID,
		s.Status,
		s.EditOrder,
		s.Description,
		s.CGDescription,
		s.TimecodeIn,
		s.TimecodeOut,
		s.Duration,
		pq.Array(s.Tags),
	}
}

var ShotTableFields = []string{
	// uniqid는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	"uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"id STRING NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %')",
	"project_id STRING NOT NULL CHECK (length(project_id) > 0) CHECK (project_id NOT LIKE '% %')",
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
	"UNIQUE(id, project_id)",
}

var ShotTableKeys = []string{
	"id",
	"project_id",
	"status",
	"edit_order",
	"description",
	"cg_description",
	"timecode_in",
	"timecode_out",
	"duration",
	"tags",
}

var ShotTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9", "$10",
}

// AddShot은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddShot(db *sql.DB, prj string, s *Shot) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	if s == nil {
		return errors.New("nil Shot is invalid")
	}
	keys := strings.Join(ShotTableKeys, ", ")
	idxs := strings.Join(ShotTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO shots (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, s.dbValues()...); err != nil {
		return err
	}
	return nil
}

// ShotExist는 db에 해당 샷이 존재하는지를 검사한다.
func ShotExist(db *sql.DB, prj, shot string) (bool, error) {
	stmt := "SELECT id FROM shots WHERE project_id=$1 AND id=$2 LIMIT 1"
	rows, err := db.Query(stmt, prj, shot)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// shotFromRows는 테이블의 한 열에서 샷을 받아온다.
func shotFromRows(rows *sql.Rows) (*Shot, error) {
	s := &Shot{}
	err := rows.Scan(
		&s.ID, &s.ProjectID, &s.Status,
		&s.EditOrder, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut,
		&s.Duration, pq.Array(&s.Tags),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetShot은 db의 특정 프로젝트에서 샷 이름으로 해당 샷을 찾는다.
// 만일 그 이름의 샷이 없다면 nil이 반환된다.
func GetShot(db *sql.DB, prj string, shot string) (*Shot, error) {
	keystr := strings.Join(ShotTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM shots WHERE project_id=$1 AND id=$2 LIMIT 1", keystr)
	rows, err := db.Query(stmt, prj, shot)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, nil
	}
	return shotFromRows(rows)
}

// SearchShots는 db의 특정 프로젝트에서 검색 조건에 맞는 샷 리스트를 반환한다.
func SearchShots(db *sql.DB, prj, shot, tag, status string) ([]*Shot, error) {
	keystr := strings.Join(ShotTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM shots", keystr)
	where := make([]string, 0)
	vals := make([]interface{}, 0)
	where = append(where, "project_id=$1")
	vals = append(vals, prj)
	i := 2 // $1 은 이미 project_id를 찾는데 쓰임
	if shot != "" {
		where = append(where, fmt.Sprintf("id=$%d", i))
		vals = append(vals, shot)
		i++
	}
	if tag != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(tags)", i))
		vals = append(vals, tag)
		i++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("status=$%d", i))
		vals = append(vals, status)
	}
	wherestr := strings.Join(where, " AND ")
	if wherestr != "" {
		stmt += " WHERE " + wherestr
	}
	rows, err := db.Query(stmt, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shots := make([]*Shot, 0)
	for rows.Next() {
		s, err := shotFromRows(rows)
		if err != nil {
			return nil, err
		}
		shots = append(shots, s)
	}
	sort.Slice(shots, func(i int, j int) bool {
		return shots[i].ID <= shots[j].ID
	})
	return shots, nil
}

// DeleteShot은 db의 특정 프로젝트에서 샷을 하나 지운다.
func DeleteShot(db *sql.DB, prj string, shot string) error {
	stmt := "DELETE FROM shots WHERE project_id=$1 AND id=$2"
	if _, err := db.Exec(stmt, prj, shot); err != nil {
		return err
	}
	return nil
}
