package roi

import (
	"reflect"
	"testing"
)

type TestStructForDBKeysIndicesValues struct {
	A string `db:"a"`
	B int    `db:"b"`
	C bool   `db:"c"`
}

func TestDBKeysIndicesValues(t *testing.T) {
	testVal := TestStructForDBKeysIndicesValues{
		A: "a",
		B: 1,
		C: true,
	}
	keys, idxs, vals, err := dbKeysIndicesValues(testVal)
	if err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{"a", "b", "c"}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Fatalf("dbKeys(TestStruct{}): keys: want %v, got %v", wantKeys, keys)
	}
	wantVals := []interface{}{interface{}("a"), interface{}(1), interface{}(true)}
	if !reflect.DeepEqual(vals, wantVals) {
		t.Fatalf("dbKeys(TestStruct{}): vals: want %v, got %v", wantVals, vals)
	}
	wantIdxs := []string{"$1", "$2", "$3"}
	if !reflect.DeepEqual(idxs, wantIdxs) {
		t.Fatalf("dbKeys(TestStruct{}): idxs: want %v, got %v", wantIdxs, idxs)
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
