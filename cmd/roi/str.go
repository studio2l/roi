package main

import (
	"strconv"
	"strings"
)

// fields는 문자열을 콤마로 잘라 슬라이스로 반환한다.
// 잘린 문자열 양 옆의 빈 문자열은 함께 지워진다.
// 혹시 필드가 빈 문자열이라면 그 항목은 포함되지 않는다.
//
// 예) fields("a, b, c,, ") => []string{"a", "b", "c"}
func fields(s string) []string {
	ss := strings.Split(s, ",")
	fs := make([]string, 0, len(ss))
	for _, f := range ss {
		f = strings.TrimSpace(f)
		if f != "" {
			fs = append(fs, f)
		}
	}
	return fs
}

// atoi는 받아 들인 문자열을 정수로 변환한다.
// 만일 변환할 수 없으면 0을 반환한다.
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
