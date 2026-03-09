package clean

import "testing"

func TestParseContainerSize(t *testing.T) {
	bytes, label, err := parseContainerSize("12.3MB (virtual 50.2MB)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := uint64(123 * 1024 * 1024 / 10)
	if bytes != expected {
		t.Fatalf("unexpected bytes: %d", bytes)
	}
	if label != "12.3MB" {
		t.Fatalf("unexpected label: %q", label)
	}
}

func TestParseContainerSizeEmpty(t *testing.T) {
	bytes, label, err := parseContainerSize("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes != 0 || label != "0B" {
		t.Fatalf("unexpected result: %d %s", bytes, label)
	}
}

func TestParseReclaimedSpace(t *testing.T) {
	label, bytes := parseReclaimedSpace("Total reclaimed space: 2.1GB\n")
	if label != "2.1GB" {
		t.Fatalf("unexpected label: %q", label)
	}

	expected := uint64(21 * 1024 * 1024 * 1024 / 10)
	if bytes != expected {
		t.Fatalf("unexpected bytes: %d", bytes)
	}
}

func TestParseSizeToBytesInvalid(t *testing.T) {
	_, err := parseSizeToBytes("GB")
	if err == nil {
		t.Fatal("expected error")
	}
	expected := "invalid size: \"GB\""
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
