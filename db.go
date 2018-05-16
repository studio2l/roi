package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type dbItem interface {
	dbKeyTypeValues() []KTV
}

type KTV struct {
	K string
	T string
	V string
}

func q(s string) string {
	return fmt.Sprint("'", s, "'")
}

func createTableIfNotExists(db *sql.DB, table string, item dbItem) error {
	fields := []string{
		// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
		"id UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	}
	for _, ktv := range item.dbKeyTypeValues() {
		f := ktv.K + " " + ktv.T
		fields = append(fields, f)
	}
	field := strings.Join(fields, ", ")
	stmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, field)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

func insertInto(db *sql.DB, table string, item dbItem) error {
	keys := make([]string, 0)
	values := make([]string, 0)
	for _, ktv := range item.dbKeyTypeValues() {
		keys = append(keys, ktv.K)
		values = append(values, ktv.V)
	}
	keystr := strings.Join(keys, ", ")
	valuestr := strings.Join(values, ", ")
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, keystr, valuestr)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

func selectAll(db *sql.DB, table string) (*sql.Rows, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s", table)
	fmt.Println(stmt)
	return db.Query(stmt)
}

func selectShots(db *sql.DB, prj string) ([]Shot, error) {
	rows, err := selectAll(db, prj+"_shot")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	shots := make([]Shot, 0)
	for rows.Next() {
		var id string
		var s Shot
		if err := rows.Scan(&id, &s.Project, &s.Book, &s.Scene, &s.Name, &s.Status, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut); err != nil {
			return nil, err
		}
		shots = append(shots, s)
	}
	return shots, nil
}
