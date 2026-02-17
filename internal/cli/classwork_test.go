package cli

import (
	"testing"
	"time"
)

func TestParseDueInputDateOnly(t *testing.T) {
	date, tod, ok, err := parseDueInput("2026-02-16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected parse success")
	}
	if date == nil || date.Year != 2026 || date.Month != 2 || date.Day != 16 {
		t.Fatalf("unexpected date payload: %#v", date)
	}
	if tod == nil {
		t.Fatalf("expected time payload")
	}
}

func TestParseDueInputRFC3339(t *testing.T) {
	timestamp := "2026-02-16T14:30:00Z"
	date, tod, ok, err := parseDueInput(timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected parse success")
	}
	if date.Year != 2026 || date.Month != 2 || date.Day != 16 {
		t.Fatalf("unexpected date fields: %#v", date)
	}
	if tod.Hours != 14 || tod.Minutes != 30 {
		t.Fatalf("unexpected time fields: %#v", tod)
	}
}

func TestParseDueInputInvalid(t *testing.T) {
	_, _, _, err := parseDueInput(time.Now().Format(time.Kitchen))
	if err == nil {
		t.Fatalf("expected parse error for invalid format")
	}
}
