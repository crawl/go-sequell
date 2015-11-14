package data

import (
	"reflect"
	"testing"
)

func TestArray(t *testing.T) {
	uniques := CrawlData().StringSlice("uniques")
	expected := []string{"Ijyb", "Blork the orc", "Blork", "Urug"}
	if !reflect.DeepEqual(uniques[0:4], expected) {
		t.Errorf("Bad value for uniques Array: %v, expected it to start with: %v", uniques, expected)
	}
}
