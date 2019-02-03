package roi

import (
	"database/sql"
	"fmt"
	"strings"
)

type TaskStatus string

const (
	TaskWaiting    = TaskStatus("waiting")
	TaskAssigned   = TaskStatus("assigned")
	TaskInProgress = TaskStatus("in-progress")
	TaskPending    = TaskStatus("pending")
	TaskRetake     = TaskStatus("retake")
	TaskDone       = TaskStatus("done")
	TaskHold       = TaskStatus("hold")
	TaskOmit       = TaskStatus("omit") // 할일: task에 omit이 필요할까?
)

type Task struct {
	// 관련 아이디
	ProjectID string
	ShotID    string

	// 태스크 정보
	Name     string // 이름은 타입 또는 타입_요소로 구성된다. 예) fx, fx_fire
	Status   TaskStatus
	Assignee string
}

func (t *Task) dbValues() []interface{} {
	if t == nil {
		t = &Task{}
	}
	return []interface{}{
		t.ProjectID,
		t.ShotID,
		t.Name,
		t.Status,
		t.Assignee,
	}
}

var CreateTableIfNotExistsTasksStmt = `CREATE TABLE IF NOT EXISTS tasks (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project_id STRING NOT NULL CHECK (length(project_id) > 0) CHECK (project_id NOT LIKE '% %'),
	shot_id STRING NOT NULL CHECK (length(shot_id) > 0) CHECK (shot_id NOT LIKE '% %'),
	name STRING NOT NULL CHECK (length(name) > 0) CHECK (name NOT LIKE '% %'),
	status STRING NOT NULL,
	assignee STRING NOT NULL,
	UNIQUE(project_id, shot_id, name)
)`

var TaskTableKeys = []string{
	"project_id",
	"shot_id",
	"name",
	"status",
	"assignee",
}

var TaskTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5",
}

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
	if t.Name == "" {
		return fmt.Errorf("task name not specified")
	}
	keystr := strings.Join(TaskTableKeys, ", ")
	idxstr := strings.Join(TaskTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO tasks (%s) VALUES (%s)", keystr, idxstr)
	if _, err := db.Exec(stmt, t.dbValues()...); err != nil {
		return err
	}
	return nil
}

// UpdateTaskParam은 Task의 파라미터 중 일반적으로 업데이트 되어야 하는 값으로,
// UpdateTask의 인수로 쓰인다.
type UpdateTaskParam struct {
	Status   TaskStatus
	Assignee string
}

func (u UpdateTaskParam) keys() []string {
	return []string{
		"status",
		"assignee",
	}
}

func (u UpdateTaskParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateTaskParam) values() []interface{} {
	return []interface{}{
		u.Status,
		u.Assignee,
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
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE tasks SET (%s) = (%s) WHERE project_id='%s' AND shot_id='%s' AND name='%s'", keystr, idxstr, prj, shot, task)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// TaskExist는 db에 해당 태스크가 존재하는지를 검사한다.
func TaskExist(db *sql.DB, prj, shot, task string) (bool, error) {
	stmt := "SELECT name FROM tasks WHERE project_id=$1 AND shot_id=$2 AND name=$3 LIMIT 1"
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
		&t.ProjectID, &t.ShotID,
		&t.Name, &t.Status, &t.Assignee,
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
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE project_id=$1 AND shot_id=$2 AND name=$3 LIMIT 1", keystr)
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

// AllTasks는 db의 특정 프로젝트 특정 샷의 태스크 전체를 반환한다.
func AllTasks(db *sql.DB, prj, shot string) ([]*Task, error) {
	keystr := strings.Join(TaskTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM tasks WHERE project_id=$1 AND shot_id=$2", keystr)
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

// DeleteTask는 db의 특정 프로젝트에서 태스크를 하나 지운다.
func DeleteTask(db *sql.DB, prj, shot, task string) error {
	stmt := "DELETE FROM tasks WHERE project_id=$1 AND shot_id=$2 AND name=$3"
	res, err := db.Exec(stmt, prj, shot, task)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not exist: %s.%s.%s", prj, shot, task)
	}
	return nil
}
