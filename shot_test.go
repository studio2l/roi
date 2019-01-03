package roi

import (
	"reflect"
	"testing"

	"github.com/lib/pq"
)

func TestOrdMapFromShot(t *testing.T) {
	s := Shot{
		ID:            "CG_0010",
		ProjectID:     "test",
		Status:        "waiting",
		EditOrder:     10,
		Description:   "the first shot",
		CGDescription: "highend cg shot",
		TimecodeIn:    "00:00:00:00",
		TimecodeOut:   "00:00:00:01",
		Duration:      1,
		Tags:          []string{"money-shot"},
	}
	got := ordMapFromShot(s)

	want := newOrdMap()
	want.Set("id", "CG_0010")
	want.Set("project_id", "test")
	want.Set("status", "waiting")
	want.Set("edit_order", 10)
	want.Set("description", "the first shot")
	want.Set("cg_description", "highend cg shot")
	want.Set("timecode_in", "00:00:00:00")
	want.Set("timecode_out", "00:00:00:01")
	want.Set("duration", 1)
	want.Set("tags", pq.Array([]string{"money-shot"}))

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
