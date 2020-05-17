package main

import (
	"reflect"
	"testing"
)

func TestGetThumbnailsInExcel(t *testing.T) {
	want := []string{
		"/vfx/thumbnail1.jpg",
		"/vfx/thumbnail2.jpg",
		"/vfx/thumbnail3.jpg",
		"/vfx/thumbnail4.jpg",
		"/vfx/thumbnail5.jpg",
		"/vfx/thumbnail6.jpg",
		"/vfx/thumbnail7.jpg",
		"/vfx/thumbnail8.jpg",
		"/vfx/thumbnail9.jpg",
		"/vfx/thumbnail10.jpg",
	}
	got, err := getThumbnailsInExcel("testdata/test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}
