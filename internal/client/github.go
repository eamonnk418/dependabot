package client

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

type GitHubService interface {
	GetRepository(ctx context.Context, name string) (*github.Repository, error)
	ListRepositories(ctx context.Context, org string) ([]*github.Repository, error)
	ListRepositories2(ctx context.Context, org string) ([]*github.Repository, error)
	GetRepositoryContents(ctx context.Context, filepath string, repository *github.Repository, packageEcosystem EcosystemMap) (string, []string, error)
}

type GitHubClient struct {
	*github.Client
}

type Options func(*GitHubClient) error

func NewGitHubClient(opts ...Options) GitHubService {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Second,
	}

	client := github.NewClient(httpClient)

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
