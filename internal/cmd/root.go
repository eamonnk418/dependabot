package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dependabot",
		Short: "dependabot",
		Long:  "dependabot",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return rootCmd
}

func Execute() {
	cobra.CheckErr(NewRootCmd().Execute())
}
