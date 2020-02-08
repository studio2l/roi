package roi

// subtractStringSlice는 첫번째 문자열 슬라이스 요소중 두번째 문자열 슬라이스에
// 없는 요소만을 담은 새 문자열 슬라이스를 만들어 반환한다.
func subtractStringSlice(from []string, to []string) []string {
	toMap := make(map[string]bool)
	for _, el := range to {
		toMap[el] = true
	}
	result := make([]string, 0)
	for _, el := range from {
		if !toMap[el] {
			result = append(result, el)
		}
	}
	return result
}
