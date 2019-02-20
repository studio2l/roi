package roi

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Version은 특정 태스크의 하나의 버전이다.
type Version struct {
	ProjectID string
	ShotID    string
	TaskName  string

	Num         int       // 버전 번호
	OutputFiles []string  // 결과물 경로
	Images      []string  // 결과물을 확인할 수 있는 이미지
	Mov         string    // 결과물을 영상으로 볼 수 있는 경로
	WorkFile    string    // 이 결과물을 만든 작업 파일
	Created     time.Time // 결과물이 만들어진 시간
}

var CreateTableIfNotExistsVersionsStmt = `CREATE TABLE IF NOT EXISTS versions (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project_id STRING NOT NULL CHECK (length(project_id) > 0) CHECK (project_id NOT LIKE '% %'),
	shot_id STRING NOT NULL CHECK (length(shot_id) > 0) CHECK (shot_id NOT LIKE '% %'),
	task_name STRING NOT NULL CHECK (length(task_name) > 0) CHECK (task_name NOT LIKE '% %'),
	num INT NOT NULL,
	output_files STRING[] NOT NULL,
	images STRING[] NOT NULL,
	mov STRING NOT NULL,
	work_file STRING NOT NULL,
	created TIMESTAMPTZ NOT NULL,
	UNIQUE(project_id, shot_id, task_name, num)
)`

var VersionTableKeys = []string{
	"project_id",
	"shot_id",
	"task_name",
	"num",
	"output_files",
	"images",
	"mov",
	"work_file",
	"created",
}

var VersionTableIndices = dbIndices(VersionTableKeys)

func (v *Version) dbValues() []interface{} {
	if v == nil {
		v = &Version{}
	}
	if v.OutputFiles == nil {
		v.OutputFiles = make([]string, 0)
	}
	if v.Images == nil {
		v.Images = make([]string, 0)
	}
	return []interface{}{
		v.ProjectID,
		v.ShotID,
		v.TaskName,
		v.Num,
		pq.Array(v.OutputFiles),
		pq.Array(v.Images),
		v.Mov,
		v.WorkFile,
		v.Created,
	}
}

// AddVersion은 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddVersion(db *sql.DB, prj, shot, task string, v *Version) error {
	if prj == "" {
		return fmt.Errorf("project not specified")
	}
	if shot == "" {
		return fmt.Errorf("shot not specified")
	}
	if task == "" {
		return fmt.Errorf("task not specified")
	}
	if v == nil {
		return fmt.Errorf("nil output")
	}
	if v.Num != 0 {
		// 버전은 DB 확인 후 추가된다.
		return fmt.Errorf("version num should not be specified when adding")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	stmt := fmt.Sprintf("SELECT last_output_version FROM tasks WHERE project_id='%s' AND shot_id='%s' AND name='%s'", prj, shot, task)
	rows, err := tx.Query(stmt)
	if err != nil {
		return fmt.Errorf("could not get last version num of task: %v", err)
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
	rows.Close()
	v.Num = lastv + 1
	keystr := strings.Join(VersionTableKeys, ", ")
	idxstr := strings.Join(VersionTableIndices, ", ")
	stmt = fmt.Sprintf("INSERT INTO versions (%s) VALUES (%s)", keystr, idxstr)
	if _, err := tx.Exec(stmt, v.dbValues()...); err != nil {
		return fmt.Errorf("could not insert versions: %v", err)
	}
	if _, err := tx.Exec("UPDATE tasks SET status=$1, last_output_version=$2 WHERE project_id=$3 AND shot_id=$4 AND name=$5", TaskInProgress, v.Num, prj, shot, task); err != nil {
		return fmt.Errorf("could not update last version num of task: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit the transaction: %v", err)
	}
	return nil
}

// UpdateVersionParam은 Version에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateVersion에서 사용한다.
type UpdateVersionParam struct {
	OutputFiles []string
	Images      []string
	Mov         string
	WorkFile    string
	Created     time.Time
}

func (u UpdateVersionParam) keys() []string {
	return []string{
		"output_files",
		"images",
		"mov",
		"work_file",
		"created",
	}
}

func (u UpdateVersionParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateVersionParam) values() []interface{} {
	if u.OutputFiles == nil {
		u.OutputFiles = make([]string, 0)
	}
	if u.Images == nil {
		u.Images = make([]string, 0)
	}
	return []interface{}{
		pq.Array(u.OutputFiles),
		pq.Array(u.Images),
		u.Mov,
		u.WorkFile,
		u.Created,
	}
}

// UpdateVersion은 db의 특정 태스크를 업데이트 한다.
func UpdateVersion(db *sql.DB, prj, shot, task string, version int, upd UpdateVersionParam) error {
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
		return fmt.Errorf("version num not specified")
	}
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE versions SET (%s) = (%s) WHERE project_id='%s' AND shot_id='%s' AND task_name='%s' AND num='%d'", keystr, idxstr, prj, shot, task, version)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// VersionExist는 db에 해당 태스크가 존재하는지를 검사한다.
func VersionExist(db *sql.DB, prj, shot, task string, version int) (bool, error) {
	if version == 0 {
		// 버전 0은 존재하지 않는다.
		return false, fmt.Errorf("output version not specified")
	}
	stmt := "SELECT num FROM versions WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND num=$4 LIMIT 1"
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// versionFromRows는 테이블의 한 열에서 아웃풋을 받아온다.
func versionFromRows(rows *sql.Rows) (*Version, error) {
	v := &Version{}
	err := rows.Scan(
		&v.ProjectID, &v.ShotID, &v.TaskName,
		&v.Num, pq.Array(&v.OutputFiles), pq.Array(&v.Images), &v.Mov, &v.WorkFile, &v.Created,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetVersion은 db의 특정 태스크의 해당 아웃풋을 찾는다.
// 만일 그 버전의 아웃풋이 없다면 nil이 반환된다.
func GetVersion(db *sql.DB, prj, shot, task string, version int) (*Version, error) {
	keystr := strings.Join(VersionTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM versions WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND num=$4 LIMIT 1", keystr)
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, nil
	}
	return versionFromRows(rows)
}

// AllVersions는 db의 특정 태스크의 아웃풋 전체를 반환한다.
func AllVersions(db *sql.DB, prj, shot, task string) ([]*Version, error) {
	keystr := strings.Join(VersionTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM versions WHERE project_id=$1 AND shot_id=$2 AND task_name=$3", keystr)
	rows, err := db.Query(stmt, prj, shot, task)
	if err != nil {
		return nil, err
	}
	versions := make([]*Version, 0)
	for rows.Next() {
		v, err := versionFromRows(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// DeleteVersion은 해당 버전과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 버전이 없어도 에러를 내지 않기 때문에 검사를 원한다면 VersionExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteVersion(db *sql.DB, prj, shot, task string, version int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM versions WHERE project_id=$1 AND shot_id=$2 AND task_name=$3 AND num=$4", prj, shot, task, version); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	return tx.Commit()
}
