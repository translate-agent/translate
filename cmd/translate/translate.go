package translate

import (
	"github.com/spf13/cobra"
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "This command sets configuration",
	Long: `translate command sets configuration from:
			environment variables
			yaml file
			CLI arguments`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {}
