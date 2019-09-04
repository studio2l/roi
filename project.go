package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
)

var reValidProject = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsValidProject(id string) bool {
	return reValidProject.MatchString(id)
}

type Project struct {
	// 프로젝트 아이디. 로이 내에서 고유해야 한다.
	Project string

	Name   string
	Status string

	Client        string
	Director      string
	Producer      string
	VFXSupervisor string
	VFXManager    string
	CGSupervisor  string

	CrankIn     time.Time
	CrankUp     time.Time
	StartDate   time.Time
	ReleaseDate time.Time
	VFXDueDate  time.Time

	OutputSize   string
	ViewLUT      string
	DefaultTasks []string
}

func (p *Project) dbValues() []interface{} {
	if p == nil {
		p = &Project{}
	}
	if p.DefaultTasks == nil {
		p.DefaultTasks = []string{}
	}
	vals := []interface{}{
		p.Project,
		p.Name,
		p.Status,
		p.Client,
		p.Director,
		p.Producer,
		p.VFXSupervisor,
		p.VFXManager,
		p.CGSupervisor,
		p.CrankIn,
		p.CrankUp,
		p.StartDate,
		p.ReleaseDate,
		p.VFXDueDate,
		p.OutputSize,
		p.ViewLUT,
		pq.Array(p.DefaultTasks),
	}
	return vals
}

var ProjectTableKeys = []string{
	"project",
	"name",
	"status",
	"client",
	"director",
	"producer",
	"vfx_supervisor",
	"vfx_manager",
	"cg_supervisor",
	"crank_in",
	"crank_up",
	"start_date",
	"release_date",
	"vfx_due_date",
	"output_size",
	"view_lut",
	"default_tasks",
}

var ProjectTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9", "$10",
	"$11", "$12", "$13", "$14", "$15", "$16", "$17",
}

var CreateTableIfNotExistsProjectsStmt = `CREATE TABLE IF NOT EXISTS projects (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project STRING NOT NULL UNIQUE CHECK (LENGTH(project) > 0) CHECK (project NOT LIKE '% %'),
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

// AddProject는 db에 프로젝트를 추가한다.
func AddProject(db *sql.DB, p *Project) error {
	if p == nil {
		return errors.New("nil Project is invalid")
	}
	if !IsValidProject(p.Project) {
		return fmt.Errorf("Project id is invalid: %s", p.Project)
	}
	keystr := strings.Join(ProjectTableKeys, ", ")
	idxstr := strings.Join(ProjectTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO projects (%s) VALUES (%s)", keystr, idxstr)
	if _, err := db.Exec(stmt, p.dbValues()...); err != nil {
		return err
	}
	return nil
}

// UpdateProjectParam은 Project에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateProject에서 사용한다.
type UpdateProjectParam struct {
	Name          string
	Status        string
	Client        string
	Director      string
	Producer      string
	VFXSupervisor string
	VFXManager    string
	CGSupervisor  string
	CrankIn       time.Time
	CrankUp       time.Time
	StartDate     time.Time
	ReleaseDate   time.Time
	VFXDueDate    time.Time
	OutputSize    string
	ViewLUT       string
	DefaultTasks  []string
}

func (u UpdateProjectParam) keys() []string {
	return []string{
		"name",
		"status",
		"client",
		"director",
		"producer",
		"vfx_supervisor",
		"vfx_manager",
		"cg_supervisor",
		"crank_in",
		"crank_up",
		"start_date",
		"release_date",
		"vfx_due_date",
		"output_size",
		"view_lut",
		"default_tasks",
	}
}

func (u UpdateProjectParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateProjectParam) values() []interface{} {
	if u.DefaultTasks == nil {
		u.DefaultTasks = []string{}
	}
	return []interface{}{
		u.Name,
		u.Status,
		u.Client,
		u.Director,
		u.Producer,
		u.VFXSupervisor,
		u.VFXManager,
		u.CGSupervisor,
		u.CrankIn,
		u.CrankUp,
		u.StartDate,
		u.ReleaseDate,
		u.VFXDueDate,
		u.OutputSize,
		u.ViewLUT,
		pq.Array(u.DefaultTasks),
	}
}

// UpdateProject는 db의 프로젝트 정보를 수정한다.
func UpdateProject(db *sql.DB, prj string, upd UpdateProjectParam) error {
	if !IsValidProject(prj) {
		return fmt.Errorf("Project id is invalid: %s", prj)
	}
	keystr := strings.Join(upd.keys(), ", ")
	idxstr := strings.Join(upd.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE projects SET (%s) = (%s) WHERE project='%s'", keystr, idxstr, prj)
	if _, err := db.Exec(stmt, upd.values()...); err != nil {
		return err
	}
	return nil
}

// ProjectExist는 db에 해당 프로젝트가 존재하는지를 검사한다.
func ProjectExist(db *sql.DB, prj string) (bool, error) {
	rows, err := db.Query("SELECT project FROM projects WHERE project=$1 LIMIT 1", prj)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// projectFromRows는 테이블의 한 열에서 프로젝트를 받아온다.
func projectFromRows(rows *sql.Rows) (*Project, error) {
	p := &Project{}
	err := rows.Scan(
		&p.Project, &p.Name, &p.Status, &p.Client,
		&p.Director, &p.Producer, &p.VFXSupervisor, &p.VFXManager, &p.CGSupervisor,
		&p.CrankIn, &p.CrankUp, &p.StartDate, &p.ReleaseDate, &p.VFXDueDate, &p.OutputSize,
		&p.ViewLUT, pq.Array(&p.DefaultTasks),
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetProject는 db에서 특정 프로젝트 정보를 부른다.
// 해당 프로젝트가 없다면 nil이 반환된다.
func GetProject(db *sql.DB, prj string) (*Project, error) {
	keystr := strings.Join(ProjectTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM projects WHERE project=$1", keystr)
	rows, err := db.Query(stmt, prj)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	p, err := projectFromRows(rows)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// AllProjects는 db에서 모든 프로젝트 정보를 가져온다.
// 검색 중 문제가 있으면 nil, error를 반환한다.
func AllProjects(db *sql.DB) ([]*Project, error) {
	fields := strings.Join(ProjectTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM projects", fields)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	prjs := make([]*Project, 0)
	for rows.Next() {
		p, err := projectFromRows(rows)
		if err != nil {
			return nil, err
		}
		prjs = append(prjs, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return prjs, nil
}

// DeleteProject는 해당 프로젝트와 그 하위의 모든 데이터를 db에서 지운다.
// 해당 프로젝트가 없어도 에러를 내지 않기 때문에 검사를 원한다면 ProjectExist를 사용해야 한다.
// 만일 처리 중간에 에러가 나면 아무 데이터도 지우지 않고 에러를 반환한다.
func DeleteProject(db *sql.DB, prj string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec("DELETE FROM projects WHERE project=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'projects' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM shots WHERE project=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'shots' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM tasks WHERE project=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'tasks' table: %v", err)
	}
	if _, err := tx.Exec("DELETE FROM versions WHERE project=$1", prj); err != nil {
		return fmt.Errorf("could not delete data from 'versions' table: %v", err)
	}
	return tx.Commit()
}
