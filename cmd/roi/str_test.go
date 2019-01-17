package main

import (
	"reflect"
	"testing"
)

func TestFields(t *testing.T) {
	cases := []struct {
		s    string
		sep  string
		want []string
	}{
		{
			s:    ", a, b, c, ",
			sep:  ",",
			want: []string{"a", "b", "c"},
		},
		{
			s:    "a,  , , b, , c, ",
			sep:  ",",
			want: []string{"a", "b", "c"},
		},
	}
	for _, c := range cases {
		got := fields(c.s, c.sep)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("fleids: got: %v, want: %v", got, c.want)
		}
	}
}

func TestAtoi(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{
			s:    "1",
			want: 1,
		},
		{
			s:    "+2",
			want: 2,
		},
		{
			s:    "3+",
			want: 0,
		},
		{
			s:    "not a int",
			want: 0,
		},
		{
			s:    "",
			want: 0,
		},
	}
	for _, c := range cases {
		got := atoi(c.s)
		if got != c.want {
			t.Fatalf("atoi: got: %v, want: %v", got, c.want)
		}
	}
}
