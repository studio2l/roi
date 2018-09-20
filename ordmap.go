package roi

// ordMap은 순서를 가지는 맵이다.
// db 인 아웃에 쓰려는 목적에 따라 키는 항상 문자열이다.
// 값은 어떤 형식이든 상관없다. (interface{})
type ordMap struct {
	keys []string
	idx  map[string]int
	val  map[string]interface{}
}

// newOrdMap은 새 ordMap을 생성하고 초기화한다.
func newOrdMap() *ordMap {
	return &ordMap{
		keys: make([]string, 0),
		idx:  make(map[string]int),
		val:  make(map[string]interface{}),
	}
}

// Set은 맵에 해당 키에 대한 값을 추가하거나, 재설정한다.
func (o *ordMap) Set(k string, v interface{}) {
	_, ok := o.idx[k]
	if !ok {
		o.keys = append(o.keys, k)
		o.idx[k] = len(o.keys) - 1
	}
	o.val[k] = v
}

// Get은 맵에서 해당 키의 값을 받아온다. 만일 키가 존재하지 않으면 nil이 반환된다.
func (o *ordMap) Get(k string) interface{} {
	return o.val[k]
}

// Keys는 맵에 추가된 순서에 따른 키모음을 []string 형태로 반환한다.
func (o *ordMap) Keys() []string {
	return o.keys
}
