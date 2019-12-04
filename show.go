package roi

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ShowStatus string

const (
	ShowWaiting        = ShowStatus("")
	ShowPreProduction  = ShowStatus("pre")
	ShowProduction     = ShowStatus("prod")
	ShowPostProduction = ShowStatus("post")
	ShowDone           = ShowStatus("done")
	ShowHold           = ShowStatus("hold")
)

var AllShowStatus = []ShowStatus{
	ShowWaiting,
	ShowPreProduction,
	ShowProduction,
	ShowPostProduction,
	ShowDone,
	ShowHold,
}

// isValidShowStatus는 해당 태스크 상태가 유효한지를 반환한다.
func isValidShowStatus(ss ShowStatus) bool {
	for _, s := range AllShowStatus {
		if ss == s {
			return true
		}
	}
	return false
}

// UIString은 UI안에서 사용하는 현지화된 문자열이다.
// 할일: 한국어 외의 문자열 지원
func (s ShowStatus) UIString() string {
	switch s {
	case ShowWaiting:
		return "대기"
	case ShowPreProduction:
		return "프리 프로덕션"
	case ShowProduction:
		return "프로덕션"
	case ShowPostProduction:
		return "포스트 프로덕션"
	case ShowDone:
		return "완료"
	case ShowHold:
		return "홀드"
	}
	return ""
}

// UIColor는 UI안에서 사용하는 색상이다.
func (s ShowStatus) UIColor() string {
	switch s {
	case ShowWaiting:
		return ""
	case ShowPreProduction:
		return "yellow"
	case ShowProduction:
		return "yellow"
	case ShowPostProduction:
		return "green"
	case ShowDone:
		return "blue"
	case ShowHold:
		return "gray"
	}
	return ""
}

var reValidShow = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsValidShow(id string) bool {
	return reValidShow.MatchString(id)
}

type Show struct {
	// 쇼 아이디. 로이 내에서 고유해야 한다.
	Show string `db:"show"`

	Name   string `db:"name"`
	Status string `db:"status"`

	Client        string `db:"client"`
	Director      string `db:"director"`
	Producer      string `db:"producer"`
	VFXSupervisor string `db:"vfx_supervisor"`
	VFXManager    string `db:"vfx_manager"`
	CGSupervisor  string `db:"cg_supervisor"`

	CrankIn     time.Time `db:"crank_in"`
	CrankUp     time.Time `db:"crank_up"`
	StartDate   time.Time `db:"start_date"`
	ReleaseDate time.Time `db:"release_date"`
	VFXDueDate  time.Time `db:"vfx_due_date"`

	OutputSize   string   `db:"output_size"`
	ViewLUT      string   `db:"view_lut"`
	DefaultTasks []string `db:"default_tasks"`
}

var CreateTableIfNotExistsShowsStmt = `CREATE TABLE IF NOT EXISTS shows (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	show STRING NOT NULL UNIQUE CHECK (LENGTH(show) > 0) CHECK (show NOT LIKE '% %'),
	name STRING NOT NULL,
	status STRING NOT NULL,
	client STRING NOT NULL,
	director STRING NOT NULL,
	producer STRING NOT NULL,
	vfx_supervisor STRING NOT NULL,
	vfx_manager STRING NOT NULL,
	cg_supervisor STRING NOT NULL,
	crank_in TIMESTAMPTZ NOT NULL,
	crank_up TIMESTAMPTZ NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	release_date TIMESTAMPTZ NOT NULL,
	vfx_due_date TIMESTAMPTZ NOT NULL,
	output_size STRING NOT NULL,
	view_lut STRING NOT NULL,
	default_tasks STRING[] NOT NULL
)`

// AddShow는 db에 쇼를 추가한다.
func AddShow(db *sql.DB, p *Show) error {
	if p == nil {
		return BadRequest("nil show is invalid")
	}
	if !IsValidShow(p.Show) {
		return BadRequest(fmt.Sprintf("invalid show id: %s", p.Show))
	}
	ks, vs, err := dbKVs(p)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(dbIndices(ks), ", ")
	stmt := fmt.Sprintf("INSERT INTO shows (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// UpdateShowParam은 Show에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateShow에서 사용한다.
type UpdateShowParam struct {
	Name          string    `db:"name"`
	Status        string    `db:"status"`
	Client        string    `db:"client"`
	Director      string    `db:"director"`
	Producer      string    `db:"producer"`
	VFXSupervisor string    `db:"vfx_supervisor"`
	VFXManager    string    `db:"vfx_manager"`
	CGSupervisor  string    `db:"cg_supervisor"`
	CrankIn       time.Time `db:"crank_in"`
	CrankUp       time.Time `db:"crank_up"`
	StartDate     time.Time `db:"start_date"`
	ReleaseDate   time.Time `db:"release_date"`
	VFXDueDate    time.Time `db:"vfx_due_date"`
	OutputSize    string    `db:"output_size"`
	ViewLUT       string    `db:"view_lut"`
	DefaultTasks  []string  `db:"default_tasks"`
}

// UpdateShow는 db의 쇼 정보를 수정한다.
func UpdateShow(db *sql.DB, prj string, upd UpdateShowParam) error {
	if !IsValidShow(prj) {
		return BadRequest(fmt.Sprintf("invalid show id: %s", prj))
	}
	ks, vs, err := dbKVs(upd)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(dbIndices(ks), ", ")
	stmt := fmt.Sprintf("UPDATE shows SET (%s) = (%s) WHERE show='%s'", keys, idxs, prj)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// ShowExist는 db에 해당 쇼가 존재하는지를 검사한다.
func ShowExist(db *sql.DB, prj string) (bool, error) {
	rows, err := db.Query("SELECT show FROM shows WHERE show=$1 LIMIT 1", prj)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// showFromRows는 테이블의 한 열에서 쇼를 받아온다.
func showFromRows(rows *sql.Rows) (*Show, error) {
	p := &Show{}
	err := rows.Scan(
		&p.Show, &p.Name, &p.Status, &p.Client,
		&p.Director, &p.Producer, &p.VFXSupervisor, &p.VFXManager, &p.CGSupervisor,
		&p.CrankIn, &p.CrankUp, &p.StartDate, &p.ReleaseDate, &p.VFXDueDate, &p.OutputSize,
		&p.ViewLUT, pq.Array(&p.DefaultTasks),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetShow는 db에서 하나의 쇼를 부른다.
// 해당 쇼가 없다면 nil과 NotFoundError를 반환한다.
func GetShow(db *sql.DB, prj string) (*Show, error) {
	ks, _, err := dbKVs(&Show{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM shows WHERE show=$1", keys)
	rows, err := db.Query(stmt, prj)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, NotFound("show", prj)
	}
	p, err := showFromRows(rows)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// AllShows는 db에서 모든 쇼 정보를 가져온다.
// 검색 중 문제가 있으면 nil, error를 반환한다.
func AllShows(db *sql.DB) ([]*Show, error) {
	ks, _, err := dbKVs(&Show{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM shows", keys)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	prjs := make([]*Show, 0)
	for rows.Next() {
		p, err := showFromRows(rows)
		if err != nil {
			return nil, err
		}
		prjs = append(prjs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prjs, nil
}

// DeleteShow는 해당 쇼와 그 하위의 모든 데이터를 db에서 지운다.
// 해당 쇼가 없어도 에러를 내지 않기 때문에 검사를 원한다면 ShowExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteShow(db *sql.DB, prj string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM shows WHERE show=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'shows' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM shots WHERE show=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'shots' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM tasks WHERE show=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE show=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	return tx.Commit()
}
