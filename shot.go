package roi

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CreateTableIfNotExistShowsStmt는 DB에 shots 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsShotsStmt = `CREATE TABLE IF NOT EXISTS shots (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	name STRING NOT NULL,
	prefix STRING NOT NULL,
	status STRING NOT NULL CHECK (length(status) > 0),
	edit_order INT NOT NULL,
	description STRING NOT NULL,
	cg_description STRING NOT NULL,
	timecode_in STRING NOT NULL,
	timecode_out STRING NOT NULL,
	duration INT NOT NULL,
	tags STRING[] NOT NULL,
	working_tasks STRING[] NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	UNIQUE(show, shot)
)`

type Shot struct {
	Show string `db:"show"`
	Shot string `db:"shot"`

	// 주의: 아래 필드는 이 패키지 외부에서 접근하지 말 것.
	// 원래는 소문자로 시작해야 맞으나 아직 dbKVs가 수출되지 않는 값들의
	// 값을 받아오지 못해서 잠시 대문자로 시작하도록 하였음.
	Name   string `db:"name"`
	Prefix string `db:"prefix"`

	// 샷 정보
	Status        ShotStatus `db:"status"`
	EditOrder     int        `db:"edit_order"`
	Description   string     `db:"description"`
	CGDescription string     `db:"cg_description"`
	TimecodeIn    string     `db:"timecode_in"`
	TimecodeOut   string     `db:"timecode_out"`
	Duration      int        `db:"duration"`
	Tags          []string   `db:"tags"`

	// WorkingTasks는 샷에 작업중인 어떤 태스크가 있는지를 나타낸다.
	// 웹 페이지에는 여기에 포함된 태스크만 이 순서대로 보여져야 한다.
	//
	// 참고: 여기에 포함되어 있다면 db내에 해당 태스크가 존재해야 한다.
	// 반대로 여기에 포함되어 있지 않지만 db내에는 존재하는 태스크가 있을 수 있다.
	// 그 태스크는 (예를 들어 태스크가 Omit 되는 등의 이유로) 숨겨진 태스크이며,
	// 직접 지우지 않는 한 db에 보관된다.
	WorkingTasks []string `db:"working_tasks"`

	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	DueDate   time.Time `db:"due_date"`
}

// ID는 Shot의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Shot) ID() string {
	return s.Show + "/" + s.Shot
}

// 샷 이름은 일반적으로 (시퀀스를 나타내는) 접두어, 샷 번호, 접미어로 나뉜다.
// 접두어와 샷 번호는 항상 필요하지만, 접미어는 없어도 된다.
//
// CG0010a에서 접두어는 CG, 샷 번호는 0010, 접미어는 a이다.
//
// 접두어와 샷번호, 접미어를 언더바(_)를 통해서 떨어뜨리거나, 그냥 붙여쓰는것이 가능하다.
//
// 아래는 모두 유효한 샷 이름이다.
//
// CG0010
// CG0010a
// CG0010_a
// CG_0010a
// CG_0010_a
//
// 마지막 샷 이름의 경우 기본 이름은 CG_0010, 접미어는 CG가 된다.

var (
	// reShotName은 샷 이름을 나타내는 정규식이다.
	reShotName = regexp.MustCompile(`^[a-zA-Z]+[_]?[0-9]+[_]?[a-zA-Z]*$`)
	// reShotBase는 샷 이름에서 접미어가 빠진 이름을 나타내는 정규식이다.
	reShotBase = regexp.MustCompile(`^[a-zA-Z]+[_]?[0-9]+`)
	// reShotBase는 샷 이름의 접두어를 나타내는 정규식이다.
	reShotPrefix = regexp.MustCompile(`^[a-zA-Z]+`)
)

// IsValidShot은 해당 이름이 샷 이름으로 적절한지 여부를 반환한다.
func IsValidShot(shot string) bool {
	return reShotName.MatchString(shot)
}

// ShotBase는 샷 이름에서 접미어를 제외한 기본 이름 정보를 반환한다.
func ShotBase(shot string) string {
	return reShotBase.FindString(shot)
}

// ShotPrefix은 샷 이름에서 접두어 정보를 반환한다.
// 일반적으로 이 접두어는 시퀀스를 가리킨다.
func ShotPrefix(shot string) string {
	return reShotPrefix.FindString(shot)
}

// AddShot은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddShot(db *sql.DB, show string, s *Shot) error {
	if s == nil {
		return BadRequest("nil shot is invalid")
	}
	if !IsValidShot(s.Shot) {
		return BadRequest(fmt.Sprintf("invalid shot id: '%s'", s.Shot))
	}
	s.Name = ShotBase(s.Shot)
	s.Prefix = ShotPrefix(s.Shot)
	if !isValidShotStatus(s.Status) {
		return BadRequest(fmt.Sprintf("invalid shot status: '%s'", s.Status))
	}
	_, err := GetShow(db, show)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("INSERT INTO shots (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// GetShot은 db에서 하나의 샷을 찾는다.
// 해당 샷이 존재하지 않는다면 nil과 NotFound 에러를 반환한다.
func GetShot(db *sql.DB, show string, shot string) (*Shot, error) {
	if show == "" {
		return nil, BadRequest("show not specified")
	}
	if shot == "" {
		return nil, BadRequest("shot not specified")
	}
	_, err := GetShow(db, show)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Shot{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM shots WHERE show=$1 AND shot=$2 LIMIT 1", keys)
	rows, err := db.Query(stmt, show, shot)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		id := show + "/" + shot
		return nil, NotFound("shot", id)
	}
	s := &Shot{}
	err = scanFromRows(rows, s)
	return s, err
}

// SearchShots는 db의 특정 프로젝트에서 검색 조건에 맞는 샷 리스트를 반환한다.
func SearchShots(db *sql.DB, show, shot, tag, status, task, assignee, task_status string, task_due_date time.Time) ([]*Shot, error) {
	ks, _, _, err := dbKIVs(&Shot{})
	if err != nil {
		return nil, err
	}
	keys := ""
	for i, k := range ks {
		if i != 0 {
			keys += ", "
		}
		// 태스크에 있는 정보를 찾기 위해 JOIN 해야 할 경우가 있기 때문에
		// shots. 을 붙인다.
		keys += "shots." + k
	}
	where := make([]string, 0)
	vals := make([]interface{}, 0)
	i := 1 // 인덱스가 1부터 시작이다.
	stmt := fmt.Sprintf("SELECT %s FROM shots", keys)
	where = append(where, fmt.Sprintf("shots.show=$%d", i))
	vals = append(vals, show)
	i++
	if shot != "" {
		if shot == ShotPrefix(shot) {
			where = append(where, fmt.Sprintf("shots.prefix=$%d", i))
		} else if shot == ShotBase(shot) {
			where = append(where, fmt.Sprintf("shots.name=$%d", i))
		} else {
			where = append(where, fmt.Sprintf("shots.shot=$%d", i))
		}
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
	if task != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(shots.working_tasks)", i))
		vals = append(vals, task)
		i++
	}
	if assignee != "" || task_status != "" || !task_due_date.IsZero() {
		stmt += " JOIN tasks ON (tasks.show = shots.show AND tasks.shot = shots.shot)"
	}
	if assignee != "" {
		where = append(where, fmt.Sprintf("tasks.assignee=$%d", i))
		vals = append(vals, assignee)
		i++
	}
	if task_status != "" {
		where = append(where, fmt.Sprintf("tasks.status=$%d", i))
		vals = append(vals, task_status)
		i++
	}
	if !task_due_date.IsZero() {
		where = append(where, fmt.Sprintf("tasks.due_date=$%d", i))
		vals = append(vals, task_due_date)
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
	// 태스크 검색을 해 JOIN이 되면 샷이 중복으로 추가될 수 있다.
	// DISTINCT를 이용해 문제를 해결하려고 했으나 DB가 꺼진다.
	// 우선은 여기서 걸러낸다.
	hasShot := make(map[string]bool, 0)
	shots := make([]*Shot, 0)
	for rows.Next() {
		s := &Shot{}
		err := scanFromRows(rows, s)
		if err != nil {
			return nil, err
		}
		ok := hasShot[s.Shot]
		if ok {
			continue
		}
		hasShot[s.Shot] = true
		shots = append(shots, s)
	}
	sort.Slice(shots, func(i int, j int) bool {
		return shots[i].Shot <= shots[j].Shot
	})
	return shots, nil
}

// UpdateShotParam은 Shot에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateShot에서 사용한다.
type UpdateShotParam struct {
	Status        ShotStatus `db:"status"`
	EditOrder     int        `db:"edit_order"`
	Description   string     `db:"description"`
	CGDescription string     `db:"cg_description"`
	TimecodeIn    string     `db:"timecode_in"`
	TimecodeOut   string     `db:"timecode_out"`
	Duration      int        `db:"duration"`
	Tags          []string   `db:"tags"`
	WorkingTasks  []string   `db:"working_tasks"`
	DueDate       time.Time  `db:"due_date"`
}

// UpdateShot은 db에서 해당 샷을 수정한다.
func UpdateShot(db *sql.DB, show, shot string, upd UpdateShotParam) error {
	if !isValidShotStatus(upd.Status) {
		return BadRequest(fmt.Sprintf("invalid shot status: '%s'", upd.Status))
	}
	_, err := GetShot(db, show, shot)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(upd)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("UPDATE shots SET (%s) = (%s) WHERE show='%s' AND shot='%s'", keys, idxs, show, shot)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// DeleteShot은 해당 샷과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 샷이 없어도 에러를 내지 않기 때문에 검사를 원한다면 ShotExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteShot(db *sql.DB, show, shot string) error {
	_, err := GetShot(db, show, shot)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %w", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM shots WHERE show=$1 AND shot=$2", show, shot); err != nil {
		return fmt.Errorf("could not delete data from 'shots' table: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM tasks WHERE show=$1 AND shot=$2", show, shot); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE show=$1 AND shot=$2", show, shot); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
