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

// CreateTableIfNotExistShowsStmt는 DB에 tasks 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsTasksStmt = `CREATE TABLE IF NOT EXISTS tasks (
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	category STRING NOT NULL CHECK (length(category) > 0)  CHECK (category NOT LIKE '% %'),
	unit STRING NOT NULL CHECK (length(unit) > 0) CHECK (unit NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	due_date TIMESTAMPTZ NOT NULL,
	assignee STRING NOT NULL,
	publish_version STRING NOT NULL,
	approved_version STRING NOT NULL,
	review_versions STRING[] NOT NULL,
	working_version STRING NOT NULL,
	UNIQUE(show, category, unit, task),
	CONSTRAINT tasks_pk PRIMARY KEY (show, category, unit, task)
)`

type Task struct {
	// 관련 아이디
	Show     string `db:"show"`
	Category string `db:"category"` // 현재는 "shot" 만 존재
	Unit     string `db:"unit"`     // 샷 또는 애셋 유닛
	Task     string `db:"task"`     // 파트 또는 파트_요소로 구성된다. 예) fx, fx_fire

	Status   Status    `db:"status"`
	DueDate  time.Time `db:"due_date"`
	Assignee string    `db:"assignee"`

	PublishVersion  string   `db:"publish_version"`
	ApprovedVersion string   `db:"approved_version"`
	ReviewVersions  []string `db:"review_versions"`
	WorkingVersion  string   `db:"working_version"`
}

var taskDBKey string = strings.Join(dbKeys(&Task{}), ", ")
var taskDBIdx string = strings.Join(dbIdxs(&Task{}), ", ")
var _ []interface{} = dbVals(&Task{})

// ID는 Task의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (t *Task) ID() string {
	return t.Show + "/" + t.Category + "/" + t.Unit + "/" + t.Task
}

// UnitID는 부모 유닛의 아이디를 반환한다.
// 유닛은 샷 또는 애셋이다.
func (t *Task) UnitID() string {
	return t.Show + "/" + t.Category + "/" + t.Unit
}

// reTaskName은 가능한 태스크명을 정의하는 정규식이다.
// 태스크명은 언더바를 이용해 서브 태스크를 나타낼 수 있다.
var reTaskName = regexp.MustCompile(`^[a-zA-Z0-9]+(_[a-zA-Z0-9]+)?$`)

// verifyTaskName은 받아들인 샷 이름이 유효하지 않다면 에러를 반환한다.
func verifyTaskName(task string) error {
	if !reTaskName.MatchString(task) {
		return BadRequest(fmt.Sprintf("invalid task name: %s", task))
	}
	return nil
}

// SplitTaskID는 받아들인 샷 아이디를 쇼, 카테고리, 샷, 태스크로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitTaskID(id string) (string, string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 4 {
		return "", "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	unit := ns[2]
	task := ns[3]
	if show == "" || ctg == "" || unit == "" || task == "" {
		return "", "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	return show, ctg, unit, task, nil
}

// verifyTaskID는 받아들인 태스크 아이디가 유효하지 않다면 에러를 반환한다.
func verifyTaskID(id string) error {
	show, ctg, unit, task, err := SplitTaskID(id)
	if err != nil {
		return err
	}
	err = verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyCategoryName(ctg)
	if err != nil {
		return err
	}
	err = verifyUnitName(ctg, unit)
	if err != nil {
		return err
	}
	err = verifyTaskName(task)
	if err != nil {
		return err
	}
	return nil
}

// verifyTask는 받아들인 태스크가 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyTask(db *sql.DB, t *Task) error {
	if t == nil {
		return fmt.Errorf("nil task")
	}
	err := verifyTaskID(t.ID())
	if err != nil {
		return err
	}
	err = verifyTaskStatus(t.Status)
	if err != nil {
		return err
	}
	t.PublishVersion = strings.TrimSpace(t.PublishVersion)
	if t.PublishVersion != "" {
		_, err = GetVersion(db, t.ID()+"/"+t.PublishVersion)
		if err != nil {
			return fmt.Errorf("publish version: %w", err)
		}
	}
	t.ApprovedVersion = strings.TrimSpace(t.ApprovedVersion)
	if t.ApprovedVersion != "" {
		_, err = GetVersion(db, t.ID()+"/"+t.ApprovedVersion)
		if err != nil {
			return fmt.Errorf("approved version: %w", err)
		}
	}
	hasVersion := make(map[string]bool)
	for _, v := range t.ReviewVersions {
		v = strings.TrimSpace(v)
		if v != "" && !hasVersion[v] {
			hasVersion[v] = true
			_, err = GetVersion(db, t.ID()+"/"+v)
			if err != nil {
				return fmt.Errorf("review version: %w", err)
			}
		}
	}
	t.ReviewVersions = make([]string, 0, len(hasVersion))
	for v := range hasVersion {
		t.ReviewVersions = append(t.ReviewVersions, v)
	}
	sort.Strings(t.ReviewVersions)
	t.WorkingVersion = strings.TrimSpace(t.WorkingVersion)
	if t.WorkingVersion != "" {
		_, err = GetVersion(db, t.ID()+"/"+t.WorkingVersion)
		if err != nil {
			return fmt.Errorf("working version: %w", err)
		}
	}
	if t.Status == StatusDone && t.PublishVersion == "" {
		return BadRequest(fmt.Sprintf("cannot set task status to TaskDone: no publish version"))
	}
	if t.Status == StatusApproved && t.ApprovedVersion == "" {
		return BadRequest(fmt.Sprintf("cannot set task status to TaskApproved: no approved version"))
	}
	if t.Status == StatusNeedReview && len(t.ReviewVersions) == 0 {
		return BadRequest(fmt.Sprintf("cannot set task status to TaskNeedReview: no review versions"))
	}
	return nil
}

// AddTask는 db의 특정 쇼, 카테고리, 유닛에 태스크를 추가한다.
func AddTask(db *sql.DB, t *Task) error {
	err := verifyTask(db, t)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	ok, err := CheckUnitExist(db, t.UnitID())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unit not found: %s", t.UnitID())
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", taskDBKey, taskDBIdx), dbVals(t)...),
	}
	return dbExec(db, stmts)
}

// addTaskStmts는 태스크를 추가하는 db 구문을 반환한다.
// 부모가 있는지는 검사하지 않는다.
func addTaskStmts(t *Task) ([]dbStatement, error) {
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", taskDBKey, taskDBIdx), dbVals(t)...),
	}
	return stmts, nil
}

// UpdateTask는 db의 특정 태스크를 업데이트 한다.
// 이 함수를 호출하기 전 해당 태스크가 존재하는지 사용자가 검사해야 한다.
func UpdateTask(db *sql.DB, id string, t *Task) error {
	err := verifyTask(db, t)
	if err != nil {
		return err
	}
	oldt, err := GetTask(db, id)
	if err != nil {
		return err
	}
	if oldt.Category != t.Category {
		return fmt.Errorf("not allowed to change category of task")
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE tasks SET (%s) = (%s) WHERE show='%s' AND category='%s' AND unit='%s' AND task='%s'", taskDBKey, taskDBIdx, t.Show, t.Category, t.Unit, t.Task), dbVals(t)...),
	}
	return dbExec(db, stmts)
}

// GetTask는 db에서 하나의 태스크를 찾는다.
// 해당 태스크가 없다면 nil과 NotFound 에러를 반환한다.
func GetTask(db *sql.DB, id string) (*Task, error) {
	show, ctg, unit, task, err := SplitTaskID(id)
	if err != nil {
		return nil, err
	}
	// 부모가 있는지 검사
	unitID := show + "/" + ctg + "/" + unit
	ok, err := CheckUnitExist(db, unitID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unit not found: %s", unitID)
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND unit=$2 AND task=$3 LIMIT 1", taskDBKey), show, unit, task)
	t := &Task{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, t)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("task", id)
		}
		return nil, err
	}
	return t, err
}

// TasksHavingDue는 db에서 마감일이 정해진 태스크를 불러온다.
func TasksHavingDue(db *sql.DB, show, ctg string) ([]*Task, error) {
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND category=$2 AND due_date!=$3", taskDBKey), show, ctg, time.Time{})
	ts := make([]*Task, 0)
	err := dbQuery(db, stmt, func(rows *sql.Rows) error {
		s := &Task{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		ts = append(ts, s)
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("task", "having-due-date")
		}
		return nil, err
	}
	return ts, nil
}

// TasksNeedReview는 db에서 리뷰가 필요한 태스크를 불러온다.
func TasksNeedReview(db *sql.DB, show, ctg string) ([]*Task, error) {
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND category=$2 AND status=$3", taskDBKey), show, ctg, StatusNeedReview)
	ts := make([]*Task, 0)
	err := dbQuery(db, stmt, func(rows *sql.Rows) error {
		s := &Task{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		ts = append(ts, s)
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("task", "status-need-review")
		}
		return nil, err
	}
	return ts, nil
}

// ShotTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func ShotTasks(db *sql.DB, id string) ([]*Task, error) {
	show, ctg, unit, err := SplitUnitID(id)
	if err != nil {
		return nil, err
	}
	if ctg != "shot" {
		return nil, fmt.Errorf("invalid shot id: %s", id)
	}
	s, err := GetShot(db, id)
	if err != nil {
		return nil, err
	}
	tasks := s.Tasks
	// DB에 있는 태스크 중 샷의 Tasks에 정의된 태스크만 보이고, 그 순서대로 정렬한다.
	taskNotHidden := make(map[string]bool)
	taskIdx := make(map[string]int)
	for i, task := range tasks {
		taskNotHidden[task] = true
		taskIdx[task] = i
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND category=$2 AND unit=$3", taskDBKey), show, ctg, unit)
	ts := make([]*Task, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		t := &Task{}
		err := scan(rows, t)
		if err != nil {
			return err
		}
		if taskNotHidden[t.Task] {
			ts = append(ts, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(ts, func(i, j int) bool {
		return taskIdx[ts[i].Task] < taskIdx[ts[j].Task]
	})
	return ts, nil
}

// AssetTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func AssetTasks(db *sql.DB, id string) ([]*Task, error) {
	show, ctg, unit, err := SplitUnitID(id)
	if err != nil {
		return nil, err
	}
	if ctg != "asset" {
		return nil, fmt.Errorf("invalid asset id: %s", id)
	}
	a, err := GetAsset(db, id)
	if err != nil {
		return nil, err
	}
	tasks := a.Tasks
	// DB에 있는 태스크 중 샷의 Tasks에 정의된 태스크만 보이고, 그 순서대로 정렬한다.
	taskNotHidden := make(map[string]bool)
	taskIdx := make(map[string]int)
	for i, task := range tasks {
		taskNotHidden[task] = true
		taskIdx[task] = i
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND category=$2 AND unit=$3", taskDBKey), show, ctg, unit)
	ts := make([]*Task, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		t := &Task{}
		err := scan(rows, t)
		if err != nil {
			return err
		}
		if taskNotHidden[t.Task] {
			ts = append(ts, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(ts, func(i, j int) bool {
		return taskIdx[ts[i].Task] < taskIdx[ts[j].Task]
	})
	return ts, nil
}

// UserTasks는 해당 유저의 모든 태스크를 db에서 검색해 반환한다.
func UserTasks(db *sql.DB, user string) ([]*Task, error) {
	// 샷의 tasks에 속하지 않은 태스크는 보이지 않는다.
	keys := ""
	for i, k := range dbKeys(&Task{}) {
		if i != 0 {
			keys += ", "
		}
		keys += "tasks." + k
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks JOIN shots ON (tasks.show = shots.show AND tasks.unit = shots.shot)  WHERE tasks.category='%s' AND tasks.assignee='%s' AND tasks.task = ANY(shots.tasks)", keys, "shot", user))
	tasks := make([]*Task, 0)
	err := dbQuery(db, stmt, func(rows *sql.Rows) error {
		t := &Task{}
		err := scan(rows, t)
		if err != nil {
			return err
		}
		tasks = append(tasks, t)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// DeleteTask는 해당 태스크와 그 하위의 모든 데이터를 db에서 지운다.
// 해당 태스크가 없어도 에러를 내지 않기 때문에 검사를 원한다면 TaskExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteTask(db *sql.DB, id string) error {
	show, ctg, unit, task, err := SplitTaskID(id)
	if err != nil {
		return err
	}
	_, err = GetTask(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM tasks WHERE show=$1 AND category=$2 AND unit=$3 AND task=$4", show, ctg, unit, task),
		dbStmt("DELETE FROM versions WHERE show=$1 AND category=$2 AND unit=$3 AND task=$4", show, ctg, unit, task),
	}
	return dbExec(db, stmts)
}
