package cmd

import (
	"os"

	"github.com/eamonnk418/dependabot/internal/client"
	"github.com/spf13/cobra"
)

func NewPreviewCmd() *cobra.Command {
	var (
		repoName string
		path     string
	)

	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "preview",
		Long:  "preview",
		RunE: func(cmd *cobra.Command, args []string) error {
			// analyze github repo what eco-system it uses etc.

			token := os.Getenv("GITHUB_TOKEN")

			githubClient := client.NewGitHubClient(client.WithAuthToken(token))

			// repos, err := githubClient.ListRepositories(cmd.Context(), org)
			// if err != nil {
			// 	return err
			// }

			// for i, r := range repos {
			// 	cmd.Printf("Name[%d]: %s, Visibility: %s\n", i+1, r.GetFullName(), r.GetVisibility())
			// }

			// repos2, err := githubClient.ListRepositories2(cmd.Context(), repoName)
			// if err != nil {
			// 	return err
			// }

			// for i, r := range repos2 {
			// 	cmd.Printf("Name[%d]: %s, Visibility: %s\n", i+1, r.GetFullName(), r.GetVisibility())
			// }

			repo, err := githubClient.GetRepository(cmd.Context(), repoName)
			if err != nil {
				return err
			}

			ecosystem, err := githubClient.GetRepositoryContents(cmd.Context(), path, repo)
			if err != nil {
				return err
			}

			for _, path := range ecosystem {
				cmd.Printf("Dictectory: %s\n", path)
			}

			// generate the dependabot.yml template

			// output the preview of the dependabot config file

			return nil
		},
	}

	previewCmd.Flags().StringVarP(&repoName, "repo-name", "r", "", "Name of the GitHub repository")
	previewCmd.Flags().StringVarP(&path, "filepath", "f", "", "Path to the dependency file")

	return previewCmd
}
