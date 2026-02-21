package killer

import (
	"fmt"
	"slices"
	"time"

	"github.com/aiomayo/hdf/internal/process"
)

type Options struct {
	Action  Action
	Tree    bool
	Timeout time.Duration
	DryRun  bool
}

type Result struct {
	PID     int32
	Name    string
	Success bool
	Error   error
	DryRun  bool
}

type Killer struct {
	provider process.Provider
}

func New(provider process.Provider) *Killer {
	return &Killer{provider: provider}
}

func (k *Killer) Execute(targets []process.Info, opts Options) []Result {
	if opts.Tree {
		targets = k.expandTree(targets)
	}

	var results []Result
	for _, target := range targets {
		r := k.killOne(target, opts)
		results = append(results, r)
	}
	return results
}

func (k *Killer) killOne(target process.Info, opts Options) Result {
	r := Result{
		PID:    target.PID,
		Name:   target.Name,
		DryRun: opts.DryRun,
	}

	if opts.DryRun {
		r.Success = true
		return r
	}

	var err error
	switch opts.Action {
	case ActionKill:
		err = k.provider.Kill(target.PID)
	case ActionTerminate:
		err = k.provider.Terminate(target.PID)
	case ActionGraceful:
		err = k.graceful(target.PID, opts.Timeout)
	}

	if err != nil {
		r.Error = err
		return r
	}
	r.Success = true
	return r
}

func (k *Killer) graceful(pid int32, timeout time.Duration) error {
	if err := k.provider.Terminate(pid); err != nil {
		return err
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return k.provider.Kill(pid)
		case <-ticker.C:
			if !k.provider.IsRunning(pid) {
				return nil
			}
		}
	}
}

func (k *Killer) expandTree(targets []process.Info) []process.Info {
	seen := make(map[int32]bool)
	var expanded []process.Info

	for _, target := range targets {
		k.collectTree(target.PID, &expanded, seen)
	}

	slices.Reverse(expanded)
	return expanded
}

func (k *Killer) collectTree(pid int32, result *[]process.Info, seen map[int32]bool) {
	if seen[pid] {
		return
	}
	seen[pid] = true

	info, err := k.provider.FindByPID(pid)
	if err != nil {
		return
	}

	children, _ := k.provider.Children(pid)
	for _, child := range children {
		k.collectTree(child.PID, result, seen)
	}

	*result = append(*result, *info)
}

func FormatResult(r Result) string {
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would kill %s (PID %d)", r.Name, r.PID)
	}
	if r.Success {
		return fmt.Sprintf("killed %s (PID %d)", r.Name, r.PID)
	}
	return fmt.Sprintf("failed to kill %s (PID %d): %v", r.Name, r.PID, r.Error)
}
