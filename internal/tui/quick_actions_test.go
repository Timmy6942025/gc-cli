package tui

import "testing"

func TestParseTodoRef(t *testing.T) {
	courseID, courseWorkID, submissionID, err := parseTodoRef("abc::def::ghi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if courseID != "abc" || courseWorkID != "def" || submissionID != "ghi" {
		t.Fatalf("unexpected parse result: %q %q %q", courseID, courseWorkID, submissionID)
	}
}

func TestParseTodoRefInvalid(t *testing.T) {
	if _, _, _, err := parseTodoRef("abc::def"); err == nil {
		t.Fatalf("expected error for invalid todo ref")
	}
}

func TestJoinTodoRef(t *testing.T) {
	got := joinTodoRef("course1", "work1", "sub1")
	if got != "course1::work1::sub1" {
		t.Fatalf("unexpected join result: %q", got)
	}
}
