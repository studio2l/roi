package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	if !roi.IsValidProject(prj) {
		fmt.Fprintln(os.Stderr, prj, "이 프로젝트 아이디로 적절치 않습니다.")
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
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	_, err = http.PostForm("https://localhost/api/v1/project/add", url.Values{
		"project":       []string{"test"},
		"default_tasks": []string{"fx, lit"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not add project: %v", err)
	}
	for _, row := range rows[1:] {
		formData := url.Values{}
		for i := range title {
			formData.Set(title[i], row[i])
		}
		if formData.Get("shot") == "" {
			continue
		}
		formData.Set("project", prj)
		formData.Set("shot", formData.Get("shot"))
		resp, err := http.PostForm("https://localhost/api/v1/shot/add", formData)
		if err != nil {
			log.Fatal(err)
		}
		apiResp := roi.APIResponse{}
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(apiResp.Msg)
		if apiResp.Err != "" {
			log.Fatal("could not create shot: ", apiResp.Err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
