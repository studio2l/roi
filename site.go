package roi

import "github.com/lib/pq"

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
	vfx_supervisors STRING[] NOT NULL,
	vfx_producers STRING[] NOT NULL,
	cg_supervisors STRING[] NOT NULL,
	project_managers STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	leads STRING[] NOT NULL
)`
