package movie

import "testing"

func TestParseSehuatangMovieName(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "uppercase", title: "[中文字幕] ABP-123 标题", want: "ABP-123"},
		{name: "lowercase", title: "example fc2ppv-1234567 title", want: "FC2PPV-1234567"},
		{name: "first match", title: "SSIS-123 / SSIS-456", want: "SSIS-123"},
		{name: "missing", title: "没有影片番号", want: ""},
		{name: "empty", title: "  ", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSehuatangMovieName(tt.title); got != tt.want {
				t.Fatalf("ParseSehuatangMovieName(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}
