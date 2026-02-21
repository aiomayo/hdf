package config

import "time"

type Kind int

const (
	Bool Kind = iota
	String
	Select
	Duration
	StringSlice
	StringMap
)

type Field struct {
	Key     string
	Label   string
	Group   string
	Kind    Kind
	Default any
	Options []string
	Desc    string
}

func (f *Field) DisplayName() string {
	if f.Label != "" {
		return f.Label
	}
	return f.Key
}

var Schema = []Field{
	{
		Key:     "graceful_timeout",
		Label:   "Graceful timeout",
		Kind:    Duration,
		Default: 5 * time.Second,
		Desc:    "Graceful shutdown timeout before SIGKILL",
	},
	{
		Key:     "default_force",
		Label:   "Force kill",
		Kind:    Bool,
		Default: false,
		Desc:    "Always use SIGKILL instead of SIGTERM",
	},
	{
		Key:     "default_verbose",
		Label:   "Verbose output",
		Kind:    Bool,
		Default: false,
		Desc:    "Enable verbose output by default",
	},
	{
		Key:     "default_editor",
		Label:   "Config editor",
		Kind:    Select,
		Default: "",
		Options: []string{"", "tui", "vim", "nvim", "nano", "vi"},
		Desc:    "Preferred editor for hdf config edit",
	},
	{
		Key:   "protected",
		Label: "Protected processes",
		Group: "protected",
		Kind:  StringSlice,
		Default: []string{
			"init", "systemd", "launchd", "kernel_task",
			"WindowServer", "loginwindow", "sshd",
		},
		Desc: "Processes that cannot be killed",
	},
	{
		Key:     "aliases",
		Label:   "Aliases",
		Group:   "aliases",
		Kind:    StringMap,
		Default: map[string]string{},
		Desc:    "Query shortcuts (name → target)",
	},
}

func LookupField(key string) *Field {
	for i := range Schema {
		if Schema[i].Key == key {
			return &Schema[i]
		}
	}
	return nil
}
