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
	Show string `db:"show"`
	Shot string `db:"shot"`
	Task string `db:"task"`

	Version     string    `db:"version"`      // 버전명
	OutputFiles []string  `db:"output_files"` // 결과물 경로
	Images      []string  `db:"images"`       // 결과물을 확인할 수 있는 이미지
	Mov         string    `db:"mov"`          // 결과물을 영상으로 볼 수 있는 경로
	WorkFile    string    `db:"work_file"`    // 이 결과물을 만든 작업 파일
	Created     time.Time `db:"created"`      // 결과물이 만들어진 시간
}

var CreateTableIfNotExistsVersionsStmt = `CREATE TABLE IF NOT EXISTS versions (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	version STRING NOT NULL CHECK (length(version) > 0),
	output_files STRING[] NOT NULL,
	images STRING[] NOT NULL,
	mov STRING NOT NULL,
	work_file STRING NOT NULL,
	created TIMESTAMPTZ NOT NULL,
	UNIQUE(show, shot, task, version)
)`

// AddVersion은 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddVersion(db *sql.DB, prj, shot, task string, v *Version) error {
	if prj == "" {
		return BadRequest("show not specified")
	}
	if shot == "" {
		return BadRequest("shot not specified")
	}
	if task == "" {
		return fmt.Errorf("task not specified")
	}
	if v == nil {
		return fmt.Errorf("nil version")
	}
	ks, vs, err := dbKVs(v)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(dbIndices(ks), ", ")
	tx, err := db.Begin()
	if err != nil {
		return Internal(fmt.Errorf("could not begin a transaction: %v", err))
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	stmt := fmt.Sprintf("INSERT INTO versions (%s) VALUES (%s)", keys, idxs)
	if _, err := tx.Exec(stmt, vs...); err != nil {
		return Internal(fmt.Errorf("could not insert versions: %v", err))
	}
	if _, err := tx.Exec("UPDATE tasks SET status=$1, last_version=$2 WHERE show=$3 AND shot=$4 AND task=$5", TaskInProgress, v.Version, prj, shot, task); err != nil {
		return Internal(fmt.Errorf("could not update last version of task: %v", err))
	}
	err = tx.Commit()
	if err != nil {
		return Internal(fmt.Errorf("could not commit the transaction: %v", err))
	}
	return nil
}

// UpdateVersionParam은 Version에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateVersion에서 사용한다.
type UpdateVersionParam struct {
	OutputFiles []string  `db:"output_files"`
	Images      []string  `db:"images"`
	Mov         string    `db:"mov"`
	WorkFile    string    `db:"work_file"`
	Created     time.Time `db:"created"`
}

// UpdateVersion은 db의 특정 태스크를 업데이트 한다.
func UpdateVersion(db *sql.DB, prj, shot, task string, version string, upd UpdateVersionParam) error {
	if prj == "" {
		return BadRequest("show not specified")
	}
	if shot == "" {
		return BadRequest("shot not specified")
	}
	if task == "" {
		return BadRequest("task not specified")
	}
	if version == "" {
		return BadRequest("version not specified")
	}
	ks, vs, err := dbKVs(upd)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(dbIndices(ks), ", ")
	stmt := fmt.Sprintf("UPDATE versions SET (%s) = (%s) WHERE show='%s' AND shot='%s' AND task='%s' AND version='%s'", keys, idxs, prj, shot, task, version)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return Internal(err)
	}
	return nil
}

// VersionExist는 db에 해당 태스크가 존재하는지를 검사한다.
func VersionExist(db *sql.DB, prj, shot, task string, version string) (bool, error) {
	if prj == "" {
		return false, BadRequest("show not specified")
	}
	if shot == "" {
		return false, BadRequest("shot not specified")
	}
	if task == "" {
		return false, BadRequest("task not specified")
	}
	if version == "" {
		return false, BadRequest("version not specified")
	}
	stmt := "SELECT version FROM versions WHERE show=$1 AND shot=$2 AND task=$3 AND version=$4 LIMIT 1"
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return false, Internal(err)
	}
	return rows.Next(), nil
}

// versionFromRows는 테이블의 한 열에서 아웃풋을 받아온다.
func versionFromRows(rows *sql.Rows) (*Version, error) {
	v := &Version{}
	err := rows.Scan(
		&v.Show, &v.Shot, &v.Task,
		&v.Version, pq.Array(&v.OutputFiles), pq.Array(&v.Images), &v.Mov, &v.WorkFile, &v.Created,
	)
	if err != nil {
		return nil, Internal(err)
	}
	return v, nil
}

// GetVersion은 db에서 하나의 버전을 찾는다.
// 해당 버전이 없다면 nil과 NotFound 에러를 반환한다.
func GetVersion(db *sql.DB, prj, shot, task string, version string) (*Version, error) {
	ks, _, err := dbKVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2 AND task=$3 AND version=$4 LIMIT 1", keys)
	rows, err := db.Query(stmt, prj, shot, task, version)
	if err != nil {
		return nil, Internal(err)
	}
	ok := rows.Next()
	if !ok {
		id := prj + "/" + shot + "/" + task + "/" + version
		return nil, NotFound("version", id)
	}
	return versionFromRows(rows)
}

// TaskVersions는 db에서 특정 태스크의 버전 전체를 검색해 반환한다.
func TaskVersions(db *sql.DB, prj, shot, task string) ([]*Version, error) {
	ks, _, err := dbKVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2 AND task=$3", keys)
	rows, err := db.Query(stmt, prj, shot, task)
	if err != nil {
		return nil, Internal(err)
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

// ShotVersions는 db에서 특정 샷의 버전 전체를 검색해 반환한다.
func ShotVersions(db *sql.DB, prj, shot string) ([]*Version, error) {
	ks, _, err := dbKVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2", keys)
	rows, err := db.Query(stmt, prj, shot)
	if err != nil {
		return nil, Internal(err)
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
func DeleteVersion(db *sql.DB, prj, shot, task string, version string) error {
	tx, err := db.Begin()
	if err != nil {
		return Internal(fmt.Errorf("could not begin a transaction: %w", err))
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM versions WHERE show=$1 AND shot=$2 AND task=$3 AND version=$4", prj, shot, task, version); err != nil {
		return Internal(fmt.Errorf("could not delete data from 'versions' table: %w", err))
	}
	err = tx.Commit()
	if err != nil {
		return Internal(err)
	}
	return nil
}
