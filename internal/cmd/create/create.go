package create

import "github.com/spf13/cobra"

func NewCreateCmd() *cobra.Command{
	createCmd := &cobra.Command{
		Use: "create",
		Short: "create",
		Long: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Hello from create command")
			return nil
		},
	}

	return createCmd
}