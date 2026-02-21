package ui

import (
	"fmt"

	"github.com/aiomayo/hdf/internal/process"
	"github.com/charmbracelet/huh"
)

func PickProcesses(procs []process.Info) ([]process.Info, error) {
	options := make([]huh.Option[int], 0, len(procs))
	for i, p := range procs {
		label := fmt.Sprintf("%-8d %-20s %-15s", p.PID, p.Name, p.User)
		if p.Port > 0 {
			label += fmt.Sprintf(" :%d", p.Port)
		}
		options = append(options, huh.NewOption(label, i))
	}

	var selected []int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select processes to kill").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	result := make([]process.Info, 0, len(selected))
	for _, idx := range selected {
		result = append(result, procs[idx])
	}
	return result, nil
}
