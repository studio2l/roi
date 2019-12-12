package roi

import (
	"reflect"
	"testing"

	"github.com/lib/pq"
)

type TestStructForDBKVs struct {
	A string   `db:"a"`
	B int      `db:"b"`
	C bool     `db:"c"`
	D []int    `db:"d"`
	E []string `db:"e"`
}

func TestDBKVs(t *testing.T) {
	testVal := TestStructForDBKVs{
		A: "a",
		B: 1,
		C: true,
		D: []int{1, 2, 3},
		E: nil,
	}
	ks, is, vs, err := dbKIVs(testVal)
	if err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(ks, wantKeys) {
		t.Fatalf("keys: want %v, got %v", wantKeys, ks)
	}
	wantIndices := []string{"$1", "$2", "$3", "$4", "$5"}
	if !reflect.DeepEqual(is, wantIndices) {
		t.Fatalf("keys: want %v, got %v", wantKeys, ks)
	}
	// 마지막이 pq.Array([]string(nil))이 아닌 이유는 dbKVs함수가 nil 슬라이스를 만들지 않기 때문이다.
	wantVals := []interface{}{"a", 1, true, pq.Array([]int{1, 2, 3}), pq.Array([]string{})}
	if !reflect.DeepEqual(vs, wantVals) {
		t.Fatalf("vals: want %v, got %v", wantVals, vs)
	}
}

func TestDBIndices(t *testing.T) {
	cases := []struct {
		keys []string
		want []string
	}{
		{
			keys: []string{},
			want: []string{},
		},
		{
			keys: []string{"a"},
			want: []string{"$1"},
		},
		{
			keys: []string{"a", "b", "c"},
			want: []string{"$1", "$2", "$3"},
		},
	}
	for _, c := range cases {
		got := dbIndices(c.keys)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("TestDBIndices: got %v, want %v", got, c.want)
		}
	}
}
