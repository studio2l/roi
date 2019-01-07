package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var reValidProjectID = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsValidProjectID(id string) bool {
	return reValidProjectID.MatchString(id)
}

type Project struct {
	// 프로젝트 아이디. 로이 내에서 고유해야 한다.
	ID string

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

	OutputSize string
	ViewLUT    string
}

func (p *Project) dbValues() []interface{} {
	if p == nil {
		p = &Project{}
	}
	vals := []interface{}{
		p.ID,
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
	}
	return vals
}

var ProjectTableKeys = []string{
	"id",
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
}

var ProjectTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9", "$10",
	"$11", "$12", "$13", "$14", "$15", "$16",
}

var ProjectTableFields = []string{
	"uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"id STRING NOT NULL UNIQUE CHECK (LENGTH(id) > 0) CHECK (id NOT LIKE '% %')",
	"name STRING NOT NULL",
	"status STRING NOT NULL",
	"client STRING NOT NULL",
	"director STRING NOT NULL",
	"producer STRING NOT NULL",
	"vfx_supervisor STRING NOT NULL",
	"vfx_manager STRING NOT NULL",
	"cg_supervisor STRING NOT NULL",
	"crank_in TIMESTAMPTZ NOT NULL",
	"crank_up TIMESTAMPTZ NOT NULL",
	"start_date TIMESTAMPTZ NOT NULL",
	"release_date TIMESTAMPTZ NOT NULL",
	"vfx_due_date TIMESTAMPTZ NOT NULL",
	"output_size STRING NOT NULL",
	"view_lut STRING NOT NULL",
}

// AddProject는 db에 프로젝트를 추가한다.
func AddProject(db *sql.DB, p *Project) error {
	if p == nil {
		return errors.New("nil Project is invalid")
	}
	if !IsValidProjectID(p.ID) {
		return fmt.Errorf("Project id is invalid: %s", p.ID)
	}
	keystr := strings.Join(ProjectTableKeys, ", ")
	idxstr := strings.Join(ProjectTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO projects (%s) VALUES (%s)", keystr, idxstr)
	if _, err := db.Exec(stmt, p.dbValues()...); err != nil {
		return err
	}
	// TODO: add project info, task, tracking table
	if err := CreateTableIfNotExists(db, p.ID+"_shots", ShotTableFields); err != nil {
		return err
	}
	return nil
}

// ProjectExist는 db에 해당 프로젝트가 존재하는지를 검사한다.
func ProjectExist(db *sql.DB, prj string) (bool, error) {
	rows, err := db.Query("SELECT id FROM projects WHERE id=$1 LIMIT 1", prj)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

// projectFromRows는 테이블의 한 열에서 프로젝트를 받아온다.
func projectFromRows(rows *sql.Rows) (*Project, error) {
	p := &Project{}
	err := rows.Scan(
		&p.ID, &p.Name, &p.Status, &p.Client,
		&p.Director, &p.Producer, &p.VFXSupervisor, &p.VFXManager, &p.CGSupervisor,
		&p.CrankIn, &p.CrankUp, &p.StartDate, &p.ReleaseDate, &p.VFXDueDate, &p.OutputSize,
		&p.ViewLUT,
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
	stmt := fmt.Sprintf("SELECT %s FROM projects WHERE id=$1", keystr)
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

// SearchAllProjects는 db에서 모든 프로젝트 정보를 가져온다.
// 검색 중 문제가 있으면 nil, error를 반환한다.
func SearchAllProjects(db *sql.DB) ([]*Project, error) {
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
