package ui

import (
	"fmt"
	"strings"

	"github.com/aiomayo/hdf/internal/process"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func RenderTable(procs []process.Info, verbose bool) string {
	headers := []string{"PID", "Name", "User", "Port"}
	if verbose {
		headers = append(headers, "CPU%", "MEM", "Cmdline")
	}

	rows := make([][]string, 0, len(procs))
	for _, p := range procs {
		port := ""
		if p.Port > 0 {
			port = fmt.Sprintf("%d", p.Port)
		}
		row := []string{
			fmt.Sprintf("%d", p.PID),
			p.Name,
			p.User,
			port,
		}
		if verbose {
			row = append(row,
				fmt.Sprintf("%.1f", p.CPUPercent),
				formatBytes(p.MemRSS),
				truncate(p.Cmdline, 60),
			)
		}
		rows = append(rows, row)
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cellStyle := lipgloss.NewStyle().PaddingRight(1)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		})

	return t.Render()
}

func formatBytes(b uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1fG", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1fM", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1fK", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
