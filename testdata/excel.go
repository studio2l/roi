package main

import (
	"database/sql"
	"log"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	xl, err := excelize.OpenFile("roi.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	rows := xl.GetRows("Sheet1")
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
	shots := make([]Shot, 0)
	for i, row := range rows[1:] {
		if i == 0 {
			continue
		}
		xlrow := make(map[string]string)
		for j := range title {
			k := title[j]
			v := row[j]
			xlrow[k] = v
		}
		shot := ShotFromMap(xlrow)
		shots = append(shots, shot)
	}

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	for _, shot := range shots {
		if err := insertInto(db, "rd7_shot", shot); err != nil {
			log.Fatal(err)
		}
	}

}
