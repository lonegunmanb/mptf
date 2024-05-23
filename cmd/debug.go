package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Azure/golden"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/lonegunmanb/mptf/pkg"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

func NewDebugCmd() *cobra.Command {
	var tfDir, mptfDir string
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Start REPL mode, mptf debug [path to config files]",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: replFunc(&tfDir, &mptfDir),
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("error on getting working dir:%s", err.Error()))
	}
	debugCmd.Flags().StringVar(&tfDir, "tf-dir", pwd, "Terraform directory")
	debugCmd.Flags().StringVar(&mptfDir, "mptf-dir", "", "MPTF directory, you can assign only one mptf-dir for debug command")
	err = debugCmd.MarkFlagRequired("mptf-dir")
	if err != nil {
		panic(err)
	}
	return debugCmd
}

func replFunc(tfDir, mptfDir *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		varFlags, err := varFlags(os.Args)
		if err != nil {
			return err
		}
		hclBlocks, err := pkg.LoadMPTFHclBlocks(false, *mptfDir)
		if err != nil {
			return err
		}
		cfg, err := pkg.NewMetaProgrammingTFConfig(*tfDir, hclBlocks, varFlags, c.Context())
		if err != nil {
			return err
		}
		_, err = pkg.RunMetaProgrammingTFPlan(cfg)
		if err != nil {
			return err
		}
		line := liner.NewLiner()
		defer func() {
			_ = line.Close()
		}()

		line.SetCtrlCAborts(true)
		fmt.Println("Entering debuging mode, press `quit` or `exit` or Ctrl+C to quit.")

		for {
			if input, err := line.Prompt("debug> "); err == nil {
				if input == "quit" || input == "exit" {
					return nil
				}
				line.AppendHistory(input)
				expression, diag := hclsyntax.ParseExpression([]byte(input), "repl.hcl", hcl.InitialPos)
				if diag.HasErrors() {
					fmt.Printf("%s\n", diag.Error())
					continue
				}
				value, diag := expression.Value(cfg.EvalContext())
				if diag.HasErrors() {
					fmt.Printf("%s\n", diag.Error())
					continue
				}
				fmt.Println(golden.CtyValueToString(value))
			} else if errors.Is(err, liner.ErrPromptAborted) {
				fmt.Println("Aborted")
				break
			} else {
				fmt.Println("Error reading line: ", err)
				break
			}
		}

		return nil
	}
}

func init() {
	rootCmd.AddCommand(NewDebugCmd())
}
