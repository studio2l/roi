package main

import (
	"database/sql"
	"fmt"
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
