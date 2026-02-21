package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aiomayo/hdf/internal/config"
	"github.com/aiomayo/hdf/internal/detect"
	"github.com/aiomayo/hdf/internal/finder"
	"github.com/aiomayo/hdf/internal/killer"
	"github.com/aiomayo/hdf/internal/process"
	"github.com/aiomayo/hdf/internal/ui"
	"github.com/aiomayo/hdf/internal/update"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	versionString = "dev"
	rawVersion    = "dev"
)

func SetVersionInfo(version, commit, date string) {
	rawVersion = version
	versionString = version
	if commit != "none" {
		versionString = fmt.Sprintf("%s\n  commit: %s\n  built:  %s", version, commit, date)
	}
}

type exitError struct {
	code    int
	message string
}

func (e *exitError) Error() string { return e.message }

type flags struct {
	port       uint32
	name       string
	pid        int32
	user       string
	force      bool
	all        bool
	yes        bool
	dryRun     bool
	graceful   bool
	timeout    string
	tree       bool
	list       bool
	verbose    bool
	quiet      bool
	interact   bool
	completion string
}

func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	update.StartBackgroundRefresh(rawVersion)

	rootCmd := newRootCmd()
	err := rootCmd.ExecuteContext(ctx)

	if info := update.CheckCached(rawVersion); info != nil {
		log.Warn("A new version of hdf is available",
			"current", info.CurrentVersion,
			"latest", info.LatestVersion)
	}

	if err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			log.Error(ee.message)
			return ee.code
		}
		if ctx.Err() != nil {
			return 130
		}
		log.Error("unexpected error", "err", err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	f := &flags{}

	cmd := &cobra.Command{
		Use:     "hdf [query]",
		Short:   "Kill processes by port, name, PID, or pattern",
		Long:    fmt.Sprintf("hdf — a smart process killer. Pass a port number, process name, PID, or glob pattern.\n\nConfig: %s", config.Path()),
		Version: versionString,
		Args:    cobra.MaximumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			v, _ := cmd.Flags().GetBool("verbose")
			q, _ := cmd.Flags().GetBool("quiet")
			if v {
				log.SetLevel(log.DebugLevel)
			}
			if q {
				log.SetLevel(log.FatalLevel)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.completion != "" {
				return runCompletion(cmd, f.completion)
			}
			if len(args) == 0 && !hasQueryFlags(f) {
				return cmd.Help()
			}
			return run(f, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(newConfigCmd())

	cmd.Flags().Uint32VarP(&f.port, "port", "p", 0, "kill by port number")
	cmd.Flags().StringVarP(&f.name, "name", "n", "", "kill by process name")
	cmd.Flags().Int32Var(&f.pid, "pid", 0, "kill by PID")
	cmd.Flags().StringVarP(&f.user, "user", "u", "", "filter by user")
	cmd.Flags().BoolVarP(&f.force, "force", "f", false, "force kill (SIGKILL)")
	cmd.Flags().BoolVarP(&f.all, "all", "a", false, "kill all matching processes")
	cmd.Flags().BoolVarP(&f.yes, "yes", "y", false, "skip confirmation")
	cmd.Flags().BoolVarP(&f.dryRun, "dry-run", "d", false, "show what would be killed")
	cmd.Flags().BoolVarP(&f.graceful, "graceful", "g", false, "graceful shutdown (SIGTERM then SIGKILL)")
	cmd.Flags().StringVar(&f.timeout, "timeout", "5s", "graceful shutdown timeout")
	cmd.Flags().BoolVarP(&f.tree, "tree", "t", false, "kill process tree")
	cmd.Flags().BoolVarP(&f.list, "list", "l", false, "list matching processes without killing")
	cmd.Flags().BoolVarP(&f.interact, "interactive", "i", false, "interactive process selection")
	cmd.Flags().StringVarP(&f.completion, "completion", "c", "", "generate completion script (bash|zsh|fish|powershell)")

	cmd.PersistentFlags().BoolVarP(&f.verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&f.quiet, "quiet", "q", false, "suppress output")

	return cmd
}

func hasQueryFlags(f *flags) bool {
	return f.port > 0 || f.name != "" || f.pid > 0 || f.user != ""
}

func run(f *flags, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		log.Warn("config load failed", "err", err)
	}

	f.force = f.force || cfg.DefaultForce
	if cfg.DefaultVerbose && !f.verbose {
		f.verbose = true
		log.SetLevel(log.DebugLevel)
	}

	provider := process.New()
	find := finder.New(provider)
	kill := killer.New(provider)

	query := resolveQuery(f, args, cfg)

	var procs []process.Info
	switch {
	case f.user != "" && query == nil:
		procs, err = find.FindByUser(f.user)
	case query != nil:
		q := *query
		procs, err = find.Find(q)
		if f.user != "" {
			procs = filterByUser(procs, f.user)
		}
	default:
		return &exitError{code: 1, message: "no query provided — pass a port, name, PID, or use flags"}
	}
	if err != nil {
		return &exitError{code: 1, message: fmt.Sprintf("find error: %v", err)}
	}

	if len(procs) == 0 {
		log.Info("no matching processes found")
		return nil
	}

	if f.list {
		fmt.Println(ui.RenderTable(procs, f.verbose))
		return nil
	}

	procs = filterProtected(procs, cfg)
	if len(procs) == 0 {
		return &exitError{code: 1, message: "all matching processes are protected"}
	}

	if f.interact || (len(procs) > 1 && !f.all && !f.yes && !f.dryRun) {
		procs, err = ui.PickProcesses(procs)
		if err != nil {
			return &exitError{code: 130, message: "selection cancelled"}
		}
		if len(procs) == 0 {
			return nil
		}
	} else if len(procs) > 1 && !f.all && !f.dryRun {
		fmt.Println(ui.RenderTable(procs, f.verbose))
		return &exitError{code: 1, message: fmt.Sprintf("found %d processes — use -a to kill all, -i for interactive selection", len(procs))}
	}

	if !f.yes && !f.dryRun {
		fmt.Println(ui.RenderTable(procs, f.verbose))
		confirmed, err := ui.Confirm(fmt.Sprintf("Kill %d process(es)?", len(procs)))
		if err != nil || !confirmed {
			return &exitError{code: 130, message: "cancelled"}
		}
	}

	timeout, err := parseTimeout(f.timeout)
	if err != nil {
		return &exitError{code: 1, message: fmt.Sprintf("invalid timeout: %v", err)}
	}

	action := killer.ActionTerminate
	if f.force {
		action = killer.ActionKill
	} else if f.graceful {
		action = killer.ActionGraceful
	}

	opts := killer.Options{
		Action:  action,
		Tree:    f.tree,
		Timeout: timeout,
		DryRun:  f.dryRun,
	}

	results := kill.Execute(procs, opts)

	hasFailure := false
	for _, r := range results {
		if !f.quiet {
			fmt.Println(killer.FormatResult(r))
		}
		if !r.Success {
			hasFailure = true
		}
	}

	if hasFailure {
		return &exitError{code: 1, message: "some processes could not be killed"}
	}
	return nil
}

func resolveQuery(f *flags, args []string, cfg *config.Config) *detect.Query {
	if f.port > 0 {
		q := detect.Query{Type: detect.TypePort, Port: f.port, Raw: fmt.Sprintf("%d", f.port)}
		return &q
	}
	if f.pid > 0 {
		q := detect.Query{Type: detect.TypePID, PID: f.pid, Raw: fmt.Sprintf("%d", f.pid)}
		return &q
	}
	if f.name != "" {
		input := cfg.ResolveAlias(f.name)
		q := detect.Classify(input)
		if q.Type == detect.TypePort || q.Type == detect.TypePID {
			q.Type = detect.TypeName
			q.Name = input
		}
		return &q
	}
	if len(args) > 0 {
		input := cfg.ResolveAlias(args[0])
		q := detect.Classify(input)
		return &q
	}
	return nil
}

func filterByUser(procs []process.Info, user string) []process.Info {
	var result []process.Info
	for _, p := range procs {
		if strings.EqualFold(p.User, user) {
			result = append(result, p)
		}
	}
	return result
}

func filterProtected(procs []process.Info, cfg *config.Config) []process.Info {
	var result []process.Info
	for _, p := range procs {
		if cfg.IsProtected(p.Name) {
			log.Warn("skipping protected process", "name", p.Name, "pid", p.PID)
			continue
		}
		result = append(result, p)
	}
	return result
}

func parseTimeout(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
