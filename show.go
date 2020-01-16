package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// CreateTableIfNotExistShowsStmt는 DB에 shows 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
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

// ID는 Show의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Show) ID() string {
	return s.Show
}

// reShowName는 유효한 쇼 이름을 정의하는 정규식이다.
var reShowName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// verifyShowName은 받아들인 쇼 이름이 유효하지 않다면 에러를 반환한다.
func verifyShowName(name string) error {
	if !reShowName.MatchString(name) {
		return BadRequest(fmt.Sprintf("invalid show name: %s", name))
	}
	return nil
}

// verifyShow는 받아들인 쇼가 유효하지 않다면 에러를 반환한다.
func verifyShow(s *Show) error {
	if s == nil {
		return fmt.Errorf("nil show")
	}
	err := verifyShowName(s.Show)
	if err != nil {
		return err
	}
	return nil
}

// AddShow는 db에 쇼를 추가한다.
func AddShow(db *sql.DB, s *Show) error {
	err := verifyShow(s)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("INSERT INTO shows (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// UpdateShow는 db의 쇼 정보를 수정한다.
// 이 함수를 호출하기 전 해당 쇼가 존재하는지 사용자가 검사해야 한다.
func UpdateShow(db *sql.DB, show string, s *Show) error {
	err := verifyShow(s)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("UPDATE shows SET (%s) = (%s) WHERE show='%s'", keys, idxs, show)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// GetShow는 db에서 하나의 쇼를 부른다.
// 해당 쇼가 없다면 nil과 NotFoundError를 반환한다.
func GetShow(db *sql.DB, show string) (*Show, error) {
	if show == "" {
		return nil, BadRequest("show not specified")
	}
	ks, _, _, err := dbKIVs(&Show{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shows WHERE show=$1", keys), show)
	s := &Show{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("show", show)
		}
		return nil, err
	}
	return s, nil
}

// AllShows는 db에서 모든 쇼 정보를 가져온다.
// 검색 중 문제가 있으면 nil, error를 반환한다.
func AllShows(db *sql.DB) ([]*Show, error) {
	ks, _, _, err := dbKIVs(&Show{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	shows := make([]*Show, 0)
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shows", keys))
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		s := &Show{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		shows = append(shows, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return shows, nil
}

// DeleteShow는 해당 쇼와 그 하위의 모든 데이터를 db에서 지운다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteShow(db *sql.DB, show string) error {
	_, err := GetShow(db, show)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM shows WHERE show=$1", show); err != nil {
		return fmt.Errorf("could not delete data from 'shows' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM shots WHERE show=$1", show); err != nil {
		return fmt.Errorf("could not delete data from 'shots' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM tasks WHERE show=$1", show); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE show=$1", show); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	return tx.Commit()
}
