package docker

import "testing"

func TestParseDockerInfo(t *testing.T) {
	running, stopped, err := parseDockerInfo("{\"ContainersRunning\":2,\"ContainersStopped\":3}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running != 2 || stopped != 3 {
		t.Fatalf("unexpected result: running=%d stopped=%d", running, stopped)
	}
}

func TestParseDockerInfoEmpty(t *testing.T) {
	_, _, err := parseDockerInfo("  ")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: info returned empty"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestParseDockerInfoInvalidJSON(t *testing.T) {
	_, _, err := parseDockerInfo("{bad}")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: invalid info JSON: invalid character 'b' looking for beginning of object key string"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestParseSystemDF(t *testing.T) {
	input := "" +
		"{\"Type\":\"Images\",\"TotalCount\":\"37\",\"Size\":\"14.8GB\"}\n" +
		"{\"Type\":\"Local Volumes\",\"TotalCount\":\"19\",\"Size\":\"5.2GB\"}\n" +
		"{\"Type\":\"Build Cache\",\"TotalCount\":\"3\",\"Size\":\"1.1GB\"}\n"

	images, imagesSize, volumes, volumesSize, cache, cacheSize, err := parseSystemDF(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if images != 37 || imagesSize != "14.8GB" {
		t.Fatalf("unexpected images: %d %s", images, imagesSize)
	}
	if volumes != 19 || volumesSize != "5.2GB" {
		t.Fatalf("unexpected volumes: %d %s", volumes, volumesSize)
	}
	if cache != 3 || cacheSize != "1.1GB" {
		t.Fatalf("unexpected build cache: %d %s", cache, cacheSize)
	}
}

func TestParseSystemDFInvalidJSON(t *testing.T) {
	_, _, _, _, _, _, err := parseSystemDF("{bad}\n")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: invalid system df line 1: invalid character 'b' looking for beginning of object key string"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestParseSystemDFInvalidCount(t *testing.T) {
	input := "{\"Type\":\"Images\",\"TotalCount\":\"NaN\",\"Size\":\"1GB\"}\n"
	_, _, _, _, _, _, err := parseSystemDF(input)
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: invalid system df total count on line 1: strconv.Atoi: parsing \"NaN\": invalid syntax"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
