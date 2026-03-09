package docker

import (
	"reflect"
	"testing"
)

func TestParseDockerPSOutput(t *testing.T) {
	input := "" +
		"{\"Names\":\"api\",\"Image\":\"nginx\",\"Status\":\"Up 2 minutes\",\"Ports\":\"0.0.0.0:80->80/tcp\"}\n" +
		"{\"Names\":\"db\",\"Image\":\"postgres\",\"Status\":\"Up\",\"Ports\":\"\"}\n"

	got, err := parseDockerPSOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []ContainerInfo{
		{
			Name:   "api",
			Image:  "nginx",
			Status: "Up 2 minutes",
			Ports:  "0.0.0.0:80->80/tcp",
		},
		{
			Name:   "db",
			Image:  "postgres",
			Status: "Up",
			Ports:  "",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestParseDockerPSOutputInvalidJSON(t *testing.T) {
	_, err := parseDockerPSOutput("{bad}\n")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "invalid line 1 in docker ps: invalid character 'b' looking for beginning of object key string"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
