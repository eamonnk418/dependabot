package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/eamonnk418/dependabot/internal/client"
	"github.com/eamonnk418/dependabot/internal/dependabot"
	"github.com/eamonnk418/dependabot/internal/schema"
	"github.com/spf13/cobra"
)

func NewCmdConfigPreview() *cobra.Command {
	var configPreviewOptions struct {
		repoName string
	}

	cmd := &cobra.Command{
		Use:   "config-preview",
		Short: "Preview the dependabot configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			tarball, err := downloadRepository(configPreviewOptions.repoName)
			if err != nil {
				return fmt.Errorf("failed to download repository: %w", err)
			}
			defer os.Remove(tarball.Name())

			return analyzeRepository(tarball)
		},
	}

	cmd.Flags().StringVarP(&configPreviewOptions.repoName, "repo", "r", "", "The repository to generate the configuration for")

	return cmd
}

func downloadRepository(repoName string) (*os.File, error) {
	ghc := client.NewGitHubClient(client.WithAuthToken(os.Getenv("GITHUB_TOKEN")))
	return ghc.DownloadRepository(context.Background(), repoName)
}

func analyzeRepository(tarball *os.File) error {
	ghc := client.NewGitHubClient(client.WithAuthToken(os.Getenv("GITHUB_TOKEN")))
	_, directories, err := ghc.GetDependabotTemplateData(tarball)
	if err != nil {
		return err
	}

	factory := dependabot.Factory{
		PackageEcosystem: "none",
		Directories:      directories,
	}

	template := dependabot.NewDependabotTemplateFactory(factory)

	dependabotYml, err := template.GenerateTemplate(&schema.Dependabot{
		Version: 2,
		Updates: []*schema.Update{
			{
				PackageEcosystem: "none",
				Directory:        "/",
				Schedule: &schema.Schedule{
					Interval: "weekly",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	fmt.Println(dependabotYml)

	return nil
}
