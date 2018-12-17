package main

import "testing"

func TestBakeGo(t *testing.T) {
	err := bakego.Identical()
	if err != nil {
		t.Fatal(err)
	}
}