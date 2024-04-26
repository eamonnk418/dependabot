package cmd

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eamonnk418/dependabot/internal/client"
	"github.com/spf13/cobra"
)

var supportedEcosystems = map[string][]string{
	"npm":    {"package.json"},
	"gomod":  {"go.mod"},
	"maven":  {"pom.xml"},
	"gradle": {"build.gradle", "build.gradle.kts", "gradle/libs.versions.toml"},
}

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
			ghClient := client.NewGitHubClient(client.WithAuthToken(os.Getenv("GITHUB_TOKEN")))
			tarball, err := ghClient.DownloadRepository(cmd.Context(), repoName)
			if err != nil {
				return fmt.Errorf("failed to download repository: %w", err)
			}
			defer os.Remove(tarball.Name())

			cmd.Printf("Downloaded repository %s\n", repoName)

			b, err := os.ReadFile(tarball.Name())
			if err != nil {
				return fmt.Errorf("failed to read tarball: %w", err)
			}

			gzipReader, err := gzip.NewReader(strings.NewReader(string(b)))
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzipReader.Close()

			tarReader := tar.NewReader(gzipReader)

			var directories []string

			supportedFiles := make(map[string]struct{})
			for _, manifests := range supportedEcosystems {
				for _, manifest := range manifests {
					supportedFiles[manifest] = struct{}{}
				}
			}

			for {
				header, err := tarReader.Next()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}

				for manifest := range supportedFiles {
					if !strings.Contains(header.Name, manifest) {
						continue
					}

					parts := strings.Split(header.Name, "/")
					if len(parts) < 2 {
						continue
					}

					directory := strings.Join(parts[1:], "/")
					if !strings.HasSuffix(directory, manifest) {
						continue
					}

					directory = filepath.Dir(directory)
					if directory == "." {
						directory = "/"
					}

					directories = append(directories, directory)
					break
				}
			}

			for _, directory := range directories {
				cmd.Printf("Found dependency file in %s\n", directory)
			}

			return nil
		},
	}

	previewCmd.Flags().StringVarP(&repoName, "repo-name", "r", "", "Name of the GitHub repository")
	previewCmd.Flags().StringVarP(&path, "filepath", "f", "", "Path to the dependency file")

	return previewCmd
}
