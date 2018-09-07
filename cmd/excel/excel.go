package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/studio2l/roi"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	var (
		prj   string
		sheet string
	)
	flag.StringVar(&prj, "prj", "", "샷을 추가할 프로젝트, 없으면 엑셀 파일이름을 따른다.")
	flag.StringVar(&sheet, "sheet", "Sheet1", "엑셀 시트명")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "엑셀 파일 경로를 입력하세요.")
		os.Exit(1)
	}
	f := flag.Arg(0)

	if prj == "" {
		fname := filepath.Base(f)
		prj = strings.TrimSuffix(fname, filepath.Ext(fname))
	}
	if !roi.IsValidProjectName(prj) {
		fmt.Fprintln(os.Stderr, prj, "이 프로젝트 이름으로 적절치 않습니다.")
		os.Exit(1)
	}

	xl, err := excelize.OpenFile(f)
	if err != nil {
		log.Fatal(err)
	}
	rows := xl.GetRows(sheet)
	if len(rows) == 0 {
		return
	}
	row0 := rows[0]
	title := make(map[int]string)
	for j, cell := range row0 {
		if cell != "" {
			title[j] = cell
		}
	}
	shots := make([]roi.Shot, 0)
	for _, row := range rows[1:] {
		xlrow := make(map[string]string)
		for j := range title {
			k := title[j]
			v := row[j]
			xlrow[k] = v
		}
		shot := roi.ShotFromMap(xlrow)
		shots = append(shots, shot)
	}

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	if err := roi.CreateTableIfNotExists(db, "projects", roi.ProjectTableFields); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 기존의 데이터를 일단 지운다. 더 쉽게 테스트하기 위한 임시방편이다.
	if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s_shots", prj)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := db.Exec(fmt.Sprintf("DELETE FROM projects WHERE code='%s'", prj)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := roi.AddProject(db, prj); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, shot := range shots {
		if err := roi.InsertInto(db, prj+"_shots", shot); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
