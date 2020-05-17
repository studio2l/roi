// +build js,wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	js.Global().Set("wasmGetThumnailsInExcel", js.FuncOf(wasmGetThumbnailsInExcel))
	select {}
}

type Ret struct {
	Val interface{}
	Err error
}

func Err(msg string, vals ...interface{}) Ret {
	return Ret{Err: fmt.Errorf(msg, vals...)}
}

func Val(v interface{}) Ret {
	return Ret{Val: v}
}

func wasmGetThumbnailsInExcel(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		return "need excel file path as argument"
	}
	xlfile := args[0].String()
	thumbs, err := getThumbnailsInExcel(xlfile)
	if err != nil {
		return err.Error()
	}
	return thumbs
}

func getThumbnailsInExcel(xlfile string) ([]string, error) {
	xl, err := excelize.OpenFile(xlfile)
	if err != nil {
		return nil, err
	}
	idx := xl.GetSheetIndex("roi")
	if idx == 0 {
		// roi 시트를 못찾으면 첫번째 열을 사용한다.
		idx = 1
	}
	sh := xl.GetSheetName(idx)
	firstRow := xl.GetRows(sh)[0]
	thumbCol := -1
	for i, col := range firstRow {
		if col == "thumbnail" {
			thumbCol = i
		}
	}
	if thumbCol == -1 {
		return nil, fmt.Errorf("couldn't find 'thumbnail' row")
	}
	thumbs := make([]string, 0)
	for _, row := range xl.GetRows(sh)[1:] {
		th := row[thumbCol]
		th = strings.TrimSpace(th)
		if th == "" {
			continue
		}
		thumbs = append(thumbs, row[thumbCol])
	}
	return thumbs, nil
}
