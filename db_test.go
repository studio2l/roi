package roi

import (
	"reflect"
	"testing"
)

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
