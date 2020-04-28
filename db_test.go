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
	gotKeys := dbKeys(testVal)
	wantKeys := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("keys: want %v, got %v", wantKeys, gotKeys)
	}
	gotIdxs := dbIdxs(testVal)
	wantIdxs := []string{"$1", "$2", "$3", "$4", "$5"}
	if !reflect.DeepEqual(gotIdxs, wantIdxs) {
		t.Fatalf("idxs: want %v, got %v", wantIdxs, gotIdxs)
	}
	gotVals := dbVals(testVal)
	wantVals := []interface{}{"a", 1, true, pq.Array([]int{1, 2, 3}), pq.Array([]string{})}
	// wantVals의 마지막 항목이 pq.Array([]string(nil))이 아닌 이유는
	// dbKVs함수가 nil 슬라이스를 만들지 않기 때문이다.
	if !reflect.DeepEqual(gotVals, wantVals) {
		t.Fatalf("vals: want %v, got %v", wantVals, gotVals)
	}
}
