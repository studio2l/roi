package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// CreateTableIfNotExistGroupsStmt는 DB에 groups 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsGroupsStmt = `CREATE TABLE IF NOT EXISTS groups (
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	grp STRING NOT NULL CHECK (length(grp) > 0) CHECK (grp NOT LIKE '% %'),
	default_tasks STRING[] NOT NULL,
	notes STRING NOT NULL,
	attrs STRING NOT NULL,
	UNIQUE(show, grp),
	CONSTRAINT groups_pk PRIMARY KEY (show, grp)
)`

type Group struct {
	Show  string `db:"show"`
	Group string `db:"grp"` // group이 sql 구문이기 때문에 줄여서 씀.

	DefaultTasks []string `db:"default_tasks"`

	Notes string `db:"notes"`

	// Attrs는 커스텀 속성으로 db에는 여러줄의 문자열로 저장된다. 각 줄은 키: 값의 쌍이다.
	Attrs DBStringMap `db:"attrs"`
}

var groupDBKey string = strings.Join(dbKeys(&Group{}), ", ")
var groupDBIdx string = strings.Join(dbIdxs(&Group{}), ", ")
var _ []interface{} = dbVals(&Group{})

// ID는 Group의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Group) ID() string {
	return s.Show + "/" + s.Group
}

// SplitGroupID는 받아들인 샷 아이디를 쇼, 샷으로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitGroupID(id string) (string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 2 {
		return "", "", BadRequest(fmt.Sprintf("invalid group id: %s", id))
	}
	show := ns[0]
	group := ns[1]
	if show == "" || group == "" {
		return "", "", BadRequest(fmt.Sprintf("invalid group id: %s", id))
	}
	return show, group, nil
}

func JoinGroupID(show, grp string) string {
	return show + "/" + grp
}

func verifyGroupPrimaryKeys(show, grp string) error {
	err := verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyGroupName(grp)
	if err != nil {
		return err
	}
	return nil
}

var (
	// reGroupName은 유효한 샷 이름을 나타내는 정규식이다.
	reGroupName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+$`)
)

// verifyGroupame은 받아들인 샷 이름이 유효하지 않다면 에러를 반환한다.
func verifyGroupName(group string) error {
	if !reGroupName.MatchString(group) {
		return BadRequest(fmt.Sprintf("invalid group name: %s", group))
	}
	return nil
}

// verifyGroup은 받아들인 샷이 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyGroup(db *sql.DB, s *Group) error {
	if s == nil {
		return fmt.Errorf("nil group")
	}
	err := verifyGroupPrimaryKeys(s.Show, s.Group)
	if err != nil {
		return err
	}
	si, err := GetSite(db)
	if err != nil {
		return err
	}
	hasTask := make(map[string]bool)
	for _, t := range si.Tasks {
		hasTask[t] = true
	}
	for _, t := range s.DefaultTasks {
		if !hasTask[t] {
			return fmt.Errorf("task not defined in site: %s", t)
		}
	}
	return nil
}

// AddGroup은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddGroup(db *sql.DB, s *Group) error {
	err := verifyGroup(db, s)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetShow(db, s.Show)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO groups (%s) VALUES (%s)", groupDBKey, groupDBIdx), dbVals(s)...),
	}
	return dbExec(db, stmts)
}

// GetGroup은 db에서 하나의 샷을 찾는다.
// 해당 샷이 존재하지 않는다면 nil과 NotFound 에러를 반환한다.
func GetGroup(db *sql.DB, show, grp string) (*Group, error) {
	err := verifyGroupPrimaryKeys(show, grp)
	if err != nil {
		return nil, err
	}
	_, err = GetShow(db, show)
	if err != nil {
		return nil, err
	}
	s := &Group{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM groups WHERE show=$1 AND grp=$2 LIMIT 1", groupDBKey), show, grp)
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("group", JoinGroupID(show, grp))
		}
		return nil, err
	}
	return s, err
}

func ShowGroups(db *sql.DB, show string) ([]*Group, error) {
	err := verifyShowPrimaryKeys(show)
	if err != nil {
		return nil, err
	}
	grps := make([]*Group, 0)
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM groups WHERE show=$1", groupDBKey), show)
	err = dbQuery(db, stmt, func(row *sql.Rows) error {
		g := &Group{}
		err := scan(row, g)
		if err != nil {
			return err
		}
		grps = append(grps, g)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return grps, nil
}

// UpdateGroup은 db에서 해당 샷을 수정한다.
func UpdateGroup(db *sql.DB, s *Group) error {
	err := verifyGroup(db, s)
	if err != nil {
		return err
	}
	_, err = GetGroup(db, s.Show, s.Group)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE groups SET (%s) = (%s) WHERE show='%s' AND grp='%s'", groupDBKey, groupDBIdx, s.Show, s.Group), dbVals(s)...),
	}
	return dbExec(db, stmts)
}

// DeleteGroup은 해당 그룹과 그 하위의 모든 데이터를 db에서 지운다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteGroup(db *sql.DB, show, grp string) error {
	_, err := GetGroup(db, show, grp)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM groups WHERE show=$1 AND grp=$2", show, grp),
		dbStmt("DELETE FROM units WHERE show=$1 AND grp=$2", show, grp),
		dbStmt("DELETE FROM tasks WHERE show=$1 AND grp=$2", show, grp),
		dbStmt("DELETE FROM versions WHERE show=$1 AND grp=$2", show, grp),
	}
	return dbExec(db, stmts)
}
