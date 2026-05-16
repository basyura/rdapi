package term

import "testing"

func TestTruncateByDisplayWidth(t *testing.T) {
	tests := []struct {
		name  string
		value string
		width int
		want  string
	}{
		{name: "fits", value: "hello", width: 5, want: "hello"},
		{name: "ascii", value: "hello world", width: 8, want: "hello …"},
		{name: "wide", value: "日本語タイトル", width: 7, want: "日本…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateByDisplayWidth(tt.value, tt.width)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if DisplayWidth(got) > tt.width {
				t.Fatalf("display width %d exceeds %d", DisplayWidth(got), tt.width)
			}
		})
	}
}
