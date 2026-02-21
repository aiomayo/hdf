package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aiomayo/hdf/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage hdf configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newConfigPathCmd(),
		newConfigShowCmd(),
		newConfigEditCmd(),
		newConfigResetCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigAliasCmd(),
		newConfigProtectCmd(),
	)

	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Path())
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print all settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			fmt.Print(formatShow(cfg))
			return nil
		},
	}
}

func formatShow(cfg *config.Config) string {
	type group struct {
		name   string
		fields []config.Field
	}

	seen := map[string]int{}
	var groups []group

	for _, f := range config.Schema {
		key := f.Group
		if key == "" {
			key = "general"
		}
		display := strings.ToUpper(key[:1]) + key[1:]

		idx, ok := seen[key]
		if !ok {
			idx = len(groups)
			seen[key] = idx
			groups = append(groups, group{name: display})
		}
		groups[idx].fields = append(groups[idx].fields, f)
	}

	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "# %s\n", g.name)

		for _, f := range g.fields {
			val, _ := config.GetValue(cfg, f.Key)
			formatted := config.FormatValue(&f, val)

			if f.Kind == config.StringMap {
				fmt.Fprintf(&b, "%s\n", formatted)
			} else {
				fmt.Fprintf(&b, "%-17s = %-10s  # %s\n", f.Key, formatted, f.Desc)
			}
		}
	}

	return b.String()
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config in editor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				editor = "vi"
			}

			if _, err := config.Load(); err != nil {
				return err
			}

			c := exec.Command(editor, config.Path())
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset config to defaults",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Reset(); err != nil {
				return err
			}
			fmt.Println("config reset to defaults")
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "get <key>",
		Short:             "Get a config value",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeConfigKeys,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			f := config.LookupField(key)
			if f == nil {
				return fmt.Errorf("unknown config key: %s", key)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			val, err := config.GetValue(cfg, key)
			if err != nil {
				return err
			}

			fmt.Println(config.FormatValue(f, val))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "set <key> <value>",
		Short:             "Set a config value",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeConfigKeys,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, raw := args[0], args[1]
			f := config.LookupField(key)
			if f == nil {
				return fmt.Errorf("unknown config key: %s", key)
			}

			if f.Kind == config.StringMap || f.Kind == config.StringSlice {
				return fmt.Errorf("%q is a collection type, use a dedicated subcommand to manage it", key)
			}

			val, err := config.ParseValue(f, raw)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %w", key, err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := config.SetValue(cfg, key, val); err != nil {
				return err
			}

			return config.Save(cfg)
		},
	}
}

func newConfigAliasCmd() *cobra.Command {
	var del bool

	cmd := &cobra.Command{
		Use:   "alias <name> [value]",
		Short: "Add or remove an alias",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			name := args[0]

			if del {
				if _, exists := cfg.Aliases[name]; !exists {
					return fmt.Errorf("alias %q not found", name)
				}
				delete(cfg.Aliases, name)
				if err := config.Save(cfg); err != nil {
					return err
				}
				fmt.Printf("alias %q removed\n", name)
				return nil
			}

			if len(args) < 2 {
				return fmt.Errorf("usage: hdf config alias <name> <value>")
			}

			cfg.Aliases[name] = args[1]
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("alias %s = %q\n", name, args[1])
			return nil
		},
	}

	cmd.Flags().BoolVar(&del, "delete", false, "remove an alias")
	return cmd
}

func newConfigProtectCmd() *cobra.Command {
	var del bool

	cmd := &cobra.Command{
		Use:   "protect <name>",
		Short: "Add or remove a protected process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			name := args[0]

			if del {
				idx := -1
				for i, p := range cfg.Protected {
					if strings.EqualFold(p, name) {
						idx = i
						break
					}
				}
				if idx < 0 {
					return fmt.Errorf("process %q not in protected list", name)
				}
				cfg.Protected = append(cfg.Protected[:idx], cfg.Protected[idx+1:]...)
				if err := config.Save(cfg); err != nil {
					return err
				}
				fmt.Printf("removed %q from protected list\n", name)
				return nil
			}

			for _, p := range cfg.Protected {
				if strings.EqualFold(p, name) {
					fmt.Printf("%q is already protected\n", name)
					return nil
				}
			}

			cfg.Protected = append(cfg.Protected, name)
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("added %q to protected list\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&del, "delete", false, "remove from protected list")
	return cmd
}

func completeConfigKeys(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var keys []string
	for _, f := range config.Schema {
		if f.Kind != config.StringMap && f.Kind != config.StringSlice {
			keys = append(keys, f.Key)
		}
	}
	return keys, cobra.ShellCompDirectiveNoFileComp
}
