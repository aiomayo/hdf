package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func runCompletion(cmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return fmt.Errorf("invalid shell %q: must be one of: bash, zsh, fish, powershell", shell)
	}
}
