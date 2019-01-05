package roi

import (
	"regexp"
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
	"crank_in DATE NOT NULL",
	"crank_up DATE NOT NULL",
	"start_date DATE NOT NULL",
	"release_date DATE NOT NULL",
	"vfx_due_date DATE NOT NULL",
	"output_size STRING NOT NULL",
	"view_lut STRING NOT NULL",
}

// ordMapFromProject는 프로젝트 정보를 OrdMap에 담는다.
func ordMapFromProject(p *Project) *ordMap {
	o := newOrdMap()
	o.Set("id", p.ID)
	o.Set("name", p.Name)
	o.Set("status", p.Status)
	o.Set("client", p.Client)
	o.Set("director", p.Director)
	o.Set("producer", p.Producer)
	o.Set("vfx_supervisor", p.VFXSupervisor)
	o.Set("vfx_manager", p.VFXManager)
	o.Set("cg_supervisor", p.CGSupervisor)
	o.Set("crank_in", p.CrankIn)
	o.Set("crank_up", p.CrankUp)
	o.Set("start_date", p.StartDate)
	o.Set("release_date", p.ReleaseDate)
	o.Set("vfx_due_date", p.VFXDueDate)
	o.Set("output_size", p.OutputSize)
	o.Set("view_lut", p.ViewLUT)
	return o
}
