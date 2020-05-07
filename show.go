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

// CreateTableIfNotExistShowsStmt는 DB에 shows 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsShowsStmt = `CREATE TABLE IF NOT EXISTS shows (
	show STRING NOT NULL UNIQUE CHECK (LENGTH(show) > 0) CHECK (show NOT LIKE '% %'),
	status STRING NOT NULL,
	supervisor STRING NOT NULL,
	cg_supervisor STRING NOT NULL,
	pd STRING NOT NULL,
	managers STRING[] NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	default_shot_tasks STRING[] NOT NULL,
	default_asset_tasks STRING[] NOT NULL,
	tags STRING[] NOT NULL,
	notes STRING NOT NULL,
	attrs STRING NOT NULL,
	CONSTRAINT shows_pk PRIMARY KEY (show)
)`

type Show struct {
	// 쇼 아이디. 로이 내에서 고유해야 한다.
	Show string `db:"show"`

	Status string `db:"status"`

	Supervisor   string   `db:"supervisor"`
	CGSupervisor string   `db:"cg_supervisor"`
	PD           string   `db:"pd"`
	Managers     []string `db:"managers"`

	DueDate           time.Time `db:"due_date"`
	DefaultShotTasks  []string  `db:"default_shot_tasks"`
	DefaultAssetTasks []string  `db:"default_asset_tasks"`
	Tags              []string  `db:"tags"`
	Notes             string    `db:"notes"`

	// Attrs는 커스텀 속성으로 db에는 여러줄의 문자열로 저장된다. 각 줄은 키: 값의 쌍이다.
	Attrs DBStringMap `db:"attrs"`
}

var showDBKey string = strings.Join(dbKeys(&Show{}), ", ")
var showDBIdx string = strings.Join(dbIdxs(&Show{}), ", ")
var _ []interface{} = dbVals(&Show{})

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
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyShow(db *sql.DB, s *Show) error {
	if s == nil {
		return fmt.Errorf("nil show")
	}
	err := verifyShowName(s.Show)
	if err != nil {
		return err
	}
	sort.Slice(s.Tags, func(i, j int) bool {
		return strings.Compare(s.Tags[i], s.Tags[j]) <= 0
	})
	return nil
}

// AddShow는 db에 쇼를 추가한다.
func AddShow(db *sql.DB, s *Show) error {
	err := verifyShow(db, s)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO shows (%s) VALUES (%s)", showDBKey, showDBIdx), dbVals(s)...),
	}
	return dbExec(db, stmts)
}

// UpdateShow는 db의 쇼 정보를 수정한다.
// 이 함수를 호출하기 전 해당 쇼가 존재하는지 사용자가 검사해야 한다.
func UpdateShow(db *sql.DB, show string, s *Show) error {
	err := verifyShow(db, s)
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf("UPDATE shows SET (%s) = (%s) WHERE show='%s'", showDBKey, showDBIdx, show)
	if _, err := db.Exec(stmt, dbVals(s)...); err != nil {
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
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shows WHERE show=$1", showDBKey), show)
	s := &Show{}
	err := dbQueryRow(db, stmt, func(row *sql.Row) error {
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
	shows := make([]*Show, 0)
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM shows", showDBKey))
	err := dbQuery(db, stmt, func(rows *sql.Rows) error {
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
	stmts := []dbStatement{
		dbStmt("DELETE FROM shows WHERE show=$1", show),
		dbStmt("DELETE FROM groups WHERE show=$1", show),
		dbStmt("DELETE FROM units WHERE show=$1", show),
		dbStmt("DELETE FROM tasks WHERE show=$1", show),
		dbStmt("DELETE FROM versions WHERE show=$1", show),
	}
	return dbExec(db, stmts)
}
