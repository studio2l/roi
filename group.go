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
	category STRING NOT NULL CHECK (length(category) > 0),
	grp STRING NOT NULL CHECK (length(grp) > 0) CHECK (grp NOT LIKE '% %'),
	notes STRING NOT NULL,
	attrs STRING NOT NULL,
	UNIQUE(show, grp),
	CONSTRAINT groups_pk PRIMARY KEY (show, grp)
)`

type Group struct {
	Show     string `db:"show"`
	Category string `db:"category"`
	Group    string `db:"grp"` // group이 sql 구문이기 때문에 줄여서 씀.

	Notes string `db:"notes"`

	// Attrs는 커스텀 속성으로 db에는 여러줄의 문자열로 저장된다. 각 줄은 키: 값의 쌍이다.
	Attrs DBStringMap `db:"attrs"`
}

var groupDBKey string = strings.Join(dbKeys(&Group{}), ", ")
var groupDBIdx string = strings.Join(dbIdxs(&Group{}), ", ")
var _ []interface{} = dbVals(&Group{})

// ID는 Group의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Group) ID() string {
	return s.Show + "/" + s.Category + "/" + s.Group
}

// SplitGroupID는 받아들인 샷 아이디를 쇼, 샷으로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitGroupID(id string) (string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", "", BadRequest(fmt.Sprintf("invalid group id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	group := ns[2]
	if show == "" || ctg == "" || group == "" {
		return "", "", "", BadRequest(fmt.Sprintf("invalid group id: %s", id))
	}
	return show, ctg, group, nil
}

// verifyGroupID는 받아들인 샷 아이디가 유효하지 않다면 에러를 반환한다.
func verifyGroupID(id string) error {
	show, ctg, group, err := SplitGroupID(id)
	if err != nil {
		return err
	}
	err = verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyCategoryName(ctg)
	if err != nil {
		return err
	}
	err = verifyGroupName(group)
	if err != nil {
		return err
	}
	return err
}

// 샷 이름은 일반적으로 (시퀀스를 나타내는) 접두어, 샷 번호, 접미어로 나뉜다.
// 접두어와 샷 번호는 항상 필요하지만, 접미어는 없어도 된다.
//
// CG0010a에서 접두어는 CG, 샷 번호는 0010, 접미어는 a이다.
//
// 접두어와 샷번호, 접미어를 언더바(_)를 통해서 떨어뜨리거나, 그냥 붙여쓰는것이 가능하다.
//
// 아래는 모두 유효한 샷 이름이다.
//
// CG0010
// CG0010a
// CG0010_a
// CG_0010a
// CG_0010_a
//
// 마지막 샷 이름의 경우 기본 이름은 CG_0010, 접미어는 CG가 된다.

var (
	// reGroupName은 유효한 샷 이름을 나타내는 정규식이다.
	reGroupName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
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
	return verifyGroupID(s.ID())
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
func GetGroup(db *sql.DB, id string) (*Group, error) {
	show, ctg, group, err := SplitGroupID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShow(db, show)
	if err != nil {
		return nil, err
	}
	s := &Group{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM groups WHERE show=$1 AND category=$2 AND grp=$3 LIMIT 1", groupDBKey), show, ctg, group)
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("group", id)
		}
		return nil, err
	}
	return s, err
}

func Groups(db *sql.DB, show, ctg string) ([]*Group, error) {
	grps := make([]*Group, 0)
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM groups WHERE show=$1 AND category=$2", groupDBKey), show, ctg)
	err := dbQuery(db, stmt, func(row *sql.Rows) error {
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
// 이 함수를 호출하기 전 해당 샷이 존재하는지 사용자가 검사해야 한다.
func UpdateGroup(db *sql.DB, id string, s *Group) error {
	err := verifyGroupID(id)
	if err != nil {
		return err
	}
	err = verifyGroup(db, s)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE groups SET (%s) = (%s) WHERE show='%s' AND category='%s' AND grp='%s'", groupDBKey, groupDBIdx, s.Show, s.Category, s.Group), dbVals(s)...),
	}
	return dbExec(db, stmts)
}

// DeleteGroup은 해당 샷과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 샷이 없어도 에러를 내지 않기 때문에 검사를 원한다면 GroupExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteGroup(db *sql.DB, id string) error {
	show, ctg, group, err := SplitGroupID(id)
	if err != nil {
		return err
	}
	_, err = GetGroup(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM groups WHERE show=$1 AND category=$2 AND grp=$3", show, ctg, group),
		dbStmt("DELETE FROM units WHERE show=$1 AND category=$2 AND grp=$3", show, ctg, group),
		dbStmt("DELETE FROM tasks WHERE show=$1 AND category=$2 AND grp=$3", show, ctg, group),
		dbStmt("DELETE FROM versions WHERE show=$1 AND category=$2 AND grp=$3", show, ctg, group),
	}
	return dbExec(db, stmts)
}
