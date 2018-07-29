package roi

import "time"

type Project struct {
	Code string
	Name string

	Status string

	Client        string
	Director      string
	Producer      string
	VFXSupervisor string
	VFXManager    string

	CrankIn     time.Time
	CrankUp     time.Time
	StartDate   time.Time
	ReleaseDate time.Time
	VFXDueDate  time.Time

	OutputSize string
	LutFile    string
}

var ProjectTableFields = []string{
	"code STRING NOT NULL UNIQUE CHECK (LENGTH(code) > 0) CHECK (code NOT LIKE '% %')",
	"name STRING",
	"status STRING",
	"client STRING",
	"director STRING",
	"producer STRING",
	"vfx_supervisor STRING",
	"vfx_manager STRING",
	"crank_in DATE",
	"crank_up DATE",
	"start_date DATE",
	"release_date DATE",
	"vfx_due_date DATE",
	"output_size STRING",
	"lut_file STRING",
}

func (p Project) dbKeyValues() []KV {
	return []KV{
		{"code", q(p.Code)},
		{"name", q(p.Name)},
		{"status", q(p.Status)},
		{"client", q(p.Client)},
		{"director", q(p.Director)},
		{"producer", q(p.Producer)},
		{"vfx_supervisor", q(p.VFXSupervisor)},
		{"vfx_manager", q(p.VFXManager)},
		{"crank_in", dbDate(p.CrankIn)},
		{"crank_up", dbDate(p.CrankUp)},
		{"start_date", dbDate(p.StartDate)},
		{"release_date", dbDate(p.ReleaseDate)},
		{"vfx_due_date", dbDate(p.VFXDueDate)},
		{"output_size", q(p.OutputSize)},
		{"lut_file", q(p.LutFile)},
	}
}
