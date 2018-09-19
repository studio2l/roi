package roi

type ordMap struct {
	keys []string
	idx  map[string]int
	val  map[string]interface{}
}

func newOrdMap() *ordMap {
	return &ordMap{
		keys: make([]string, 0),
		idx:  make(map[string]int),
		val:  make(map[string]interface{}),
	}
}

func (o *ordMap) Set(k string, v interface{}) {
	_, ok := o.idx[k]
	if !ok {
		o.keys = append(o.keys, k)
		o.idx[k] = len(o.keys) - 1
	}
	o.val[k] = v
}

func (o *ordMap) Get(k string) interface{} {
	return o.val[k]
}

func (o *ordMap) Keys() []string {
	return o.keys
}
