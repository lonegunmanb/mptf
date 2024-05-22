package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/golden"
	"github.com/lonegunmanb/mptf/pkg"
	"github.com/lonegunmanb/mptf/pkg/backup"
	"github.com/spf13/cobra"
	"os"
)

func NewTransformCmd() *cobra.Command {
	recursive := false

	applyCmd := &cobra.Command{
		Use:   "transform",
		Short: "Apply the transforms, mptf transform [-r] --tf-dir --mptf-dir, support mutilple mptf dirs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			varFlags, err := varFlags(os.Args)
			if err != nil {
				return err
			}
			tfDirs := []string{cf.tfDir}
			if recursive {
				modulePaths, err := pkg.ModulePaths(tfDirs[0])
				if err != nil {
					return err
				}
				tfDirs = modulePaths
			}
			for _, tfDir := range tfDirs {
				d := tfDir
				err = backup.BackupFolder(d)
				if err != nil {
					return err
				}
			}
			var mptfDirs []string
			for _, dir := range cf.mptfDirs {
				localizedDir, dispose, err := localizeConfigFolder(dir, cmd.Context())
				if err != nil {
					return err
				}
				if dispose != nil {
					defer dispose()
				}
				mptfDirs = append(mptfDirs, localizedDir)
			}
			for _, mptfDir := range mptfDirs {
				hclBlocks, err := pkg.LoadMPTFHclBlocks(false, mptfDir)
				if err != nil {
					return err
				}
				for _, tfDir := range tfDirs {
					err = applyTransform(cmd.Context(), tfDir, hclBlocks, varFlags)
					if err != nil {
						return err
					}
				}
			}
			fmt.Println("Plan applied successfully.")
			return nil
		},
	}

	applyCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Apply transforms to all modules or not, default to the root module only.")
	return applyCmd
}

func applyTransform(ctx context.Context, tfDir string, hclBlocks []*golden.HclBlock, varFlags []golden.CliFlagAssignedVariables) error {

	cfg, err := pkg.NewMetaProgrammingTFConfig(tfDir, hclBlocks, varFlags, ctx)
	if err != nil {
		return err
	}
	plan, err := pkg.RunMetaProgrammingTFPlan(cfg)
	if err != nil {
		return err
	}
	if len(plan.Transforms) == 0 {
		fmt.Println("No transforms to apply.")
		return nil
	}
	fmt.Println(plan.String())
	err = plan.Apply()
	if err != nil {
		return fmt.Errorf("error applying plan: %s\n", err.Error())
	}
	return nil
}

func init() {
	rootCmd.AddCommand(NewTransformCmd())
}