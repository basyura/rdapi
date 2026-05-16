package term

import (
	"strings"
	"unicode/utf8"
)

func TruncateByDisplayWidth(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if DisplayWidth(value) <= maxWidth {
		return value
	}
	if maxWidth <= 1 {
		return "…"
	}

	ellipsis := "…"
	limit := maxWidth - DisplayWidth(ellipsis)
	var builder strings.Builder
	current := 0
	for _, r := range value {
		width := runeDisplayWidth(r)
		if current+width > limit {
			break
		}
		builder.WriteRune(r)
		current += width
	}
	builder.WriteString(ellipsis)
	return builder.String()
}

func DisplayWidth(value string) int {
	width := 0
	for _, r := range value {
		width += runeDisplayWidth(r)
	}
	return width
}

func runeDisplayWidth(r rune) int {
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
