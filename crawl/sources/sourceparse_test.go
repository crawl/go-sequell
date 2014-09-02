package sources

import (
	"fmt"
	"testing"

	"github.com/greensnark/go-sequell/resource"
)

func TestSources(t *testing.T) {
	schema := resource.ResourceYamlMustExist("config/sources.yml")
	src, err := Sources(schema, "test")
	if err != nil {
		t.Errorf("Error parsing sources: %s", err)
		return
	}
	expectedCount := 8
	if len(src.Servers) != expectedCount {
		t.Errorf("Expected %d sources, got %d", expectedCount, len(src.Servers))
		return
	}

	cao := src.Server("cao")
	if cao == nil {
		t.Errorf("Couldn't find source cao in %s", src)
	}

	for _, srv := range src.Servers {
		fmt.Println()
		fmt.Println(srv.Name)
		for i, log := range srv.Logfiles {
			fmt.Printf("%02d) %s\n", i+1, log)
		}
		n := len(cao.Logfiles)
		for i, log := range srv.Milestones {
			fmt.Printf("%02d) %s\n", i+1+n, log)
		}
	}
}
