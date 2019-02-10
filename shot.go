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

	// WorkingTasks는 샷에 작업중인 어떤 태스크가 있는지를 나타낸다.
	// 웹 페이지에는 여기에 포함된 태스크만 이 순서대로 보여져야 한다.
	//
	// 참고: 여기에 포함되어 있다면 db내에 해당 태스크가 존재해야 한다.
	// 반대로 여기에 포함되어 있지 않지만 db내에는 존재하는 태스크가 있을 수 있다.
	// 그 태스크는 (예를 들어 태스크가 Omit 되는 등의 이유로) 숨겨진 태스크이며,
	// 직접 지우지 않는 한 db에 보관된다.
	WorkingTasks []string
}

func (s *Shot) dbValues() []interface{} {
	if s == nil {
		s = &Shot{}
	}
	if s.Tags == nil {
		s.Tags = make([]string, 0)
	}
	if s.WorkingTasks == nil {
		s.WorkingTasks = make([]string, 0)
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
		pq.Array(s.WorkingTasks),
	}
}

var CreateTableIfNotExistsShotsStmt = `CREATE TABLE IF NOT EXISTS shots (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	id STRING NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %'),
	project_id STRING NOT NULL CHECK (length(project_id) > 0) CHECK (project_id NOT LIKE '% %'),
	status STRING NOT NULL,
	edit_order INT NOT NULL,
	description STRING NOT NULL,
	cg_description STRING NOT NULL,
	timecode_in STRING NOT NULL,
	timecode_out STRING NOT NULL,
	duration INT NOT NULL,
	tags STRING[] NOT NULL,
	working_tasks STRING[] NOT NULL,
	UNIQUE(id, project_id)
)`

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
	"working_tasks",
}

var ShotTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9", "$10",
	"$11",
}

// AddShot은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddShot(db *sql.DB, prj string, s *Shot) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	if s == nil {
		return errors.New("nil Shot is invalid")
	}
	if s.Tags == nil {
		s.Tags = make([]string, 0)
	}
	if s.WorkingTasks == nil {
		s.WorkingTasks = make([]string, 0)
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
		&s.Duration, pq.Array(&s.Tags), pq.Array(&s.WorkingTasks),
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
func SearchShots(db *sql.DB, prj, shot, tag, status, assignee string) ([]*Shot, error) {
	keystr := ""
	for i, k := range ShotTableKeys {
		if i != 0 {
			keystr += ", "
		}
		// 태스크에 있는 정보를 찾기 위해 JOIN 해야 할 경우가 있기 때문에
		// shots. 을 붙인다.
		keystr += "shots." + k
	}
	where := make([]string, 0)
	vals := make([]interface{}, 0)
	i := 1 // 인덱스가 1부터 시작이다.
	stmt := fmt.Sprintf("SELECT %s FROM shots", keystr)
	where = append(where, fmt.Sprintf("shots.project_id=$%d", i))
	vals = append(vals, prj)
	i++
	if shot != "" {
		where = append(where, fmt.Sprintf("shots.id=$%d", i))
		vals = append(vals, shot)
		i++
	}
	if tag != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(shots.tags)", i))
		vals = append(vals, tag)
		i++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("shots.status=$%d", i))
		vals = append(vals, status)
		i++
	}
	if assignee != "" {
		stmt += " JOIN tasks ON (tasks.shot_id = shots.id)"
		where = append(where, fmt.Sprintf("tasks.assignee=$%d", i))
		vals = append(vals, assignee)
		i++
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

// UpdateShotParam은 Shot에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateShot에서 사용한다.
type UpdateShotParam struct {
	Status        ShotStatus
	EditOrder     int
	Description   string
	CGDescription string
	TimecodeIn    string
	TimecodeOut   string
	Duration      int
	Tags          []string
	WorkingTasks  []string
}

func (u UpdateShotParam) keys() []string {
	return []string{
		"status",
		"edit_order",
		"description",
		"cg_description",
		"timecode_in",
		"timecode_out",
		"duration",
		"tags",
		"working_tasks",
	}
}

func (u UpdateShotParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateShotParam) values() []interface{} {
	if u.Tags == nil {
		u.Tags = make([]string, 0)
	}
	if u.WorkingTasks == nil {
		u.WorkingTasks = make([]string, 0)
	}
	return []interface{}{
		u.Status,
		u.EditOrder,
		u.Description,
		u.CGDescription,
		u.TimecodeIn,
		u.TimecodeOut,
		u.Duration,
		pq.Array(u.Tags),
		pq.Array(u.WorkingTasks),
	}
}

// UpdateShot은 db에서 해당 샷을 수정한다.
func UpdateShot(db *sql.DB, prj, shot string, upd UpdateShotParam) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	if shot == "" {
		return errors.New("shot id empty")
	}
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE shots SET (%s) = (%s) WHERE project_id='%s' AND id='%s'", keystr, idxstr, prj, shot)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// DeleteShot은 db의 특정 프로젝트에서 샷을 하나 지운다.
func DeleteShot(db *sql.DB, prj string, shot string) error {
	stmt := "DELETE FROM shots WHERE project_id=$1 AND id=$2"
	if _, err := db.Exec(stmt, prj, shot); err != nil {
		return err
	}
	return nil
}
