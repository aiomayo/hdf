package ui

import "github.com/charmbracelet/huh"

func Confirm(message string) (bool, error) {
	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Affirmative("Kill").
				Negative("Cancel").
				Value(&confirmed),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}
	return confirmed, nil
}
