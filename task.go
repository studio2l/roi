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
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	due_date TIMESTAMPTZ NOT NULL,
	assignee STRING NOT NULL,
	publish_version STRING NOT NULL,
	working_version STRING NOT NULL,
	working_version_status STRING NOT NULL,
	UNIQUE(show, shot, task)
)`

type Task struct {
	// 관련 아이디
	Show string `db:"show"`
	Shot string `db:"shot"`
	Task string `db:"task"` // 타입 또는 타입_요소로 구성된다. 예) fx, fx_fire

	Status   TaskStatus `db:"status"`
	DueDate  time.Time  `db:"due_date"`
	Assignee string     `db:"assignee"`

	// WorkingVersion은 현재 작업중인 버전이다.
	WorkingVersion string `db:"working_version"`
	// WorkingVersionStatus는 현재 작업중인 버전의 상태이다.
	// 물론 버전의 상태를 검사해보면 얻을수 있지만 샷 페이지에 띄워야 할
	// 중요한 정보이기 때문에 태스크에도 함께 저장한다.
	// 둘은 항상 싱크가 되어있어야 한다.
	WorkingVersionStatus VersionStatus `db:"working_version_status"`

	// PublishVersion은 퍼블리시된 버전이다.
	PublishVersion string `db:"publish_version"`
}

// ID는 Task의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (t *Task) ID() string {
	return t.Show + "/" + t.Shot + "/" + t.Task
}

// ShotID는 부모 샷의 아이디를 반환한다.
func (t *Task) ShotID() string {
	return t.Show + "/" + t.Shot
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

// SplitTaskID는 받아들인 샷 아이디를 쇼, 샷, 태스크로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitTaskID(id string) (string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	show := ns[0]
	shot := ns[1]
	task := ns[2]
	if show == "" || shot == "" || task == "" {
		return "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	return show, shot, task, nil
}

// verifyTaskID는 받아들인 태스크 아이디가 유효하지 않다면 에러를 반환한다.
func verifyTaskID(id string) error {
	show, shot, task, err := SplitTaskID(id)
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
	err = verifyTaskName(task)
	if err != nil {
		return err
	}
	return nil
}

// verifyTask는 받아들인 태스크가 유효하지 않다면 에러를 반환한다.
func verifyTask(t *Task) error {
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
	return nil
}

// AddTask는 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddTask(db *sql.DB, t *Task) error {
	err := verifyTask(t)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetShot(db, t.ShotID())
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(t)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", keys, idxs), vs...),
	}
	return dbExec(db, stmts)
}

// addTaskStmts는 태스크를 추가하는 db 구문을 반환한다.
// 부모가 있는지는 검사하지 않는다.
func addTaskStmts(t *Task) ([]dbStatement, error) {
	err := verifyTask(t)
	if err != nil {
		return nil, err
	}
	ks, is, vs, err := dbKIVs(t)
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", keys, idxs), vs...),
	}
	return stmts, nil
}

// UpdateTask는 db의 특정 태스크를 업데이트 한다.
// 이 함수를 호출하기 전 해당 태스크가 존재하는지 사용자가 검사해야 한다.
func UpdateTask(db *sql.DB, id string, t *Task) error {
	err := verifyTask(t)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(t)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE tasks SET (%s) = (%s) WHERE show='%s' AND shot='%s' AND task='%s'", keys, idxs, t.Show, t.Shot, t.Task), vs...),
	}
	return dbExec(db, stmts)
}

// UpdateTaskWorkingVersion는 db의 특정 태스크의 현재 작업중인 버전을 업데이트 한다.
func UpdateTaskWorkingVersion(db *sql.DB, id, version string) error {
	show, shot, task, err := SplitTaskID(id)
	if err != nil {
		return err
	}
	v, err := GetVersion(db, id+"/"+version)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("UPDATE tasks SET (working_version) = ($1) WHERE show=$2 AND shot=$3 AND task=$4", version, show, shot, task),
		dbStmt("UPDATE tasks SET (working_version_status) = ($1) WHERE show=$2 AND shot=$3 AND task=$4", v.Status, show, shot, task),
	}
	return dbExec(db, stmts)
}

// UpdateTaskPublishVersion는 db의 특정 태스크의 현재 퍼블리시 버전을 업데이트 한다.
func UpdateTaskPublishVersion(db *sql.DB, id, version string) error {
	show, shot, task, err := SplitTaskID(id)
	if err != nil {
		return err
	}
	t, err := GetTask(db, id)
	if err != nil {
		return err
	}
	_, err = GetVersion(db, id+"/"+version)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("UPDATE tasks SET (publish_version) = ($1) WHERE show=$2 AND shot=$3 AND task=$4", version, show, shot, task),
	}
	if version == t.WorkingVersion {
		stmts = append(stmts, dbStmt("UPDATE tasks SET (working_version) = ('') WHERE show=$1 AND shot=$2 AND task=$3", show, shot, task))
	}
	return dbExec(db, stmts)
}

// GetTask는 db에서 하나의 태스크를 찾는다.
// 해당 태스크가 없다면 nil과 NotFound 에러를 반환한다.
func GetTask(db *sql.DB, id string) (*Task, error) {
	show, shot, task, err := SplitTaskID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShot(db, show+"/"+shot)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Task{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND shot=$2 AND task=$3 LIMIT 1", keys), show, shot, task)
	t := &Task{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, t)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			id := show + "/" + shot + "/" + task
			return nil, NotFound("task", id)
		}
		return nil, err
	}
	return t, err
}

// ShotTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func ShotTasks(db *sql.DB, id string) ([]*Task, error) {
	show, shot, err := SplitShotID(id)
	if err != nil {
		return nil, err
	}
	s, err := GetShot(db, show+"/"+shot)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Task{})
	if err != nil {
		return nil, err
	}
	// DB에 있는 태스크 중 샷의 Tasks에 정의된 태스크만 보이고, 그 순서대로 정렬한다.
	taskNotHidden := make(map[string]bool)
	taskIdx := make(map[string]int)
	for i, task := range s.Tasks {
		taskNotHidden[task] = true
		taskIdx[task] = i
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND shot=$2", keys), show, shot)
	tasks := make([]*Task, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		t := &Task{}
		err := scan(rows, t)
		if err != nil {
			return err
		}
		if taskNotHidden[t.Task] {
			tasks = append(tasks, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		return taskIdx[tasks[i].Task] < taskIdx[tasks[j].Task]
	})
	return tasks, nil
}

// UserTasks는 해당 유저의 모든 태스크를 db에서 검색해 반환한다.
func UserTasks(db *sql.DB, user string) ([]*Task, error) {
	// 샷의 working_tasks에 속하지 않은 태스크는 보이지 않는다.
	ks, _, _, err := dbKIVs(&Task{})
	if err != nil {
		return nil, err
	}
	keys := ""
	for i := range ks {
		if i != 0 {
			keys += ", "
		}
		keys += "tasks." + ks[i]
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM tasks JOIN shots ON (tasks.show = shots.show AND tasks.shot = shots.shot)  WHERE tasks.assignee='%s' AND tasks.task = ANY(shots.working_tasks)", keys, user))
	tasks := make([]*Task, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
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
	show, shot, task, err := SplitTaskID(id)
	if err != nil {
		return err
	}
	_, err = GetTask(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM tasks WHERE show=$1 AND shot=$2 AND task=$3", show, shot, task),
		dbStmt("DELETE FROM versions WHERE show=$1 AND shot=$2 AND task=$3", show, shot, task),
	}
	return dbExec(db, stmts)
}
