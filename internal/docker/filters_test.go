package docker

import (
	"reflect"
	"testing"
)

func TestAppendFilterArgs(t *testing.T) {
	got := appendFilterArgs([]string{"ps"}, "name", []string{"api", "worker"})
	want := []string{"ps", "--filter", "name=api", "--filter", "name=worker"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected filter args: %#v", got)
	}
}
