package roi

import (
	"log"
	"testing"
)

type kv struct {
	k string
	v interface{}
}

func sameOrdMap(a, b *ordMap) bool {
	for i, av := range a.keys {
		bv := b.keys[i]
		if av != bv {
			return false
		}
	}
	for ak, av := range a.idx {
		bv, ok := b.idx[ak]
		if !ok {
			return false
		}
		if av != bv {
			return false
		}
	}
	for ak, av := range a.val {
		bv, ok := b.val[ak]
		if !ok {
			return false
		}
		if av != bv {
			return false
		}
	}
	return true
}

func TestOrdMapSet(t *testing.T) {
	cases := []struct {
		name string
		kvs  []kv
		want *ordMap
	}{
		{
			name: "basic",
			kvs: []kv{
				{"a", 101},
				{"b", 102},
			},
			want: &ordMap{
				keys: []string{"a", "b"},
				idx: map[string]int{
					"a": 0,
					"b": 1,
				},
				val: map[string]interface{}{
					"a": 101,
					"b": 102,
				},
			},
		},
		{
			name: "overwrap",
			kvs: []kv{
				{"a", 101},
				{"b", 102},
				{"b", 103},
				{"b", 104},
				{"c", 105},
			},
			want: &ordMap{
				keys: []string{"a", "b", "c"},
				idx: map[string]int{
					"a": 0,
					"b": 1,
					"c": 2,
				},
				val: map[string]interface{}{
					"a": 101,
					"b": 104,
					"c": 105,
				},
			},
		},
	}
	for _, c := range cases {
		m := newOrdMap()
		for _, kv := range c.kvs {
			m.Set(kv.k, kv.v)
		}
		if !sameOrdMap(m, c.want) {
			log.Fatalf("TestOrdMapSet: %s: want %v got %v", c.name, *c.want, *m)
		}
	}
}

func TestOrdMapDelete(t *testing.T) {
	fromMap := &ordMap{
		keys: []string{"a", "b", "c"},
		idx: map[string]int{
			"a": 0,
			"b": 1,
			"c": 2,
		},
		val: map[string]interface{}{
			"a": 101,
			"b": 102,
			"c": 103,
		},
	}

	cases := []struct {
		name    string
		delKeys []string
		want    *ordMap
	}{
		{
			name:    "one",
			delKeys: []string{"c"},
			want: &ordMap{
				keys: []string{"a", "b"},
				idx: map[string]int{
					"a": 0,
					"b": 1,
				},
				val: map[string]interface{}{
					"a": 101,
					"b": 102,
				},
			},
		},
		{
			name:    "two",
			delKeys: []string{"a", "c"},
			want: &ordMap{
				keys: []string{"b"},
				idx: map[string]int{
					"b": 0,
				},
				val: map[string]interface{}{
					"b": 102,
				},
			},
		},
	}
	for _, c := range cases {
		m := &(*fromMap) // 이번 루프를 위해 fromMap에서 복사한 맵 생성
		for _, k := range c.delKeys {
			m.Delete(k)
		}
		if !sameOrdMap(m, c.want) {
			log.Fatalf("TestOrdMapSet: %s: got %v, want %v", c.name, *c.want, *m)
		}
	}
}
