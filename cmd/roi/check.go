package main

import "os"

// anyFileExist는 받아들인 파일 중 하나라도 존재한다면 true를
// 모든 파일이 존재하지 않는다면 false를 반환한다.
// 만일 파일 검사 중 에러가 발생하면 false와 해당 에러를 반환한다.
func anyFileExist(files ...string) (bool, error) {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, err
		}
		return true, nil
	}
	return false, nil
}
