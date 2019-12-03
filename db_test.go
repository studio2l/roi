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
	keys, vals, err := dbKVs(testVal)
	if err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Fatalf("keys: want %v, got %v", wantKeys, keys)
	}
	wantVals := []interface{}{"a", 1, true, pq.Array([]int{1, 2, 3}), pq.Array([]string(nil))}
	if !reflect.DeepEqual(vals, wantVals) {
		t.Fatalf("vals: want %v, got %v", wantVals, vals)
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
