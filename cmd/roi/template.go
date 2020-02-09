package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// templates에는 사용자에게 보일 페이지의 템플릿이 담긴다.
var templates *template.Template

// executeTemplate은 템플릿과 정보를 이용하여 w에 응답한다.
// templates.ExecuteTemplate 대신 이 함수를 쓰는 이유는 개발모드일 때
// 재 컴파일 없이 업데이트된 템플릿을 사용할 수 있기 때문이다.
func executeTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if dev {
		parseTemplate()
	}
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

// parseTemplate은 tmpl 디렉토리 안의 html파일들을 파싱하여 http 응답에 사용될 수 있도록 한다.
func parseTemplate() {
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"hasThumbnail":        hasThumbnail,
		"stringFromTime":      stringFromTime,
		"stringFromDate":      stringFromDate,
		"shortStringFromDate": shortStringFromDate,
		"isSunday":            isSunday,
		"dayColorInTimeline":  dayColorInTimeline,
		"mod":                 func(i, m int) int { return i % m },
		"sub":                 func(a, b int) int { return a - b },
		"fieldJoin":           fieldJoin,
		"spaceJoin":           func(words []string) string { return strings.Join(words, " ") },
	}).ParseGlob("tmpl/*.html"))
}

// 아래는 템플릿 안에서 사용되는 함수들이다.
//

// hasThumbnail은 해당 특정 프로젝트 샷에 썸네일이 있는지 검사한다.
//
// 주의: 만일 썸네일 파일 검사시 에러가 나면 이 함수는 썸네일이 있다고 판단한다.
// 이 함수는 템플릿 안에서 쓰이기 때문에 프론트 엔드에서 한번 더 검사하게
// 만들기 위해서이다.
func hasThumbnail(id string) bool {
	_, err := os.Stat(fmt.Sprintf("data/show/%s/thumbnail.png", id))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return true // 함수 주석 참고
	}
	return true
}

// isSunday는 해당일이 일요일인지를 검사한다.
func isSunday(day string) bool {
	t, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return false
	}
	wd := t.Weekday()
	return wd == time.Sunday
}

// stringFromTime은 시간을 rfc3339 형식에서 존 정보를 제외한 문자열을 반환한다.
func stringFromTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02T15:04:05")
}

// stringFromDate는 시간을 rfc3339 형식의 문자열로 표현하되 '연-월-일' 만 표시한다.
func stringFromDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02")
}

// shortStringFromDate는 시간을 rfc3339 형식의 문자열로 표현하되 '월-일' 만 표시한다.
func shortStringFromDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("01-02")
}

// timeFromString는 존 정보를 제외한 rfc3339 형식의 날짜 또는 시간 문자열에서
// 시간을 얻는다. 받은 문자열이 형식에 맞지 않으면 에러를 반환한다.
func timeFromString(s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
}

// parseTimeforms는 http.Request.Form에서 시간 형식의 Form에 대해 파싱해
// 맵으로 반환한다. 만일 받아들인 문자열이 시간 형식에 맞지 않으면 에러를 낸다.
func parseTimeForms(form url.Values, keys ...string) (map[string]time.Time, error) {
	tforms := make(map[string]time.Time)
	for _, k := range keys {
		v := form.Get(k)
		if v == "" {
			continue
		}
		t, err := timeFromString(v)
		if err != nil {
			return nil, fmt.Errorf("invalid time string '%s' for '%s'", v, k)
		}
		tforms[k] = t
	}
	return tforms, nil
}

// dayColorInTimeline은 해당 일의 태스크 갯수를 받아 이를 UI 색상으로 표현한다.
func dayColorInTimeline(i int) string {
	switch i {
	case 0:
		return "grey"
	case 1:
		return "yellow"
	case 2:
		return "orange"
	}
	return "red"
}
