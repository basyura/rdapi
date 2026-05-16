package term

import (
	"strings"
	"unicode/utf8"
)

func TruncateByDisplayWidth(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if GetDisplayWidth(value) <= maxWidth {
		return value
	}
	if maxWidth <= 1 {
		return "…"
	}

	ellipsis := "…"
	limit := maxWidth - GetDisplayWidth(ellipsis)
	var builder strings.Builder
	current := 0
	for _, r := range value {
		width := getRuneDisplayWidth(r)
		if current+width > limit {
			break
		}
		builder.WriteRune(r)
		current += width
	}
	builder.WriteString(ellipsis)
	return builder.String()
}

func GetDisplayWidth(value string) int {
	width := 0
	for _, r := range value {
		width += getRuneDisplayWidth(r)
	}
	return width
}

func getRuneDisplayWidth(r rune) int {
	if r == '\t' {
		return 4
	}
	if r < 0x20 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if utf8.RuneLen(r) > 1 {
		return 2
	}
	return 1
}
