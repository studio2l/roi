package main

// appendIfNotExist는 문자열 슬라이스에 해당 문자열이 없다면 더한다.
func appendIfNotExist(strs []string, v string) []string {
	found := false
	for _, s := range strs {
		if s == v {
			found = true
			break
		}
	}
	if !found {
		strs = append(strs, v)
	}
	return strs
}

// removeIfExist는 문자열 슬라이스에 해당 문자열이 존재한다면 지운다.
func removeIfExist(strs []string, v string) []string {
	nstrs := make([]string, 0)
	for _, s := range strs {
		if s != v {
			nstrs = append(nstrs, s)
		}
	}
	return nstrs
}
