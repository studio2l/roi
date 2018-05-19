package roi

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type dbItem interface {
	dbKeyValues() []KV
}

type KV struct {
	K string
	V string
}

func q(s string) string {
	s = strings.Replace(s, "'", "''", -1)
	return fmt.Sprint("'", s, "'")
}

func dbDate(t time.Time) string {
	ft := t.Format("2006-01-02")
	return "DATE " + q(ft)
}

func CreateTableIfNotExists(db *sql.DB, table string, fields []string) error {
	// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	fields = append(
		[]string{"id UUID PRIMARY KEY DEFAULT gen_random_uuid()"},
		fields...,
	)
	field := strings.Join(fields, ", ")
	stmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, field)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

func InsertInto(db *sql.DB, table string, item dbItem) error {
	keys := make([]string, 0)
	values := make([]string, 0)
	for _, kv := range item.dbKeyValues() {
		keys = append(keys, kv.K)
		values = append(values, kv.V)
	}
	keystr := strings.Join(keys, ", ")
	valuestr := strings.Join(values, ", ")
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, keystr, valuestr)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

func SelectAll(db *sql.DB, table string, where map[string]string) (*sql.Rows, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s", table)
	if len(where) != 0 {
		wheres := ""
		for k, v := range where {
			if wheres != "" {
				wheres += " AND "
			}
			wheres += fmt.Sprintf("(%s = '%s')", k, v)
		}
		stmt += " WHERE " + wheres
	}
	fmt.Println(stmt)
	return db.Query(stmt)
}

func AddProject(db *sql.DB, prj string) error {
	if err := InsertInto(db, "projects", Project{Code: prj}); err != nil {
		return err
	}
	// TODO: add project info, task, tracking table
	if err := CreateTableIfNotExists(db, prj+"_shots", ShotTableFields); err != nil {
		return err
	}
	return nil
}

func SelectScenes(db *sql.DB, prj string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT DISTINCT scene FROM %s_shots", prj)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	scenes := make([]string, 0)
	for rows.Next() {
		var sc string
		if err := rows.Scan(&sc); err != nil {
			return nil, err
		}
		scenes = append(scenes, sc)
	}
	sort.Slice(scenes, func(i int, j int) bool {
		if strings.Compare(scenes[i], scenes[j]) < 0 {
			return true
		}
		return false
	})
	return scenes, nil
}

func SelectShots(db *sql.DB, prj string, where map[string]string) ([]Shot, error) {
	rows, err := SelectAll(db, prj+"_shots", where)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	shots := make([]Shot, 0)
	for rows.Next() {
		var id string
		var s Shot
		if err := rows.Scan(&id, &s.Book, &s.Scene, &s.Name, &s.Status, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut); err != nil {
			return nil, err
		}
		shots = append(shots, s)
	}
	sort.Slice(shots, func(i int, j int) bool {
		if shots[i].Scene < shots[j].Scene {
			return true
		}
		if shots[i].Scene > shots[j].Scene {
			return false
		}
		return shots[i].Name <= shots[j].Name
	})
	return shots, nil
}

func AddShot(db *sql.DB, prj string, s Shot) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	if err := InsertInto(db, prj+"_shots", s); err != nil {
		return err
	}
	return nil
}
