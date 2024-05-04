package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dependabot-cli",
		Short: "A CLI tool to generate dependabot configuration files",
	}

	cmd.AddCommand(NewCmdConfigPreview())

	return cmd
}

func Execute() {
	cobra.CheckErr(NewCmdRoot().Execute())
}
