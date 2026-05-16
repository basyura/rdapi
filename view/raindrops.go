package view

import (
	"fmt"
	"io"
	"sort"

	"rdapi/api"
	"rdapi/term"
)

func FormatRaindrops(items []api.Raindrop, width int) []string {
	sorted := append([]api.Raindrop(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	lines := make([]string, 0, len(sorted)*2)
	for _, item := range sorted {
		line := fmt.Sprintf("%s : %s", formatRaindropDate(item), item.Title)
		lines = append(lines, term.TruncateByDisplayWidth(line, width))
		urlLine := fmt.Sprintf("%s%s", bookmarkURLIndent(), item.Link)
		lines = append(lines, term.TruncateByDisplayWidth(urlLine, width))
	}
	return lines
}

func PrintLines(out io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
}

func formatRaindropDate(item api.Raindrop) string {
	if item.CreatedAt.IsZero() {
		return "0000/00/00"
	}
	return item.CreatedAt.Format("2006/01/02")
}

func bookmarkURLIndent() string {
	return "             "
}
