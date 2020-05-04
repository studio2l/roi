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

// CreateTableIfNotExistUnitsStmt는 DB에 units 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsUnitsStmt = `CREATE TABLE IF NOT EXISTS units (
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	category STRING NOT NULL CHECK (length(category) > 0),
	unit STRING NOT NULL CHECK (length(unit) > 0) CHECK (unit NOT LIKE '% %'),
	status STRING NOT NULL CHECK (length(status) > 0),
	edit_order INT NOT NULL,
	description STRING NOT NULL,
	cg_description STRING NOT NULL,
	tags STRING[] NOT NULL,
	assets STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	start_date TIMESTAMPTZ NOT NULL,
	end_date TIMESTAMPTZ NOT NULL,
	due_date TIMESTAMPTZ NOT NULL,
	attrs STRING NOT NULL,
	UNIQUE(show, unit),
	CONSTRAINT units_pk PRIMARY KEY (show, unit)
)`

type Unit struct {
	Show     string `db:"show"`
	Category string `db:"category"`
	Unit     string `db:"unit"`

	// 샷 정보
	Status        Status   `db:"status"`
	EditOrder     int      `db:"edit_order"`
	Description   string   `db:"description"`
	CGDescription string   `db:"cg_description"`
	Tags          []string `db:"tags"`

	// Assets는 샷이 필요로 하는 애셋 이름 리스트이다.
	// 현재는 애셋이 같은 쇼 안에 존재할 때만 처리가 가능하다.
	// 여기에 등록된 애셋은 존재해야만 하며,
	// 애셋이 삭제되기 전 우선 모든 샷의 애셋 태그에서 지워져야 한다.
	Assets []string `db:"assets"`

	// Tasks는 샷에 작업중인 어떤 태스크가 있는지를 나타낸다.
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

	// Attrs는 커스텀 속성으로 db에는 여러줄의 문자열로 저장된다. 각 줄은 키: 값의 쌍이다.
	Attrs DBStringMap `db:"attrs"`
}

var unitDBKey string = strings.Join(dbKeys(&Unit{}), ", ")
var unitDBIdx string = strings.Join(dbIdxs(&Unit{}), ", ")
var _ []interface{} = dbVals(&Unit{})

// ID는 Unit의 고유 아이디이다. 다른 어떤 항목도 같은 아이디를 가지지 않는다.
func (s *Unit) ID() string {
	return s.Show + "/" + s.Category + "/" + s.Unit
}

// SplitUnitID는 받아들인 샷 아이디를 쇼, 샷으로 분리해서 반환한다.
// 만일 샷 아이디가 유효하지 않다면 에러를 반환한다.
func SplitUnitID(id string) (string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", "", BadRequest(fmt.Sprintf("invalid unit id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	unit := ns[2]
	if show == "" || ctg == "" || unit == "" {
		return "", "", "", BadRequest(fmt.Sprintf("invalid unit id: %s", id))
	}
	return show, ctg, unit, nil
}

// verifyUnitID는 받아들인 샷 아이디가 유효하지 않다면 에러를 반환한다.
func verifyUnitID(id string) error {
	show, ctg, unit, err := SplitUnitID(id)
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
	err = verifyUnitName(unit)
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
	// reUnitName은 유효한 샷 이름을 나타내는 정규식이다.
	reUnitName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// verifyUnitame은 받아들인 샷 이름이 유효하지 않다면 에러를 반환한다.
func verifyUnitName(unit string) error {
	if !reUnitName.MatchString(unit) {
		return BadRequest(fmt.Sprintf("invalid unit name: %s", unit))
	}
	return nil
}

// verifyUnit은 받아들인 샷이 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyUnit(db *sql.DB, s *Unit) error {
	if s == nil {
		return fmt.Errorf("nil unit")
	}
	err := verifyUnitID(s.ID())
	if err != nil {
		return err
	}
	err = verifyUnitStatus(s.Status)
	if err != nil {
		return err
	}
	// 태스크에는 순서가 있으므로 사이트에 정의된 순서대로 재정렬한다.
	si, err := GetSite(db)
	if err != nil {
		return err
	}
	var tasks []string
	hasTask := make(map[string]bool)
	taskIdx := make(map[string]int)
	switch s.Category {
	case "shot":
		tasks = si.ShotTasks
	case "asset":
		tasks = si.AssetTasks
	default:
		return fmt.Errorf("invalid unit category: %s", s.Category)
	}
	for i, task := range tasks {
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
	sort.Slice(s.Assets, func(i, j int) bool {
		return strings.Compare(s.Assets[i], s.Assets[j]) <= 0
	})
	for _, asset := range s.Assets {
		_, err := GetUnit(db, s.Show+"/asset/"+asset)
		if err != nil {
			return err
		}
	}
	sort.Slice(s.Tags, func(i, j int) bool {
		return strings.Compare(s.Tags[i], s.Tags[j]) <= 0
	})
	if len(s.Tags) != 0 {
		// 샷의 태그가 쇼에 이미 생성되어 있는 태그가 아니라면 쇼에 추가한다.
		//
		// 할일: 이 함수에서 쇼를 고치는 것이 맞는 방식이란 생각이 들지는 않는다.
		// 나중에 각각의 속성을 독립적으로 수정할 수 있게 하게 되면 여기서 이동 시킨다.
		sh, err := GetShow(db, s.Show)
		if err != nil {
			return err
		}
		showTag := make(map[string]bool)
		for _, t := range sh.Tags {
			showTag[t] = true
		}
		updateShowTag := false
		for _, t := range s.Tags {
			if !showTag[t] {
				showTag[t] = true
				updateShowTag = true
			}
		}
		if updateShowTag {
			tags := make([]string, 0, len(showTag))
			for t := range showTag {
				tags = append(tags, t)
			}
			sh.Tags = tags
			err := UpdateShow(db, sh.ID(), sh)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddUnit은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddUnit(db *sql.DB, s *Unit) error {
	err := verifyUnit(db, s)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetShow(db, s.Show)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO units (%s) VALUES (%s)", unitDBKey, unitDBIdx), dbVals(s)...),
	}
	// 하위 태스크 생성
	for _, task := range s.Tasks {
		t := &Task{
			Show:     s.Show,
			Category: s.Category,
			Unit:     s.Unit,
			Task:     task,
			Status:   StatusInProgress,
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
	return dbExec(db, stmts)
}

// GetUnit은 db에서 하나의 샷을 찾는다.
// 해당 샷이 존재하지 않는다면 nil과 NotFound 에러를 반환한다.
func GetUnit(db *sql.DB, id string) (*Unit, error) {
	show, ctg, unit, err := SplitUnitID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetShow(db, show)
	if err != nil {
		return nil, err
	}
	s := &Unit{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM units WHERE show=$1 AND category=$2 AND unit=$3 LIMIT 1", unitDBKey), show, ctg, unit)
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("unit", id)
		}
		return nil, err
	}
	return s, err
}

// SearchUnits는 db의 특정 프로젝트에서 검색 조건에 맞는 샷 리스트를 반환한다.
func SearchUnits(db *sql.DB, show, ctg string, units []string, tag, status, task, assignee, task_status string, task_due_date time.Time) ([]*Unit, error) {
	keys := ""
	for i, k := range dbKeys(&Unit{}) {
		if i != 0 {
			keys += ", "
		}
		// 태스크에 있는 정보를 찾기 위해 JOIN 해야 할 경우가 있기 때문에
		// units. 을 붙인다.
		keys += "units." + k
	}
	where := make([]string, 0)
	vals := make([]interface{}, 0)
	i := 1 // 인덱스가 1부터 시작이다.
	stmt := fmt.Sprintf("SELECT %s FROM units", keys)
	where = append(where, fmt.Sprintf("units.show=$%d", i))
	vals = append(vals, show)
	i++
	if ctg != "" {
		where = append(where, fmt.Sprintf("units.category=$%d", i))
		vals = append(vals, ctg)
		i++
	}
	if len(units) != 0 {
		j := 0
		whereUnits := "("
		for _, unit := range units {
			if unit == "" {
				continue
			}
			if j != 0 {
				whereUnits += " OR "
			}
			if strings.Contains(unit, "*") {
				unit = strings.Replace(unit, "*", "%", -1)
				whereUnits += fmt.Sprintf("units.unit LIKE $%d", i)
			} else {
				whereUnits += fmt.Sprintf("units.unit=$%d", i)
			}
			vals = append(vals, unit)
			i++
			j++
		}
		whereUnits += ")"
		where = append(where, whereUnits)
	}
	if tag != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(units.tags)", i))
		vals = append(vals, tag)
		i++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("units.status=$%d", i))
		vals = append(vals, status)
		i++
	}
	if task != "" {
		where = append(where, fmt.Sprintf("$%d::string = ANY(units.tasks)", i))
		vals = append(vals, task)
		i++
	}
	if assignee != "" || task_status != "" || !task_due_date.IsZero() {
		stmt += " JOIN tasks ON (tasks.show = units.show AND tasks.unit = units.unit)"
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
	ss := make([]*Unit, 0)
	hasUnit := make(map[string]bool)
	err := dbQuery(db, st, func(rows *sql.Rows) error {
		// 태스크 검색을 해 JOIN이 되면 샷이 중복으로 추가될 수 있다.
		// DISTINCT를 이용해 문제를 해결하려고 했으나 DB가 꺼진다.
		// 우선은 여기서 걸러낸다.
		s := &Unit{}
		err := scan(rows, s)
		if err != nil {
			return err
		}
		if !hasUnit[s.Unit] {
			hasUnit[s.Unit] = true
			ss = append(ss, s)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("search units: %w", err)
	}
	sort.Slice(ss, func(i int, j int) bool {
		return ss[i].Unit <= ss[j].Unit
	})
	return ss, nil
}

// UpdateUnit은 db에서 해당 샷을 수정한다.
// 이 함수를 호출하기 전 해당 샷이 존재하는지 사용자가 검사해야 한다.
func UpdateUnit(db *sql.DB, id string, s *Unit) error {
	err := verifyUnitID(id)
	if err != nil {
		return err
	}
	err = verifyUnit(db, s)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE units SET (%s) = (%s) WHERE show='%s' AND unit='%s'", unitDBKey, unitDBIdx, s.Show, s.Unit), dbVals(s)...),
	}
	// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
	for _, task := range s.Tasks {
		_, err := GetTask(db, id+"/"+task)
		if err != nil {
			if !errors.As(err, &NotFoundError{}) {
				return fmt.Errorf("get task: %s", err)
			} else {
				t := &Task{
					Show:     s.Show,
					Category: "unit",
					Unit:     s.Unit,
					Task:     task,
					Status:   StatusInProgress,
					DueDate:  time.Time{},
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

// DeleteUnit은 해당 샷과 그 하위의 모든 데이터를 db에서 지운다.
// 해당 샷이 없어도 에러를 내지 않기 때문에 검사를 원한다면 UnitExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteUnit(db *sql.DB, id string) error {
	show, ctg, unit, err := SplitUnitID(id)
	if err != nil {
		return err
	}
	_, err = GetUnit(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt("DELETE FROM units WHERE show=$1 AND category=$2 AND unit=$3", show, ctg, unit),
		dbStmt("DELETE FROM tasks WHERE show=$1 AND category=$2 AND unit=$3", show, ctg, unit),
		dbStmt("DELETE FROM versions WHERE show=$1 AND category=$2 AND unit=$3", show, ctg, unit),
	}
	return dbExec(db, stmts)
}
