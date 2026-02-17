package tui

import "testing"

func TestSplitCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "simple",
			input: "classwork list --course 123",
			want:  []string{"classwork", "list", "--course", "123"},
		},
		{
			name:  "quoted",
			input: "stream post --text \"hello world\"",
			want:  []string{"stream", "post", "--text", "hello world"},
		},
		{
			name:  "single_quoted",
			input: "stream post --text 'hello world'",
			want:  []string{"stream", "post", "--text", "hello world"},
		},
		{
			name:    "unclosed_quote",
			input:   "stream post --text \"hello",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := splitCommandLine(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("length mismatch: got %d want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("arg[%d] mismatch: got %q want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
