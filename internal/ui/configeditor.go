package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aiomayo/hdf/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	accentColor = lipgloss.Color("12")
	dimColor    = lipgloss.Color("240")
	errorColor  = lipgloss.Color("9")

	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(accentColor).MarginBottom(1)
	cursorStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	keyStyle    = lipgloss.NewStyle().Width(30)
	valueStyle  = lipgloss.NewStyle().Foreground(dimColor)
	activeValue = lipgloss.NewStyle().Foreground(accentColor)
	descStyle   = lipgloss.NewStyle().Foreground(dimColor).Italic(true)
	footerStyle = lipgloss.NewStyle().Foreground(dimColor).MarginTop(1)
	errorStyle  = lipgloss.NewStyle().Foreground(errorColor)
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimColor).
			Padding(0, 1)
)

type configEditor struct {
	cfg       *config.Config
	fields    []*config.Field
	filtered  []int
	cursor    int
	search    textinput.Model
	searching bool
	editInput textinput.Model
	editing   bool
	editErr   string
	dirty     bool
	saved     bool
	quitting  bool
	width     int
}

func EditConfig(cfg *config.Config) (bool, error) {
	m := newConfigEditor(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return false, err
	}
	final := result.(configEditor)
	return final.saved && final.dirty, nil
}

func newConfigEditor(cfg *config.Config) configEditor {
	search := textinput.New()
	search.Placeholder = "Search settings..."
	search.Prompt = "⌕ "
	search.CharLimit = 64
	search.Width = 74

	edit := textinput.New()
	edit.CharLimit = 256

	var fields []*config.Field
	for i := range config.Schema {
		fields = append(fields, &config.Schema[i])
	}

	m := configEditor{
		cfg:       cfg,
		fields:    fields,
		search:    search,
		editInput: edit,
		width:     80,
	}
	m.applyFilter()
	return m
}

func (m configEditor) Init() tea.Cmd {
	return nil
}

func (m configEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.search.Width = max(20, m.width-8)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.editing {
			return m.handleEditKey(msg)
		}
		if m.searching {
			return m.handleSearchKey(msg)
		}
		return m.handleListKey(msg)
	}

	return m, nil
}

func (m configEditor) currentField() *config.Field {
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return m.fields[m.filtered[m.cursor]]
}

func (m configEditor) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.saved = true
		m.quitting = true
		return m, tea.Quit

	case "esc", "q":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		} else {
			m.searching = true
			m.search.Focus()
			return m, textinput.Blink
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case "/":
		m.searching = true
		m.search.Focus()
		return m, textinput.Blink

	case " ":
		f := m.currentField()
		if f == nil {
			break
		}
		switch f.Kind {
		case config.Bool:
			val, _ := config.GetValue(m.cfg, f.Key)
			_ = config.SetValue(m.cfg, f.Key, !val.(bool))
			m.dirty = true
		case config.Select:
			val, _ := config.GetValue(m.cfg, f.Key)
			current := val.(string)
			next := f.Options[0]
			for i, opt := range f.Options {
				if opt == current && i+1 < len(f.Options) {
					next = f.Options[i+1]
					break
				}
			}
			_ = config.SetValue(m.cfg, f.Key, next)
			m.dirty = true
		default:
			m.editing = true
			m.editErr = ""
			m.editInput.SetValue(formatEditable(m.cfg, f))
			m.editInput.Focus()
			m.editInput.CursorEnd()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m configEditor) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
		m.search.SetValue("")
		m.search.Blur()
		m.applyFilter()
		return m, nil
	case "enter", "down":
		m.searching = false
		m.search.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.applyFilter()
	return m, cmd
}

func (m configEditor) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.editErr = ""
		m.editInput.Blur()
		return m, nil
	case "enter":
		f := m.currentField()
		raw := m.editInput.Value()
		if err := parseAndSet(m.cfg, f, raw); err != nil {
			m.editErr = err.Error()
			return m, nil
		}
		m.dirty = true
		m.editing = false
		m.editErr = ""
		m.editInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m *configEditor) applyFilter() {
	query := strings.ToLower(m.search.Value())
	m.filtered = m.filtered[:0]
	for i, f := range m.fields {
		if query == "" || strings.Contains(strings.ToLower(f.Key), query) || strings.Contains(strings.ToLower(f.Label), query) || strings.Contains(strings.ToLower(f.Desc), query) {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m configEditor) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Configure hdf preferences"))
	b.WriteString("\n\n")

	boxWidth := max(24, m.width-4)
	searchBox := borderStyle.Width(boxWidth).Render(m.search.View())
	b.WriteString(searchBox)
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(descStyle.Render("  No matching settings"))
		b.WriteString("\n")
	}

	active := !m.searching
	for i, idx := range m.filtered {
		f := m.fields[idx]
		isCursor := active && i == m.cursor

		val, _ := config.GetValue(m.cfg, f.Key)
		formatted := config.FormatValue(f, val)

		prefix := "  "
		if isCursor {
			prefix = cursorStyle.Render("❯ ")
		}

		key := keyStyle.Render(f.DisplayName())

		var valStr string
		switch {
		case m.editing && isCursor:
			valStr = m.editInput.View()
		case isCursor:
			valStr = activeValue.Render(formatted)
		default:
			valStr = valueStyle.Render(formatted)
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", prefix, key, valStr))

		if isCursor && !m.editing {
			b.WriteString(fmt.Sprintf("    %s\n", descStyle.Render(f.Desc)))
		}
		if isCursor && m.editErr != "" {
			b.WriteString(fmt.Sprintf("    %s\n", errorStyle.Render(m.editErr)))
		}
	}

	switch {
	case m.editing:
		b.WriteString(footerStyle.Render("Enter to confirm · Esc to cancel"))
	case m.searching:
		b.WriteString(footerStyle.Render("Type to filter · ↓ to list · Esc to clear"))
	default:
		b.WriteString(footerStyle.Render("Space to change · / to search · Enter to save · Esc to discard"))
	}

	return b.String()
}

func formatEditable(cfg *config.Config, f *config.Field) string {
	val, _ := config.GetValue(cfg, f.Key)
	switch f.Kind {
	case config.String:
		return val.(string)
	case config.Duration:
		return val.(time.Duration).String()
	case config.StringSlice:
		return strings.Join(val.([]string), ", ")
	case config.StringMap:
		m := val.(map[string]string)
		names := make([]string, 0, len(m))
		for k := range m {
			names = append(names, k)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(m))
		for _, k := range names {
			parts = append(parts, k+"="+m[k])
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", val)
	}
}

func parseAndSet(cfg *config.Config, f *config.Field, raw string) error {
	switch f.Kind {
	case config.StringSlice:
		parts := strings.Split(raw, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return config.SetValue(cfg, f.Key, result)
	case config.StringMap:
		m := make(map[string]string)
		if strings.TrimSpace(raw) != "" {
			parts := strings.Split(raw, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				kv := strings.SplitN(p, "=", 2)
				if len(kv) != 2 {
					return fmt.Errorf("invalid format %q, use name=value", p)
				}
				m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		return config.SetValue(cfg, f.Key, m)
	default:
		val, err := config.ParseValue(f, raw)
		if err != nil {
			return err
		}
		return config.SetValue(cfg, f.Key, val)
	}
}
