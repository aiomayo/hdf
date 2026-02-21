package config

import "time"

type Kind int

const (
	Bool Kind = iota
	String
	Duration
	StringSlice
	StringMap
)

type Field struct {
	Key     string
	Group   string
	Kind    Kind
	Default any
	Desc    string
}

var Schema = []Field{
	{
		Key:     "graceful_timeout",
		Kind:    Duration,
		Default: 5 * time.Second,
		Desc:    "Graceful shutdown timeout before SIGKILL",
	},
	{
		Key:     "default_force",
		Kind:    Bool,
		Default: false,
		Desc:    "Always use SIGKILL",
	},
	{
		Key:     "default_verbose",
		Kind:    Bool,
		Default: false,
		Desc:    "Enable verbose output by default",
	},
	{
		Key:   "protected",
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
		Group:   "aliases",
		Kind:    StringMap,
		Default: map[string]string{},
		Desc:    "Query shortcuts",
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
