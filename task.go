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
	Project string
	Shot    string

	// 태스크 정보
	Task              string // 이름은 타입 또는 타입_요소로 구성된다. 예) fx, fx_fire
	Status            TaskStatus
	Assignee          string
	LastOutputVersion int
	StartDate         time.Time
	EndDate           time.Time
	DueDate           time.Time
}

func (t *Task) dbValues() []interface{} {
	if t == nil {
		t = &Task{}
	}
	return []interface{}{
		t.Project,
		t.Shot,
		t.Task,
		t.Status,
		t.Assignee,
		t.LastOutputVersion,
		t.StartDate,
		t.EndDate,
		t.DueDate,
	}
}

var CreateTableIfNotExistsTasksStmt = `CREATE TABLE IF NOT EXISTS tasks (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project STRING NOT NULL CHECK (length(project) > 0) CHECK (project NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	assignee STRING NOT NULL,
	last_output_version INT NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	UNIQUE(project, shot, task)
)`

var TaskTableKeys = []string{
	"project",
	"shot",
	"task",
	"status",
	"assignee",
	"last_output_version",
	"start_date",
	"end_date",
	"due_date",
}

var TaskTableIndices = dbIndices(TaskTableKeys)

// AddTask는 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddTask(db *sql.DB, prj, shot string, t *Task) error {
	if prj == "" {
		return fmt.Errorf("project not specified")
	}
	if shot == "" {
		return fmt.Errorf("shot not specified")
	}
	if t == nil {
		return fmt.Errorf("nil task")
	}
	if t.Task == "" {
		return fmt.Errorf("task name not specified")
	}
	if !isValidTaskStatus(t.Status) {
		return fmt.Errorf("invalid task status: '%s'", t.Status)
	}
	keystr := strings.Join(TaskTableKeys, ", ")
	idxstr := strings.Join(TaskTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", keystr, idxstr)
	if _, err := db.Exec(stmt, t.dbValues()...); err != nil {
		return err
	}
	return nil
}

// UpdateTaskParam은 Task에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateTask에서 사용한다.
type UpdateTaskParam struct {
	Status   TaskStatus
	Assignee string
	DueDate  time.Time
}

func (u UpdateTaskParam) keys() []string {
	return []string{
		"status",
		"assignee",
		"due_date",
	}
}

func (u UpdateTaskParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateTaskParam) values() []interface{} {
	return []interface{}{
		u.Status,
		u.Assignee,
		u.DueDate,
	}
}

// UpdateTask는 db의 특정 태스크를 업데이트 한다.
func UpdateTask(db *sql.DB, prj, shot, task string, upd UpdateTaskParam) error {
	if prj == "" {
		return fmt.Errorf("project not specified")
	}
	if shot == "" {
		return fmt.Errorf("shot not specified")
	}
	if task == "" {
		return fmt.Errorf("task name not specified")
	}
	if !isValidTaskStatus(upd.Status) {
		return fmt.Errorf("invalid task status: '%s'", upd.Status)
	}
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE tasks SET (%s) = (%s) WHERE project='%s' AND shot='%s' AND task='%s'", keystr, idxstr, prj, shot, task)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// TaskExist는 db에 해당 태스크가 존재하는지를 검사한다.
func TaskExist(db *sql.DB, prj, shot, task string) (bool, error) {
	stmt := "SELECT task FROM tasks WHERE project=$1 AND shot=$2 AND task=$3 LIMIT 1"
	rows, err := db.Query(stmt, prj, shot, task)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// taskFromRows는 테이블의 한 열에서 태스크를 받아온다.
func taskFromRows(rows *sql.Rows) (*Task, error) {
	t := &Task{}
	err := rows.Scan(
		&t.Project, &t.Shot,
		&t.Task, &t.Status, &t.Assignee, &t.LastOutputVersion,
		&t.StartDate, &t.EndDate, &t.DueDate,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// GetTask는 db의 특정 프로젝트에서 해당 태스크를 찾는다.
// 만일 그 이름의 태스크가 없다면 nil이 반환된다.
func GetTask(db *sql.DB, prj, shot, task string) (*Task, error) {
	keystr := strings.Join(TaskTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE project=$1 AND shot=$2 AND task=$3 LIMIT 1", keystr)
	rows, err := db.Query(stmt, prj, shot, task)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, nil
	}
	return taskFromRows(rows)
}

// ShotTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func ShotTasks(db *sql.DB, prj, shot string) ([]*Task, error) {
	keystr := strings.Join(TaskTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE project=$1 AND shot=$2", keystr)
	rows, err := db.Query(stmt, prj, shot)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for rows.Next() {
		t, err := taskFromRows(rows)
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
	keystr := ""
	for i, k := range TaskTableKeys {
		if i != 0 {
			keystr += ", "
		}
		keystr += "tasks." + k
	}
	stmt := fmt.Sprintf("SELECT %s FROM tasks JOIN shots ON (tasks.project = shots.project AND tasks.shot = shots.shot)  WHERE tasks.assignee='%s' AND tasks.task = ANY(shots.working_tasks)", keystr, user)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for rows.Next() {
		t, err := taskFromRows(rows)
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
func DeleteTask(db *sql.DB, prj, shot, task string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM tasks WHERE project=$1 AND shot=$2 AND task=$3", prj, shot, task); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE project=$1 AND shot=$2 AND task=$3", prj, shot, task); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	return tx.Commit()
}
