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

// CreateTableIfNotExistShowsStmt는 DB에 versions 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsVersionsStmt = `CREATE TABLE IF NOT EXISTS versions (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	shot STRING NOT NULL CHECK (length(shot) > 0) CHECK (shot NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	version STRING NOT NULL CHECK (length(version) > 0),
	owner STRING NOT NULL CHECK (length(owner) > 0),
	status STRING NOT NULL,
	output_files STRING[] NOT NULL,
	images STRING[] NOT NULL,
	mov STRING NOT NULL,
	work_file STRING NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	UNIQUE(show, shot, task, version)
)`

// Version은 특정 태스크의 하나의 버전이다.
type Version struct {
	Show    string `db:"show"`
	Shot    string `db:"shot"`
	Task    string `db:"task"`
	Version string `db:"version"` // 버전명

	Owner       string        `db:"owner"`        // 버전 소유자
	Status      VersionStatus `db:"status"`       // 버전 상태
	OutputFiles []string      `db:"output_files"` // 결과물 경로
	Images      []string      `db:"images"`       // 결과물을 확인할 수 있는 이미지
	Mov         string        `db:"mov"`          // 결과물을 영상으로 볼 수 있는 경로
	WorkFile    string        `db:"work_file"`    // 이 결과물을 만든 작업 파일
	StartDate   time.Time     `db:"start_date"`   // 버전 작업 시작 시간
	EndDate     time.Time     `db:"end_date"`     // 버전 작업 마감 시간
}

// ID는 Version의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (v *Version) ID() string {
	return v.Show + "/" + v.Shot + "/" + v.Task + "/" + v.Version
}

// ShotID는 부모 샷의 아이디를 반환한다.
func (v *Version) ShotID() string {
	return v.Show + "/" + v.Shot
}

// TaskID는 부모 태스크의 아이디를 반환한다.
func (v *Version) TaskID() string {
	return v.Show + "/" + v.Shot + "/" + v.Task
}

// reVersionName은 가능한 버전명을 정의하는 정규식이다.
var reVersionName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// verifyVersionName은 받아들인 버전 이름이 유효하지 않다면 에러를 반환한다.
func verifyVersionName(version string) error {
	if !reVersionName.MatchString(version) {
		return BadRequest(fmt.Sprintf("invalid version name: %s", version))
	}
	return nil
}

// SplitVersionID는 받아들인 버전 아이디를 쇼, 샷, 태스크, 버전으로 분리해서 반환한다.
// 만일 버전 아이디가 유효하지 않다면 에러를 반환한다.
func SplitVersionID(id string) (string, string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 4 {
		return "", "", "", "", BadRequest(fmt.Sprintf("invalid version id: %s", id))
	}
	show := ns[0]
	shot := ns[1]
	task := ns[2]
	version := ns[3]
	if show == "" || shot == "" || task == "" || version == "" {
		return "", "", "", "", BadRequest(fmt.Sprintf("invalid version id: %s", id))
	}
	return show, shot, task, version, nil
}

// verifyVersionID는 받아들인 버전 아이디가 유효하지 않다면 에러를 반환한다.
func verifyVersionID(id string) error {
	show, shot, task, version, err := SplitVersionID(id)
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
	err = verifyVersionName(version)
	if err != nil {
		return err
	}
	return nil
}

// verifyVersion은 받아들인 버전이 유효하지 않다면 에러를 반환한다.
func verifyVersion(v *Version) error {
	if v == nil {
		return fmt.Errorf("nil version")
	}
	err := verifyVersionID(v.ID())
	if err != nil {
		return err
	}
	err = verifyVersionStatus(v.Status)
	if err != nil {
		return err
	}
	return nil
}

// AddVersion은 db의 특정 프로젝트, 특정 샷에 태스크를 추가한다.
func AddVersion(db *sql.DB, v *Version) error {
	err := verifyVersion(v)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetTask(db, v.TaskID())
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(v)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO versions (%s) VALUES (%s)", keys, idxs), vs...),
		dbStmt("UPDATE tasks SET (working_version, working_version_status) = ($1, $2) WHERE show=$3 AND shot=$4 AND task=$5", v.Version, v.Status, v.Show, v.Shot, v.Task),
	}
	return dbExec(db, stmts)
}

// UpdateVersion은 db의 특정 태스크를 업데이트 한다.
// 이 함수를 호출하기 전 해당 태스크가 존재하는지 사용자가 검사해야 한다.
func UpdateVersion(db *sql.DB, id string, v *Version) error {
	err := verifyVersion(v)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(v)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE versions SET (%s) = (%s) WHERE show='%s' AND shot='%s' AND task='%s' AND version='%s'", keys, idxs, v.Show, v.Shot, v.Task, v.Version), vs...),
	}
	return dbExec(db, stmts)
}

// UpdateVersionStatus은 db의 특정 태스크를 업데이트 한다.
func UpdateVersionStatus(db *sql.DB, id string, status VersionStatus) error {
	show, shot, task, version, err := SplitVersionID(id)
	if err != nil {
		return err
	}
	err = verifyVersionStatus(status)
	if err != nil {
		return err
	}
	t, err := GetTask(db, show+"/"+shot+"/"+task)
	if err != nil {
		return err
	}
	_, err = GetVersion(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("UPDATE versions SET (status) = ($1) WHERE show=$2 AND shot=$3 AND task=$4 AND version=$5", status, show, shot, task, version),
	}
	// 작업중인 버전의 상태는 태스크에도 업데이트 되어야한다.
	if version == t.WorkingVersion {
		stmt := dbStmt("UPDATE tasks SET (working_version_status) = ($1) WHERE show=$2 AND shot=$3 AND task=$4", status, show, shot, task)
		stmts = append(stmts, stmt)
	}
	return dbExec(db, stmts)
}

// GetVersion은 db에서 하나의 버전을 찾는다.
// 해당 버전이 없다면 nil과 NotFound 에러를 반환한다.
func GetVersion(db *sql.DB, id string) (*Version, error) {
	show, shot, task, version, err := SplitVersionID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetTask(db, show+"/"+shot+"/"+task)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2 AND task=$3 AND version=$4 LIMIT 1", keys), show, shot, task, version)
	v := &Version{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, v)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			id := show + "/" + shot + "/" + task + "/" + version
			return nil, NotFound("version", id)
		}
		return nil, err
	}
	return v, err
}

// TaskVersions는 db에서 특정 태스크의 버전 전체를 검색해 반환한다.
func TaskVersions(db *sql.DB, id string) ([]*Version, error) {
	show, shot, task, err := SplitTaskID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetTask(db, id)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2 AND task=$3", keys), show, shot, task)
	versions := make([]*Version, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		v := &Version{}
		err := scan(rows, v)
		if err != nil {
			return err
		}
		versions = append(versions, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(versions, func(i, j int) bool {
		return strings.Compare(versions[i].Version, versions[j].Version) < 0
	})
	return versions, nil
}

// ShotVersions는 db에서 특정 샷의 버전 전체를 검색해 반환한다.
func ShotVersions(db *sql.DB, id string) ([]*Version, error) {
	show, shot, err := SplitShotID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShot(db, id)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Version{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM versions WHERE show=$1 AND shot=$2", keys), show, shot)
	versions := make([]*Version, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		v := &Version{}
		err := scan(rows, v)
		if err != nil {
			return err
		}
		versions = append(versions, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(versions, func(i, j int) bool {
		return strings.Compare(versions[i].Version, versions[j].Version) < 0
	})
	return versions, nil
}

// DeleteVersion은 해당 버전과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 버전이 없어도 에러를 내지 않기 때문에 검사를 원한다면 VersionExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteVersion(db *sql.DB, id string) error {
	show, shot, task, version, err := SplitVersionID(id)
	if err != nil {
		return err
	}
	_, err = GetVersion(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM versions WHERE show=$1 AND shot=$2 AND task=$3 AND version=$4", show, shot, task, version),
	}
	return dbExec(db, stmts)
}
