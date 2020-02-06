package roi

import (
	"database/sql"
	"errors"
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
	// 샷에 생성할 수 있는 태스크
	ShotTasks []string `db:"shot_tasks"`
	// 샷이 생성될 때 기본적으로 생기는 태스크
	DefaultShotTasks []string `db:"default_shot_tasks"`
	// 애셋에 생성할 수 있는 태스크
	AssetTasks []string `db:"asset_tasks"`
	// 애셋이 생성될 때 기본적으로 생기는 태스크
	DefaultAssetTasks []string `db:"default_asset_tasks"`
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
	shot_tasks STRING[] NOT NULL,
	default_shot_tasks STRING[] NOT NULL,
	asset_tasks STRING[] NOT NULL,
	default_asset_tasks STRING[] NOT NULL,
	leads STRING[] NOT NULL
)`

// AddSite는 DB에 빈 사이트를 생성한다.
func AddSite(db *sql.DB) error {
	ks, is, vs, err := dbKIVs(&Site{Site: "only"})
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO sites (%s) VALUES (%s)", keys, idxs), vs...),
	}
	return dbExec(db, stmts)
}

// UpdateSite는 DB의 사이트 정보를 업데이트한다.
func UpdateSite(db *sql.DB, s *Site) error {
	s.Site = "only"
	ks, is, vs, err := dbKIVs(s)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE sites SET (%s) = (%s) WHERE site='only'", keys, idxs), vs...),
	}
	return dbExec(db, stmts)
}

// GetSite는 db에서 사이트 정보를 가지고 온다.
// 사이트 정보가 존재하지 않으면 nil과 NotFound 에러를 반환한다.
func GetSite(db *sql.DB) (*Site, error) {
	ks, _, _, err := dbKIVs(&Site{Site: "only"})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM sites LIMIT 1", keys))
	s := &Site{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("site", "(only one yet)")
		}
		return nil, err
	}
	return s, err
}
