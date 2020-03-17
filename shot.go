package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CreateTableIfNotExistShowsStmt는 DB에 shots 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsShotsStmt = `CREATE TABLE IF NOT EXISTS shots (
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
	assets STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	UNIQUE(show, shot),
	CONSTRAINT shots_pk PRIMARY KEY (show, shot)
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
	Status        Status   `db:"status"`
	EditOrder     int      `db:"edit_order"`
	Description   string   `db:"description"`
	CGDescription string   `db:"cg_description"`
	TimecodeIn    string   `db:"timecode_in"`
	TimecodeOut   string   `db:"timecode_out"`
	Duration      int      `db:"duration"`
	Tags          []string `db:"tags"`

	// Assets는 샷이 필요로 하는 애셋 이름 리스트이다.
	// 현재는 애셋이 같은 쇼 안에 존재할 때만 처리가 가능하다.
	// 여기에 등록된 애셋은 존재해야만 하며,
	// 애셋이 삭제되기 전 우선 모든 샷의 애셋 태그에서 지워져야 한다.
	Assets []string `db:"assets"`

	// Tasks는 샷에 작업중인 어떤 태스크가 있는지를 나타낸다.
	// 웹 페이지에는 여기에 포함된 태스크만 이 순서대로 보여져야 한다.
	//
	// 참고: 여기에 포함되어 있다면 db내에 해당 태스크가 존재해야 한다.
	// 반대로 여기에 포함되어 있지 않지만 db내에는 존재하는 태스크가 있을 수 있다.
	// 그 태스크는 (예를 들어 태스크가 Omit 되는 등의 이유로) 숨겨진 태스크이며,
	// 직접 지우지 않는 한 db에 보관된다.
	Tasks []string `db:"tasks"`

	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	DueDate   time.Time `db:"due_date"`
}

// ID는 Shot의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Shot) ID() string {
	return s.Show + "/shot/" + s.Shot
}

// ReviewTargetFromShot은 Shot을 샷과 어셋의 공통된 주요 기능을 가진 Unit으로 변경한다.
func ReviewTargetFromShot(s *Shot) *ReviewTarget {
	return &ReviewTarget{
		Show:     s.Show,
		Category: "shot",
		Level:    "unit",
		Name:     s.Shot,
		Status:   s.Status,
		DueDate:  s.DueDate,
	}
}

// SplitShotID는 받아들인 샷 아이디를 쇼, 샷으로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitShotID(id string) (string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", BadRequest(fmt.Sprintf("invalid shot id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	shot := ns[2]
	if show == "" || ctg != "shot" || shot == "" {
		return "", "", BadRequest(fmt.Sprintf("invalid shot id: %s", id))
	}
	return show, shot, nil
}

// verifyShotID는 받아들인 샷 아이디가 유효하지 않다면 에러를 반환한다.
func verifyShotID(id string) error {
	show, shot, err := SplitShotID(id)
	if err != nil {
		return err
	}
	err = verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyShotName(shot)
	if err != nil {
		return err
	}
	return err
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
	// reShotName은 유효한 샷 이름을 나타내는 정규식이다.
	reShotName = regexp.MustCompile(`^[a-zA-Z]+[_]?[0-9]+[_]?[a-zA-Z]*$`)
	// reShotBase는 샷 이름에서 접미어가 빠진 이름을 나타내는 정규식이다.
	reShotBase = regexp.MustCompile(`^[a-zA-Z]+[_]?[0-9]+`)
	// reShotBase는 샷 이름의 접두어를 나타내는 정규식이다.
	reShotPrefix = regexp.MustCompile(`^[a-zA-Z]+`)
)

// verifyShotName은 받아들인 샷 이름이 유효하지 않다면 에러를 반환한다.
func verifyShotName(shot string) error {
	if !reShotName.MatchString(shot) {
		return BadRequest(fmt.Sprintf("invalid shot name: %s", shot))
	}
	return nil
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

// verifyShot은 받아들인 샷이 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyShot(db *sql.DB, s *Shot) error {
	if s == nil {
		return fmt.Errorf("nil shot")
	}
	err := verifyShotID(s.ID())
	if err != nil {
		return err
	}
	err = verifyUnitStatus(s.Status)
	if err != nil {
		return err
	}
	// 태스크에는 순서가 있으므로 사이트에 정의된 순서대로 재정렬한다.
	si, err := GetSite(db)
	if err != nil {
		return err
	}
	hasTask := make(map[string]bool)
	taskIdx := make(map[string]int)
	for i, task := range si.ShotTasks {
		hasTask[task] = true
		taskIdx[task] = i
	}
	for _, task := range s.Tasks {
		if !hasTask[task] {
			return BadRequest(fmt.Sprintf("task %q not defined at site", task))
		}
	}
	sort.Slice(s.Tasks, func(i, j int) bool {
		return taskIdx[s.Tasks[i]] <= taskIdx[s.Tasks[j]]
	})
	sort.Slice(s.Assets, func(i, j int) bool {
		return strings.Compare(s.Assets[i], s.Assets[j]) <= 0
	})
	for _, asset := range s.Assets {
		_, err := GetAsset(db, s.Show+"/asset/"+asset)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddShot은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddShot(db *sql.DB, s *Shot) error {
	err := verifyShot(db, s)
	if err != nil {
		return err
	}
	s.Name = ShotBase(s.Shot)
	s.Prefix = ShotPrefix(s.Shot)
	// 부모가 있는지 검사
	_, err = GetShow(db, s.Show)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO shots (%s) VALUES (%s)", keys, idxs), vs...),
	}
	// 하위 태스크 생성
	for _, task := range s.Tasks {
		t := &Task{
			Show:     s.Show,
			Category: "shot",
			Unit:     s.Shot,
			Task:     task,
			Status:   StatusInProgress,
			DueDate:  time.Time{},
		}
		err := verifyTask(db, t)
		if err != nil {
			return err
		}
		st, err := addTaskStmts(t)
		if err != nil {
			return err
		}
		stmts = append(stmts, st...)
	}
	return dbExec(db, stmts)
}

// GetShot은 db에서 하나의 샷을 찾는다.
// 해당 샷이 존재하지 않는다면 nil과 NotFound 에러를 반환한다.
func GetShot(db *sql.DB, id string) (*Shot, error) {
	show, shot, err := SplitShotID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShow(db, show)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Shot{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	s := &Shot{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shots WHERE show=$1 AND shot=$2 LIMIT 1", keys), show, shot)
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("shot", id)
		}
		return nil, err
	}
	return s, err
}

// ShotsHavingDue는 db에서 마감일이 정해진 샷을 불러온다.
func ShotsHavingDue(db *sql.DB, show string) ([]*Shot, error) {
	ks, _, _, err := dbKIVs(&Shot{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shots WHERE show=$1 AND due_date!=$2", keys), show, time.Time{})
	ss := make([]*Shot, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		s := &Shot{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		ss = append(ss, s)
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("shot", "having-due-date")
		}
		return nil, err
	}
	return ss, nil
}

// ShotsNeedReview는 db에서 리뷰가 필요한 샷을 불러온다.
func ShotsNeedReview(db *sql.DB, show string) ([]*Shot, error) {
	ks, _, _, err := dbKIVs(&Shot{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shots WHERE show=$1 AND status=$2", keys), show, StatusNeedReview)
	ss := make([]*Shot, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		s := &Shot{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		ss = append(ss, s)
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("shot", "status-need-review")
		}
		return nil, err
	}
	return ss, nil
}

// SearchShots는 db의 특정 프로젝트에서 검색 조건에 맞는 샷 리스트를 반환한다.
func SearchShots(db *sql.DB, show string, shots []string, tag, status, task, assignee, task_status string, task_due_date time.Time) ([]*Shot, error) {
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
	if len(shots) != 0 {
		j := 0
		whereShots := "("
		for _, shot := range shots {
			if shot == "" {
				continue
			}
			if j != 0 {
				whereShots += " OR "
			}
			if shot == ShotPrefix(shot) {
				whereShots += fmt.Sprintf("shots.prefix=$%d", i)
			} else if shot == ShotBase(shot) {
				whereShots += fmt.Sprintf("shots.name=$%d", i)
			} else {
				whereShots += fmt.Sprintf("shots.shot=$%d", i)
			}
			vals = append(vals, shot)
			i++
			j++
		}
		whereShots += ")"
		where = append(where, whereShots)
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
		where = append(where, fmt.Sprintf("$%d::string = ANY(shots.tasks)", i))
		vals = append(vals, task)
		i++
	}
	if assignee != "" || task_status != "" || !task_due_date.IsZero() {
		stmt += " JOIN tasks ON (tasks.show = shots.show AND tasks.unit = shots.shot)"
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
	st := dbStmt(stmt, vals...)
	ss := make([]*Shot, 0)
	hasShot := make(map[string]bool)
	err = dbQuery(db, st, func(rows *sql.Rows) error {
		// 태스크 검색을 해 JOIN이 되면 샷이 중복으로 추가될 수 있다.
		// DISTINCT를 이용해 문제를 해결하려고 했으나 DB가 꺼진다.
		// 우선은 여기서 걸러낸다.
		s := &Shot{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		if !hasShot[s.Shot] {
			hasShot[s.Shot] = true
			ss = append(ss, s)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(ss, func(i int, j int) bool {
		return ss[i].Shot <= ss[j].Shot
	})
	return ss, nil
}

// UpdateShot은 db에서 해당 샷을 수정한다.
// 이 함수를 호출하기 전 해당 샷이 존재하는지 사용자가 검사해야 한다.
func UpdateShot(db *sql.DB, id string, s *Shot) error {
	err := verifyShotID(id)
	if err != nil {
		return err
	}
	err = verifyShot(db, s)
	if err != nil {
		return err
	}
	s.Name = ShotBase(s.Shot)
	s.Prefix = ShotPrefix(s.Shot)
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE shots SET (%s) = (%s) WHERE show='%s' AND shot='%s'", keys, idxs, s.Show, s.Shot), vs...),
	}
	// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
	for _, task := range s.Tasks {
		_, err := GetTask(db, id+"/"+task)
		if err != nil {
			if !errors.As(err, &NotFoundError{}) {
				return fmt.Errorf("get task: %s", err)
			} else {
				t := &Task{
					Show:     s.Show,
					Category: "shot",
					Unit:     s.Shot,
					Task:     task,
					Status:   StatusInProgress,
					DueDate:  time.Time{},
				}
				st, err := addTaskStmts(t)
				if err != nil {
					return err
				}
				stmts = append(stmts, st...)
			}
		}
	}
	return dbExec(db, stmts)
}

// DeleteShot은 해당 샷과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 샷이 없어도 에러를 내지 않기 때문에 검사를 원한다면 ShotExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteShot(db *sql.DB, id string) error {
	show, shot, err := SplitShotID(id)
	if err != nil {
		return err
	}
	_, err = GetShot(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM shots WHERE show=$1 AND shot=$2", show, shot),
		dbStmt("DELETE FROM tasks WHERE show=$1 AND category=$2 AND unit=$3", show, "shot", shot),
		dbStmt("DELETE FROM versions WHERE show=$1 AND category=$2 AND unit=$3", show, "shot", shot),
	}
	return dbExec(db, stmts)
}
