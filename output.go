package roi

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Output은 특정 태스크의 해당 버전 작업 결과물이다.
type Output struct {
	ProjectID string
	ShotID    string
	TaskName  string

	Version  int       // 결과물 버전
	Files    []string  // 결과물 경로
	Images   []string  // 결과물을 확인할 수 있는 이미지
	Mov      string    // 결과물을 영상으로 볼 수 있는 경로
	WorkFile string    // 이 결과물을 만든 작업 파일
	Created  time.Time // 결과물이 만들어진 시간
}

var CreateTableIfNotExistsOutputsStmt = `CREATE TABLE IF NOT EXISTS outputs (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project_id STRING NOT NULL CHECK (length(project_id) > 0) CHECK (project_id NOT LIKE '% %'),
	shot_id STRING NOT NULL CHECK (length(shot_id) > 0) CHECK (shot_id NOT LIKE '% %'),
	task_name STRING NOT NULL CHECK (length(task_name) > 0) CHECK (task_name NOT LIKE '% %'),
	version INT NOT NULL,
	files STRING[] NOT NULL,
	images STRING[] NOT NULL,
	mov STRING NOT NULL,
	work_file STRING NOT NULL,
	created TIMESTAMPTZ NOT NULL,
	UNIQUE(project_id, shot_id, task_name, version)
)`

var OutputTableKeys = []string{
	"project_id",
	"shot_id",
	"task_name",
	"version",
	"files",
	"images",
	"mov",
	"work_file",
	"created",
}

var OutputTableIndices = dbIndices(OutputTableKeys)

func (o *Output) dbValues() []interface{} {
	if o == nil {
		o = &Output{}
	}
	if o.Files == nil {
		o.Files = make([]string, 0)
	}
	if o.Images == nil {
		o.Images = make([]string, 0)
	}
	return []interface{}{
		o.ProjectID,
		o.ShotID,
		o.TaskName,
		o.Version,
		pq.Array(o.Files),
		pq.Array(o.Images),
		o.Mov,
		o.WorkFile,
		o.Created,
	}
}

// AddOutput은 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddOutput(db *sql.DB, prj, shot, task string, o *Output) error {
	if prj == "" {
		return fmt.Errorf("project not specified")
	}
	if shot == "" {
		return fmt.Errorf("shot not specified")
	}
	if task == "" {
		return fmt.Errorf("task not specified")
	}
	if o == nil {
		return fmt.Errorf("nil output")
	}
	if o.Version != 0 {
		// 버전은 DB 확인 후 추가된다.
		return fmt.Errorf("output version should not be specified when adding")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	rows, err := tx.Query("SELECT last_output_version FROM tasks WHERE project_id='%s' AND shot_id='%s' AND name='%s'")
	if err != nil {
		return fmt.Errorf("could not get last output version of task: %v", err)
	}
	exist := rows.Next()
	if rows.Err() != nil {
		return fmt.Errorf("rows.Next error: %v", rows.Err())
	}
	var lastv int
	if exist {
		err := rows.Scan(&lastv)
		if err != nil {
			return fmt.Errorf("rows.Scan error: %v", err)
		}
	}
	o.Version = lastv + 1
	keystr := strings.Join(OutputTableKeys, ", ")
	idxstr := strings.Join(OutputTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO outputs (%s) VALUES (%s)", keystr, idxstr)
	if _, err := tx.Exec(stmt, o.dbValues()...); err != nil {
		return fmt.Errorf("could not insert outputs: %v", err)
	}
	if _, err := tx.Exec("UPDATE tasks SET last_output_version=$1 WHERE project_id=$2 AND shot_id=$3 AND name=$4", o.Version, prj, shot, task); err != nil {
		return fmt.Errorf("could not update last output version of task: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit the transaction: %v", err)
	}
	return nil
}

// UpdateOutputParam은 Output에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateOutput에서 사용한다.
type UpdateOutputParam struct {
	Files    []string
	Images   []string
	Mov      string
	WorkFile string
	Created  time.Time
}

func (u UpdateOutputParam) keys() []string {
	return []string{
		"files",
		"images",
		"mov",
		"work_file",
		"created",
	}
}

func (u UpdateOutputParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateOutputParam) values() []interface{} {
	if u.Files == nil {
		u.Files = make([]string, 0)
	}
	if u.Images == nil {
		u.Images = make([]string, 0)
	}
	return []interface{}{
		pq.Array(u.Files),
		pq.Array(u.Images),
		u.Mov,
		u.WorkFile,
		u.Created,
	}
}

// UpdateOutput은 db의 특정 태스크를 업데이트 한다.
func UpdateOutput(db *sql.DB, prj, shot, task string, version int, upd UpdateOutputParam) error {
	if prj == "" {
		return fmt.Errorf("project not specified")
	}
	if shot == "" {
		return fmt.Errorf("shot not specified")
	}
	if task == "" {
		return fmt.Errorf("task name not specified")
	}
	if version == 0 {
		// 버전 0은 존재하지 않는다.
		return fmt.Errorf("output version not specified")
	}
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE outputs SET (%s) = (%s) WHERE project_id='%s' AND shot_id='%s' AND task_name='%s' AND version='%d'", keystr, idxstr, prj, shot, task, version)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// OutputExist는 db에 해당 태스크가 존재하는지를 검사한다.
func OutputExist(db *sql.DB, prj, shot, task string, version int) (bool, error) {
	if version == 0 {
		// 버전 0은 존재하지 않는다.
		return false, fmt.Errorf("output version not specified")
	}
	stmt := "SELECT version FROM outputs WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND version=$4 LIMIT 1"
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// outputFromRows는 테이블의 한 열에서 아웃풋을 받아온다.
func outputFromRows(rows *sql.Rows) (*Output, error) {
	o := &Output{}
	err := rows.Scan(
		&o.ProjectID, &o.ShotID, &o.TaskName,
		&o.Version, pq.Array(&o.Files), pq.Array(&o.Images), &o.Mov, &o.WorkFile, &o.Created,
	)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// GetOutput은 db의 특정 태스크의 해당 아웃풋을 찾는다.
// 만일 그 버전의 아웃풋이 없다면 nil이 반환된다.
func GetOutput(db *sql.DB, prj, shot, task string, version int) (*Output, error) {
	keystr := strings.Join(OutputTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM outputs WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND version=$4 LIMIT 1", keystr)
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, nil
	}
	return outputFromRows(rows)
}

// AllOutputs는 db의 특정 태스크의 아웃풋 전체를 반환한다.
func AllOutputs(db *sql.DB, prj, shot, task string) ([]*Output, error) {
	keystr := strings.Join(OutputTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM outputs WHERE project_id=$1 AND shot_id=$2 AND task_name=$3", keystr)
	rows, err := db.Query(stmt, prj, shot, task)
	if err != nil {
		return nil, err
	}
	outputs := make([]*Output, 0)
	for rows.Next() {
		o, err := outputFromRows(rows)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, o)
	}
	return outputs, nil
}

// DeleteOutput은 db의 특정 프로젝트에서 아웃풋을 하나 지운다.
func DeleteOutput(db *sql.DB, prj, shot, task string, version int) error {
	stmt := "DELETE FROM outputs WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND version=$4"
	res, err := db.Exec(stmt, prj, shot, task, version)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("output not exist: %s.%s.%s.v%03d", prj, shot, task, version)
	}
	return nil
}
