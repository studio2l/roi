package roi

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskNotSet     = TaskStatus("not-set")
	TaskAssigned   = TaskStatus("assigned")
	TaskInProgress = TaskStatus("in-progress")
	TaskAskConfirm = TaskStatus("ask-confirm")
	TaskRetake     = TaskStatus("retake")
	TaskDone       = TaskStatus("done")
	TaskHold       = TaskStatus("hold")
	TaskOmit       = TaskStatus("omit")
)

var AllTaskStatus = []TaskStatus{
	TaskNotSet,
	TaskAssigned,
	TaskInProgress,
	TaskAskConfirm,
	TaskRetake,
	TaskDone,
	TaskHold,
	TaskOmit,
}

// isValidTaskStatus는 해당 태스크 상태가 유효한지를 반환한다.
func isValidTaskStatus(ts TaskStatus) bool {
	for _, s := range AllTaskStatus {
		if ts == s {
			return true
		}
	}
	return false
}

// UIString은 UI안에서 사용하는 현지화된 문자열이다.
// 할일: 한국어 외의 문자열 지원
func (s TaskStatus) UIString() string {
	switch s {
	case TaskNotSet:
		return "-"
	case TaskAssigned:
		return "할당됨"
	case TaskInProgress:
		return "진행중"
	case TaskAskConfirm:
		return "컨펌요청"
	case TaskRetake:
		return "리테이크"
	case TaskDone:
		return "완료"
	case TaskHold:
		return "홀드"
	case TaskOmit:
		return "오밋"
	}
	return ""
}

type Task struct {
	// 관련 아이디
	Show string `db:"show"`
	Shot string `db:"shot"`

	// 태스크 정보
	Task        string     `db:"task"` // 이름은 타입 또는 타입_요소로 구성된다. 예) fx, fx_fire
	Status      TaskStatus `db:"status"`
	Assignee    string     `db:"assignee"`
	LastVersion string     `db:"last_version"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     time.Time  `db:"end_date"`
	DueDate     time.Time  `db:"due_date"`
}

var CreateTableIfNotExistsTasksStmt = `CREATE TABLE IF NOT EXISTS tasks (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	assignee STRING NOT NULL,
	last_version STRING NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	UNIQUE(show, shot, task)
)`

// AddTask는 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddTask(db *sql.DB, show, shot string, t *Task) error {
	if t == nil {
		return BadRequest("nil task")
	}
	if t.Task == "" {
		return BadRequest("task not specified")
	}
	if !isValidTaskStatus(t.Status) {
		return BadRequest(fmt.Sprintf("invalid task status: '%s'", t.Status))
	}
	_, err := GetShot(db, show, shot)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(t)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// UpdateTaskParam은 Task에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateTask에서 사용한다.
type UpdateTaskParam struct {
	Status   TaskStatus `db:"status"`
	Assignee string     `db:"assignee"`
	DueDate  time.Time  `db:"due_date"`
}

// UpdateTask는 db의 특정 태스크를 업데이트 한다.
func UpdateTask(db *sql.DB, show, shot, task string, upd UpdateTaskParam) error {
	if !isValidTaskStatus(upd.Status) {
		return BadRequest(fmt.Sprintf("invalid task status: '%s'", upd.Status))
	}
	_, err := GetTask(db, show, shot, task)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(upd)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("UPDATE tasks SET (%s) = (%s) WHERE show='%s' AND shot='%s' AND task='%s'", keys, idxs, show, shot, task)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// GetTask는 db에서 하나의 태스크를 찾는다.
// 해당 태스크가 없다면 nil과 NotFound 에러를 반환한다.
func GetTask(db *sql.DB, show, shot, task string) (*Task, error) {
	if show == "" {
		return nil, BadRequest("show not specified")
	}
	if shot == "" {
		return nil, BadRequest("shot not specified")
	}
	if task == "" {
		return nil, BadRequest("task not specified")
	}
	_, err := GetShot(db, show, shot)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Task{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND shot=$2 AND task=$3 LIMIT 1", keys)
	rows, err := db.Query(stmt, show, shot, task)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		id := show + "/" + shot + "/" + task
		return nil, NotFound("task", id)
	}
	t := &Task{}
	err = scanFromRows(rows, t)
	return t, err
}

// ShotTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func ShotTasks(db *sql.DB, show, shot string) ([]*Task, error) {
	_, err := GetShot(db, show, shot)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Task{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE show=$1 AND shot=$2", keys)
	rows, err := db.Query(stmt, show, shot)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for rows.Next() {
		t := &Task{}
		err := scanFromRows(rows, t)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
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
	stmt := fmt.Sprintf("SELECT %s FROM tasks JOIN shots ON (tasks.show = shots.show AND tasks.shot = shots.shot)  WHERE tasks.assignee='%s' AND tasks.task = ANY(shots.working_tasks)", keys, user)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for rows.Next() {
		t := &Task{}
		err := scanFromRows(rows, t)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// DeleteTask는 해당 태스크와 그 하위의 모든 데이터를 db에서 지운다.
// 해당 태스크가 없어도 에러를 내지 않기 때문에 검사를 원한다면 TaskExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteTask(db *sql.DB, show, shot, task string) error {
	_, err := GetTask(db, show, shot, task)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM tasks WHERE show=$1 AND shot=$2 AND task=$3", show, shot, task); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE show=$1 AND shot=$2 AND task=$3", show, shot, task); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
