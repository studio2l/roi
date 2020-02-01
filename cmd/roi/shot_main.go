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

func shotMain(args []string) {
	var (
		addr     string
		insecure bool
		show     string
		sheet    string
	)
	shotFlag := flag.NewFlagSet("shot", flag.ExitOnError)
	addrDefault := "localhost:80:443"
	addrHelp := `binding address and it's http/https port.

when two ports are specified, first port is for http and second is for https.
unless -insecure flag set, it will transfer data through https port.
with -insecure flag, it will use http port.
ex) localhost:80:443

when only one port is specified, the port will be used for current protocol.
ex) localhost:80, localhost:443

when no port is specified, it is same as :80:443.
ex) localhost

when ROI_ADDR environment variable is not empty, it will use the value as default.

`
	addrEnv := os.Getenv("ROI_ADDR")
	if addrEnv != "" {
		addrDefault = addrEnv
		addrHelp += "currently the default value is comming from ROI_ADDR"
	}
	shotFlag.StringVar(&addr, "addr", addrDefault, addrHelp)
	shotFlag.BoolVar(&insecure, "insecure", false, "use insecure http protocol instead of https.")
	shotFlag.StringVar(&show, "show", "", "샷을 추가할 프로젝트, 없으면 엑셀 파일이름을 따른다.")
	shotFlag.StringVar(&sheet, "sheet", "Sheet1", "엑셀 시트명")
	shotFlag.Parse(args)

	if len(shotFlag.Args()) == 0 {
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "엑셀 파일 경로를 입력하세요.")
		os.Exit(1)
	}
	f := shotFlag.Arg(0)

	if show == "" {
		fname := filepath.Base(f)
		show = strings.TrimSuffix(fname, filepath.Ext(fname))
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
	protocol := "https://"
	if insecure {
		protocol = "http://"
	}
	addrs := strings.Split(addr, ":")
	site := addrs[0]
	port := ""
	if len(addrs) == 3 {
		port = addrs[2]
		if insecure {
			port = addrs[1]
		}
	} else if len(addrs) == 2 {
		port = addrs[1]
	} else if len(addrs) == 1 {
		port = "443"
		if insecure {
			port = "80"
		}
	} else {
		log.Fatalf("invalid -addr flag value: %s\n", addr)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	_, err = http.PostForm(protocol+site+":"+port+"/api/v1/show/add", url.Values{
		"show":          []string{"test"},
		"default_tasks": []string{"fx, lit"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not add show: %v", err)
	}
	for _, row := range rows[1:] {
		formData := url.Values{}
		for i := range title {
			formData.Set(title[i], row[i])
		}
		if formData.Get("shot") == "" {
			continue
		}
		id := show + "/shot/" + formData.Get("shot")
		formData.Set("id", id)
		resp, err := http.PostForm(protocol+site+":"+port+"/api/v1/shot/add", formData)
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
