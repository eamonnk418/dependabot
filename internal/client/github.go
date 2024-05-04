package client

import (
	"context"
	"net/url"
	"os"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

type GitHubService interface {
	GetRepository(ctx context.Context, name string) (*github.Repository, error)
	ListRepositories(ctx context.Context, org string) ([]*github.Repository, error)
	ListRepositories2(ctx context.Context, org string) ([]*github.Repository, error)
	GetRepositoryContents(ctx context.Context, filepath string, repository *github.Repository, packageEcosystem EcosystemMap) (string, []string, error)
	GetArchiveLink(ctx context.Context, repoName string) (*url.URL, error)
	DownloadTarballArchive(ctx context.Context, archiveURL *url.URL) (*os.File, error)
	ExtractTarballArchive(ctx context.Context, file *os.File) ([]string, error)
	DownloadRepository(ctx context.Context, repoName string) (*os.File, error)
	GetDependabotTemplateData(tarball *os.File) (string, []string, error)
}

type GitHubClient struct {
	*github.Client
	*github.Repository
}

type Options func(*GitHubClient) error

func NewGitHubClient(opts ...Options) GitHubService {
	client := github.NewClient(nil)

	gitHubClient := &GitHubClient{Client: client}

	for _, opt := range opts {
		opt(gitHubClient)
	}

	return gitHubClient
}

func WithAuthToken(token string) Options {
	return func(c *GitHubClient) error {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		c.Client = github.NewClient(tc)
		return nil
	}
}
