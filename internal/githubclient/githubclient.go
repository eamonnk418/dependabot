package githubclient

import (
	"net/http"
	"time"

	"github.com/google/go-github/v59/github"
)

type GitHubClient struct {
	*github.Client
}

func NewGitHubClient() *GitHubClient {
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Second,
	}

	githubClient := github.NewClient(httpClient)

	return &GitHubClient{Client: githubClient}
}
