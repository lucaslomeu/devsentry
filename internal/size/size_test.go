package size

import "testing"

func TestParseAndFormat(t *testing.T) {
	bytes, err := Parse("1.5GB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := Format(bytes); got != "1.5GiB" {
		t.Fatalf("unexpected formatted size: %q", got)
	}
}

func TestParseRejectsUnknownUnit(t *testing.T) {
	_, err := Parse("12XB")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != `unknown size unit: "XB"` {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
