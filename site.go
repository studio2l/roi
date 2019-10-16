package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Site는 현재 스튜디오를 뜻한다.
//
// 할일: 만일 로이가 여러 사이트를 관리하게 하려면 이름을 추가한다.
type Site struct {
	Site            string
	VFXSupervisors  []string
	VFXProducers    []string
	CGSupervisors   []string
	ProjectManagers []string
	Tasks           []string
	// Leads는 task:name 형식이고 한 파트에 여러명이 등록될 수 있다.
	// 이 때 [... rnd:kybin rnd:kaycho ...] 처럼 등록한다.
	// 형식이 맞지 않거나 Tasks에 없는 태스크명을 쓰면 에러를 낸다.
	Leads []string
}

func (s *Site) dbValues() []interface{} {
	if s == nil {
		s = &Site{}
	}
	s.Site = "only" // 현재로서는 강제한다.
	if s.VFXSupervisors == nil {
		s.VFXSupervisors = []string{}
	}
	if s.VFXProducers == nil {
		s.VFXProducers = []string{}
	}
	if s.CGSupervisors == nil {
		s.CGSupervisors = []string{}
	}
	if s.ProjectManagers == nil {
		s.ProjectManagers = []string{}
	}
	if s.Tasks == nil {
		s.Tasks = []string{}
	}
	if s.Leads == nil {
		s.Leads = []string{}
	}
	vals := []interface{}{
		s.Site,
		pq.Array(s.VFXSupervisors),
		pq.Array(s.VFXProducers),
		pq.Array(s.CGSupervisors),
		pq.Array(s.ProjectManagers),
		pq.Array(s.Tasks),
		pq.Array(s.Leads),
	}
	return vals
}

var SitesTableKeys = []string{
	"site",
	"vfx_supervisors",
	"vfx_producers",
	"cg_supervisors",
	"project_managers",
	"tasks",
	"leads",
}

var SitesTableIndices = dbIndices(SitesTableKeys)

var CreateTableIfNotExistsSitesStmt = `CREATE TABLE IF NOT EXISTS sites (
	site STRING UNIQUE NOT NULL,
	vfx_supervisors STRING[] NOT NULL,
	vfx_producers STRING[] NOT NULL,
	cg_supervisors STRING[] NOT NULL,
	project_managers STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	leads STRING[] NOT NULL
)`

// AddSite는 DB에 빈 사이트를 생성한다.
func AddSite(db *sql.DB) error {
	s := &Site{}
	s.Site = "only"
	keys := strings.Join(SitesTableKeys, ", ")
	idxs := strings.Join(SitesTableIndices, ", ")
	stmt := fmt.Sprintf("INSERT INTO sites (%s) VALUES (%s) ON CONFLICT DO NOTHING", keys, idxs)
	if _, err := db.Exec(stmt, s.dbValues()...); err != nil {
		return err
	}
	return nil
}

// UpdateSite는 DB의 사이트 정보를 업데이트한다.
func UpdateSite(db *sql.DB, s *Site) error {
	s.Site = "only"
	keys := strings.Join(SitesTableKeys, ", ")
	idxs := strings.Join(SitesTableIndices, ", ")
	stmt := fmt.Sprintf("UPDATE sites SET (%s) = (%s) WHERE site='only'", keys, idxs)
	if _, err := db.Exec(stmt, s.dbValues()...); err != nil {
		return err
	}
	return nil
}

// siteFromRows는 테이블의 한 열에서 사이트를 받아온다.
func siteFromRows(rows *sql.Rows) (*Site, error) {
	s := &Site{}
	err := rows.Scan(
		&s.Site,
		pq.Array(&s.VFXSupervisors),
		pq.Array(&s.VFXProducers),
		pq.Array(&s.CGSupervisors),
		pq.Array(&s.ProjectManagers),
		pq.Array(&s.Tasks),
		pq.Array(&s.Leads),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetSite는 db에서 사이트 정보를 가지고 온다.
// 사이트 정보는 항상 존재해야 한다.
func GetSite(db *sql.DB) (*Site, error) {
	keystr := strings.Join(SitesTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM sites LIMIT 1", keystr)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, errors.New("site should exist")
	}
	return siteFromRows(rows)
}