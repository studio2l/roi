package roi

import (
	"regexp"
	"time"
)

var reValidProjectName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsValidProjectName(name string) bool {
	return reValidProjectName.MatchString(name)
}

type Project struct {
	Code string
	Name string

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
	LutFile    string
}

var ProjectTableFields = []string{
	// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	"id UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"code STRING NOT NULL UNIQUE CHECK (LENGTH(code) > 0) CHECK (code NOT LIKE '% %')",
	"name STRING",
	"status STRING",
	"client STRING",
	"director STRING",
	"producer STRING",
	"vfx_supervisor STRING",
	"vfx_manager STRING",
	"cg_supervisor STRING",
	"crank_in DATE",
	"crank_up DATE",
	"start_date DATE",
	"release_date DATE",
	"vfx_due_date DATE",
	"output_size STRING",
	"lut_file STRING",
}

func (p Project) toOrdMap() *ordMap {
	o := newOrdMap()
	o.Set("code", p.Code)
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
	o.Set("lut_file", p.LutFile)
	return o
}
