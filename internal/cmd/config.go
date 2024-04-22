package cmd

import "github.com/spf13/cobra"

func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "config",
		Long:  "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Hello from config command")
			return nil

		},
	}

	configCmd.AddCommand(NewCreateCmd())
	configCmd.AddCommand(NewPreviewCmd())

	return configCmd
}
