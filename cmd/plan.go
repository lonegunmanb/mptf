package cmd

import (
	"github.com/spf13/cobra"
)

func NewPlanCmd() *cobra.Command {
	recursive := false
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generates a plan based on the specified configuration",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: wrapTerraformCommandWithEphemeralTransform(cf.tfDir, "plan", &recursive),
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "With transforms to all modules or not, default to the root module only.")
	return cmd
}

func init() {
	rootCmd.AddCommand(NewPlanCmd())
}
