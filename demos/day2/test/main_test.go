package main

import "testing"

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"valid", "42", 42, false},
		{"invalid", "abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

// go test -bench BenchmarkParseInt -count 5

func BenchmarkParseInt(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"small", "42"},
		{"big", "12345634856"},
		{"toobig", "12345634856"},
		{"invalid", "abc"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = ParseInt(tt.input)
			}
		})
	}
}
