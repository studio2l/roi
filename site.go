package roi

import (
	"database/sql"
	"fmt"
	"strings"
)

// Site는 현재 스튜디오를 뜻한다.
type Site struct {
	// 현재로서는 only라는 하나의 사이트만 존재
	Site            string   `db:"site"`
	VFXSupervisors  []string `db:"vfx_supervisors"`
	VFXProducers    []string `db:"vfx_producers"`
	CGSupervisors   []string `db:"cg_supervisors"`
	ProjectManagers []string `db:"project_managers"`
	Tasks           []string `db:"tasks"`
	DefaultTasks    []string `db:"default_tasks"` // 샷이 생성될 때 기본적으로 생겨야 하는 태스크
	// Leads는 task:name 형식이고 한 파트에 여러명이 등록될 수 있다.
	// 이 때 [... rnd:kybin rnd:kaycho ...] 처럼 등록한다.
	// 형식이 맞지 않거나 Tasks에 없는 태스크명을 쓰면 에러를 낸다.
	Leads []string `db:"leads"`
}

var CreateTableIfNotExistsSitesStmt = `CREATE TABLE IF NOT EXISTS sites (
	site STRING UNIQUE NOT NULL,
	vfx_supervisors STRING[] NOT NULL,
	vfx_producers STRING[] NOT NULL,
	cg_supervisors STRING[] NOT NULL,
	project_managers STRING[] NOT NULL,
	tasks STRING[] NOT NULL,
	default_tasks STRING[] NOT NULL,
	leads STRING[] NOT NULL
)`

// AddSite는 DB에 빈 사이트를 생성한다.
func AddSite(db *sql.DB) error {
	k, v, err := dbKVs(&Site{Site: "only"})
	if err != nil {
		return err
	}
	keys := strings.Join(k, ", ")
	idxs := strings.Join(dbIndices(k), ", ")
	stmt := fmt.Sprintf("INSERT INTO sites (%s) VALUES (%s) ON CONFLICT DO NOTHING", keys, idxs)
	if _, err := db.Exec(stmt, v...); err != nil {
		return err
	}
	return nil
}

// UpdateSite는 DB의 사이트 정보를 업데이트한다.
func UpdateSite(db *sql.DB, s *Site) error {
	s.Site = "only"
	k, v, err := dbKVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(k, ", ")
	idxs := strings.Join(dbIndices(k), ", ")
	stmt := fmt.Sprintf("UPDATE sites SET (%s) = (%s) WHERE site='only'", keys, idxs)
	if _, err := db.Exec(stmt, v...); err != nil {
		return err
	}
	return nil
}

// GetSite는 db에서 사이트 정보를 가지고 온다.
// 사이트 정보가 존재하지 않으면 nil과 NotFound 에러를 반환한다.
func GetSite(db *sql.DB) (*Site, error) {
	k, _, err := dbKVs(&Site{Site: "only"})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(k, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM sites LIMIT 1", keys)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, NotFound("site", "(only one yet)")
	}
	s := &Site{}
	err = scanFromRows(rows, s)
	return s, err
}
