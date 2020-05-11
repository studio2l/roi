package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Site는 현재 스튜디오를 뜻한다.
type Site struct {
	// 현재로서는 빈 이름의 하나의 사이트만 존재한다.
	// 추후 여러 사이트로 확장할것인지 고민중이다.
	Site            string   `db:"site"`
	VFXSupervisors  []string `db:"vfx_supervisors"`
	VFXProducers    []string `db:"vfx_producers"`
	CGSupervisors   []string `db:"cg_supervisors"`
	ProjectManagers []string `db:"project_managers"`
	// 샷에 생성할 수 있는 태스크
	Tasks []string `db:"tasks"`
	// 샷이 생성될 때 기본적으로 생기는 태스크
	DefaultShotTasks  []string `db:"default_shot_tasks"`
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
	tasks STRING[] NOT NULL,
	default_shot_tasks STRING[] NOT NULL,
	default_asset_tasks STRING[] NOT NULL,
	leads STRING[] NOT NULL
)`

var siteDBKey string = strings.Join(dbKeys(&Site{}), ", ")
var siteDBIdx string = strings.Join(dbIdxs(&Site{}), ", ")
var _ []interface{} = dbVals(&Site{})

// DefaultSite는 기본적으로 제공되는 사이트이다.
var DefaultSite = &Site{
	Tasks: []string{
		"mod",
		"rig",
		"tex",
		"motion",
		"match",
		"ani",
		"fx",
		"lit",
		"matte",
		"comp",
	},
	DefaultShotTasks: []string{
		"comp",
	},
	DefaultAssetTasks: []string{
		"mod",
		"rig",
		"tex",
	},
}

// verifySite는 받아들인 사이트가 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifySite(db *sql.DB, s *Site) error {
	if s == nil {
		return fmt.Errorf("nil site")
	}
	hasTask := make(map[string]bool)
	for _, task := range s.Tasks {
		err := verifyTaskName(task)
		if err != nil {
			return fmt.Errorf("invalid site: %w", err)
		}
		hasTask[task] = true
	}
	for _, task := range s.DefaultShotTasks {
		if !hasTask[task] {
			return fmt.Errorf("invalid site: task %q not specified but used as default shot task", task)
		}
	}
	for _, task := range s.DefaultAssetTasks {
		if !hasTask[task] {
			return fmt.Errorf("invalid site: task %q not specified but used as default asset task", task)
		}
	}
	return nil
}

// AddSite는 DB에 하나의 사이트를 생성한다.
// 현재는 하나의 사이트만 지원하기 때문에 db생성시 한번만 사용되어야 한다.
func AddSite(db *sql.DB) error {
	err := verifySite(db, DefaultSite)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO sites (%s) VALUES (%s)", siteDBKey, siteDBIdx), dbVals(DefaultSite)...),
	}
	return dbExec(db, stmts)
}

// UpdateSite는 DB의 사이트 정보를 업데이트한다.
func UpdateSite(db *sql.DB, s *Site) error {
	err := verifySite(db, s)
	if err != nil {
		return err
	}
	oldS, err := GetSite(db)
	if err != nil {
		return err
	}
	delTasks := subtractStringSlice(oldS.Tasks, s.Tasks)
	for _, task := range delTasks {
		err := SiteMustNotHaveTask(db, task)
		if err != nil {
			return fmt.Errorf("could not delete site shot task: %w", err)
		}
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE sites SET (%s) = (%s)", siteDBKey, siteDBIdx), dbVals(s)...),
	}
	return dbExec(db, stmts)
}

// SiteMustNotHaveTask는 사이트 내의 샷에 해당 태스크가 하나라도 있으면 에러를 반환한다.
func SiteMustNotHaveTask(db *sql.DB, task string) error {
	s := &Unit{}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM units WHERE $1::string = ANY(units.tasks)", unitDBKey), task)
	err := dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, s)
	})
	if err == nil {
		return BadRequest(fmt.Sprintf("unit %q has task %q (and there's possibly more)", s.ID(), task))
	}
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}

// GetSite는 db에서 사이트 정보를 가지고 온다.
// 사이트 정보가 존재하지 않으면 nil과 NotFound 에러를 반환한다.
func GetSite(db *sql.DB) (*Site, error) {
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM sites LIMIT 1", siteDBKey))
	s := &Site{}
	err := dbQueryRow(db, stmt, func(row *sql.Row) error {
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

func DeleteSite(db *sql.DB) error {
	stmts := []dbStatement{
		dbStmt("DELETE FROM sites"),
	}
	return dbExec(db, stmts)
}
