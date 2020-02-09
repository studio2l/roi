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

// CreateTableIfNotExistAssetsStmt는 DB에 assets 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsAssetsStmt = `CREATE TABLE IF NOT EXISTS assets (
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	asset STRING NOT NULL CHECK (length(asset) > 0) CHECK (asset NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	description STRING NOT NULL,
	cg_description STRING NOT NULL,
	tags STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	UNIQUE(show, asset),
	CONSTRAINT assets_pk PRIMARY KEY (show, asset)
)`

type Asset struct {
	Show  string `db:"show"`
	Asset string `db:"asset"`

	// 애셋 정보
	Status        AssetStatus `db:"status"`
	Description   string      `db:"description"`
	CGDescription string      `db:"cg_description"`
	Tags          []string    `db:"tags"`

	// Tasks는 애셋에 작업중인 어떤 태스크가 있는지를 나타낸다.
	// 웹 페이지에는 여기에 포함된 태스크만 이 순서대로 보여져야 한다.
	//
	// 참고: 여기에 포함되어 있다면 db내에 해당 태스크가 존재해야 한다.
	// 반대로 여기에 포함되어 있지 않지만 db내에는 존재하는 태스크가 있을 수 있다.
	// 그 태스크는 (예를 들어 태스크가 Omit 되는 등의 이유로) 숨겨진 태스크이며,
	// 직접 지우지 않는 한 db에 보관된다.
	Tasks []string `db:"tasks"`

	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	DueDate   time.Time `db:"due_date"`
}

// ID는 Asset의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Asset) ID() string {
	return s.Show + "/asset/" + s.Asset
}

// SplitAssetID는 받아들인 애셋 아이디를 쇼, 애셋으로 분리해서 반환한다.
// 만일 애셋 아이디가 유효하지 않다면 에러를 반환한다.
func SplitAssetID(id string) (string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", BadRequest(fmt.Sprintf("invalid asset id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	asset := ns[2]
	if show == "" || ctg != "asset" || asset == "" {
		return "", "", BadRequest(fmt.Sprintf("invalid asset id: %s", id))
	}
	return show, asset, nil
}

// verifyAssetID는 받아들인 애셋 아이디가 유효하지 않다면 에러를 반환한다.
func verifyAssetID(id string) error {
	show, asset, err := SplitAssetID(id)
	if err != nil {
		return err
	}
	err = verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyAssetName(asset)
	if err != nil {
		return err
	}
	return err
}

// 애셋 이름은 영문으로 시작하고 영문 또는 숫자를 사용할수 있다.
var (
	// reAssetName은 유효한 애셋 이름을 나타내는 정규식이다.
	reAssetName = regexp.MustCompile(`^[a-zA-Z]+[_]?[a-zA-Z0-9]*[_]?[a-zA-Z0-9]*$`)
)

// verifyAssetName은 받아들인 애셋 이름이 유효하지 않다면 에러를 반환한다.
func verifyAssetName(asset string) error {
	if !reAssetName.MatchString(asset) {
		return BadRequest(fmt.Sprintf("invalid asset name: %s", asset))
	}
	return nil
}

// verifyAsset은 받아들인 애셋이 유효하지 않다면 에러를 반환한다.
// 필요하다면 db에 접근해서 정보를 검색한다.
func verifyAsset(db *sql.DB, s *Asset) error {
	if s == nil {
		return fmt.Errorf("nil asset")
	}
	err := verifyAssetID(s.ID())
	if err != nil {
		return err
	}
	err = verifyAssetStatus(s.Status)
	if err != nil {
		return err
	}
	// 태스크에는 순서가 있으므로 사이트에 정의된 순서대로 재정렬한다.
	si, err := GetSite(db)
	if err != nil {
		return err
	}
	hasTask := make(map[string]bool)
	taskIdx := make(map[string]int)
	for i, task := range si.AssetTasks {
		hasTask[task] = true
		taskIdx[task] = i
	}
	for _, task := range s.Tasks {
		if !hasTask[task] {
			return BadRequest(fmt.Sprintf("task %q not defined at site", task))
		}
	}
	sort.Slice(s.Tasks, func(i, j int) bool {
		return taskIdx[s.Tasks[i]] <= taskIdx[s.Tasks[j]]
	})
	return nil
}

// AddAsset은 db의 특정 프로젝트에 애셋을 하나 추가한다.
func AddAsset(db *sql.DB, s *Asset) error {
	err := verifyAsset(db, s)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetShow(db, s.Show)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO assets (%s) VALUES (%s)", keys, idxs), vs...),
	}
	// 하위 태스크 생성
	for _, task := range s.Tasks {
		t := &Task{
			Show:     s.Show,
			Category: "asset",
			Unit:     s.Asset,
			Task:     task,
			Status:   TaskInProgress,
			DueDate:  time.Time{},
		}
		st, err := addTaskStmts(t)
		if err != nil {
			return err
		}
		stmts = append(stmts, st...)
	}
	return dbExec(db, stmts)
}

// GetAsset은 db에서 하나의 애셋을 찾는다.
// 해당 애셋이 존재하지 않는다면 nil과 NotFound 에러를 반환한다.
func GetAsset(db *sql.DB, id string) (*Asset, error) {
	show, asset, err := SplitAssetID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShow(db, show)
	if err != nil {
		return nil, err
	}
	ks, _, _, err := dbKIVs(&Asset{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	s := &Asset{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM assets WHERE show=$1 AND asset=$2 LIMIT 1", keys), show, asset)
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("asset", id)
		}
		return nil, err
	}
	return s, err
}

// SearchAssets는 db의 특정 프로젝트에서 검색 조건에 맞는 애셋 리스트를 반환한다.
func SearchAssets(db *sql.DB, show string, assets []string, tag, status, task, assignee, task_status string, task_due_date time.Time) ([]*Asset, error) {
	ks, _, _, err := dbKIVs(&Asset{})
	if err != nil {
		return nil, err
	}
	keys := ""
	for i, k := range ks {
		if i != 0 {
			keys += ", "
		}
		// 태스크에 있는 정보를 찾기 위해 JOIN 해야 할 경우가 있기 때문에
		// assets. 을 붙인다.
		keys += "assets." + k
	}
	where := make([]string, 0)
	vals := make([]interface{}, 0)
	i := 1 // 인덱스가 1부터 시작이다.
	stmt := fmt.Sprintf("SELECT %s FROM assets", keys)
	where = append(where, fmt.Sprintf("assets.show=$%d", i))
	vals = append(vals, show)
	i++
	if len(assets) != 0 {
		j := 0
		whereAssets := "("
		for _, asset := range assets {
			if asset == "" {
				continue
			}
			if j != 0 {
				whereAssets += " OR "
			}
			whereAssets += fmt.Sprintf("assets.asset LIKE $%d", i)
			vals = append(vals, "%%"+asset+"%%")
			i++
			j++
		}
		whereAssets += ")"
		where = append(where, whereAssets)
	}
	if tag != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(assets.tags)", i))
		vals = append(vals, tag)
		i++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("assets.status=$%d", i))
		vals = append(vals, status)
		i++
	}
	if task != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(assets.tasks)", i))
		vals = append(vals, task)
		i++
	}
	if assignee != "" || task_status != "" || !task_due_date.IsZero() {
		stmt += " JOIN tasks ON (tasks.show = assets.show AND tasks.unit = assets.asset)"
	}
	if assignee != "" {
		where = append(where, fmt.Sprintf("tasks.assignee=$%d", i))
		vals = append(vals, assignee)
		i++
	}
	if task_status != "" {
		where = append(where, fmt.Sprintf("tasks.status=$%d", i))
		vals = append(vals, task_status)
		i++
	}
	if !task_due_date.IsZero() {
		where = append(where, fmt.Sprintf("tasks.due_date=$%d", i))
		vals = append(vals, task_due_date)
		i++
	}
	wherestr := strings.Join(where, " AND ")
	if wherestr != "" {
		stmt += " WHERE " + wherestr
	}
	st := dbStmt(stmt, vals...)
	ss := make([]*Asset, 0)
	hasAsset := make(map[string]bool)
	err = dbQuery(db, st, func(rows *sql.Rows) error {
		// 태스크 검색을 해 JOIN이 되면 애셋이 중복으로 추가될 수 있다.
		// DISTINCT를 이용해 문제를 해결하려고 했으나 DB가 꺼진다.
		// 우선은 여기서 걸러낸다.
		s := &Asset{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		if !hasAsset[s.Asset] {
			hasAsset[s.Asset] = true
			ss = append(ss, s)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(ss, func(i int, j int) bool {
		return ss[i].Asset <= ss[j].Asset
	})
	return ss, nil
}

// UpdateAsset은 db에서 해당 애셋을 수정한다.
// 이 함수를 호출하기 전 해당 애셋이 존재하는지 사용자가 검사해야 한다.
func UpdateAsset(db *sql.DB, id string, s *Asset) error {
	err := verifyAssetID(id)
	if err != nil {
		return err
	}
	err = verifyAsset(db, s)
	if err != nil {
		return err
	}
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE assets SET (%s) = (%s) WHERE show='%s' AND asset='%s'", keys, idxs, s.Show, s.Asset), vs...),
	}
	// 애셋에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
	for _, task := range s.Tasks {
		_, err := GetTask(db, id+"/"+task)
		if err != nil {
			if !errors.As(err, &NotFoundError{}) {
				return fmt.Errorf("get task: %s", err)
			} else {
				t := &Task{
					Show:     s.Show,
					Category: "asset",
					Unit:     s.Asset,
					Task:     task,
					Status:   TaskInProgress,
					DueDate:  time.Time{},
				}
				err := verifyTask(db, t)
				if err != nil {
					return err
				}
				st, err := addTaskStmts(t)
				if err != nil {
					return err
				}
				stmts = append(stmts, st...)
			}
		}
	}
	return dbExec(db, stmts)
}

// DeleteAsset은 해당 애셋과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 애셋이 없어도 에러를 내지 않기 때문에 검사를 원한다면 AssetExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteAsset(db *sql.DB, id string) error {
	show, asset, err := SplitAssetID(id)
	if err != nil {
		return err
	}
	_, err = GetAsset(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM assets WHERE show=$1 AND asset=$2", show, asset),
		dbStmt("DELETE FROM tasks WHERE show=$1 AND category=$2 AND unit=$3", show, "asset", asset),
		dbStmt("DELETE FROM versions WHERE show=$1 AND category=$2 AND unit=$3", show, "asset", asset),
	}
	return dbExec(db, stmts)
}
