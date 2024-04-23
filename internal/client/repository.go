package client

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/google/go-github/v59/github"
	"golang.org/x/sync/errgroup"
)

type EcosystemMap map[string][]string

// Initialize an EcosystemMap with supported ecosystems and their associated filenames
var supportedEcosystems = EcosystemMap{
	"npm":    []string{"package.json"},
	"gomod":  []string{"go.mod"},
	"maven":  []string{"pom.xml"},
	"gradle": []string{"build.gradle", "build.gradle.kts"},
}

func (p EcosystemMap) GetPackageEcosystem(fileName string) string {
	for ecosystem, files := range supportedEcosystems {
		for _, file := range files {
			if file == fileName {
				return ecosystem
			}
		}
	}
	return "" // Return empty string if no ecosystem found
}

type Repository struct {
	Name  string
	Owner string
}

func (c GitHubClient) GetRepository(ctx context.Context, repoName string) (*github.Repository, error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, must be owner/repo")
	}

	owner, repo := parts[0], parts[1]

	repository, _, err := c.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func (c GitHubClient) ListRepositories(ctx context.Context, org string) ([]*github.Repository, error) {
	// Fetch the first page to get the total number of pages
	_, resp, err := c.Repositories.ListByOrg(ctx, org, &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1, // Start from the first page
			PerPage: 10,
		},
	})
	if err != nil {
		return nil, err
	}

	// Determine the total number of pages
	totalPages := resp.LastPage

	// Define the number of goroutines to run concurrently
	numWorkers := runtime.GOMAXPROCS(1)

	// Use a channel to communicate the page numbers to process
	pageChan := make(chan int, totalPages)

	// Populate the channel with page numbers
	for i := 0; i < totalPages; i++ {
		pageChan <- i
	}
	close(pageChan) // Close the channel once all pages are sent

	// Create a mutex to safely append repositories to the slice
	var mu sync.Mutex
	var allRepos []*github.Repository

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start goroutines to process pages concurrently
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for page := range pageChan {
				// Fetch repositories for the current page
				repos, _, err := c.Repositories.ListByOrg(ctx, org, &github.RepositoryListByOrgOptions{
					Type: *github.String("public"),
					ListOptions: github.ListOptions{
						Page:    page,
						PerPage: 100,
					},
				})
				if err != nil {
					fmt.Printf("Error fetching page %d: %v\n", page, err)
					continue
				}

				// Lock the mutex before appending repositories to the slice
				mu.Lock()
				allRepos = append(allRepos, repos...)
				mu.Unlock()
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Return the collected repositories and nil error
	return allRepos, nil
}

func (c GitHubClient) ListRepositories2(ctx context.Context, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}

		if resp.NextPage == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c GitHubClient) GetRepositoryContents(ctx context.Context, path string, repository *github.Repository, supportedEcosystems EcosystemMap) (string, []string, error) {
	// Create a map to store directories by ecosystem
	directoriesByEcosystem := make(map[string][]string)
	var mu sync.Mutex // Mutex to synchronize access to directoriesByEcosystem

	// Create an errgroup for concurrent operations
	eg, ctx := errgroup.WithContext(ctx)

	owner := repository.GetOwner().GetLogin()
	repo := repository.GetName()
	ref := repository.GetDefaultBranch()

	// Get the contents of the specified directory
	_, directoryContent, _, err := c.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return "", nil, err
	}

	var packageEcosystem string // To store the package ecosystem found

	for _, content := range directoryContent {
		contentType := content.GetType()
		contentName := content.GetName()

		if contentType == "dir" {
			contentPath := content.GetPath() // Capture the path before goroutine
			// Launch a goroutine to fetch contents of directories concurrently
			eg.Go(func() error {
				subEcosystem, subDirectories, err := c.GetRepositoryContents(ctx, contentPath, repository, supportedEcosystems)
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				if subEcosystem != "" {
					// Associate each directory with the subEcosystem
					for _, dir := range subDirectories {
						directoriesByEcosystem[subEcosystem] = append(directoriesByEcosystem[subEcosystem], dir)
					}
				}
				return nil
			})
		} else if contentName == "package.json" {
			// If it's a package.json file, add the directory containing it to the result
			directoryPath := strings.TrimSuffix(content.GetPath(), "/package.json")
			if directoryPath == "package.json" {
				directoryPath = "/"
			}
			// Get the package ecosystem for this file
			ecosystem := supportedEcosystems.GetPackageEcosystem(contentName)
			if ecosystem != "" && packageEcosystem == "" {
				packageEcosystem = ecosystem // Update packageEcosystem if not already set
				mu.Lock()
				directoriesByEcosystem[ecosystem] = append(directoriesByEcosystem[ecosystem], directoryPath)
				mu.Unlock()
			}
		}
	}

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		return "", nil, err
	}

	// Flatten the directoriesByEcosystem map into a slice of directories
	var directories []string
	for _, dirs := range directoriesByEcosystem {
		directories = append(directories, dirs...)
	}

	return packageEcosystem, directories, nil
}
