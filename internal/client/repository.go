package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/google/go-github/v59/github"
)

type EcosystemMap map[string][]string

// Initialize an EcosystemMap with supported ecosystems and their associated filenames
var supportedEcosystems = EcosystemMap{
	"npm":    []string{"package.json"},
	"gomod":  []string{"go.mod"},
	"maven":  []string{"pom.xml"},
	"gradle": []string{"build.gradle", "build.gradle.kts"},
}

func (c GitHubClient) DownloadRepository(ctx context.Context, repoName string) (*os.File, error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, must be owner/repo")
	}

	owner, repo := parts[0], parts[1]

	archiveURL, _, err := c.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, nil, 3)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.BareDo(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pattern := owner + "_" + repo + "_*.tar.gz"
	file, err := os.CreateTemp(os.TempDir(), pattern)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return nil, err
	}

	return file, nil
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

func (c GitHubClient) GetArchiveLink(ctx context.Context, repoName string) (*url.URL, error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("GetArchiveLink, invalid repository format, must be owner/repo")
	}

	owner, repo := parts[0], parts[1]

	tarballURL, _, err := c.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, nil, 3)
	if err != nil {
		return nil, err
	}

	return tarballURL, nil
}

func (c GitHubClient) DownloadTarballArchive(ctx context.Context, archiveURL *url.URL) (*os.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.BareDo(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pattern := c.Repository.GetOwner().GetLogin() + "_" + c.Repository.GetName() + "_*.tar.gz"
	file, err := os.CreateTemp(os.TempDir(), pattern)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		return nil, err
	}

	return file, nil
}

func (c GitHubClient) ExtractTarballArchive(ctx context.Context, file *os.File) ([]string, error) {
	f, err := os.OpenFile(file.Name(), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var files []string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		files = append(files, header.Name)
	}

	return files, nil
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
	var mu sync.Mutex
	var wg sync.WaitGroup

	owner := repository.GetOwner().GetLogin()
	repo := repository.GetName()
	ref := repository.GetDefaultBranch()

	_, directoryContent, _, err := c.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		return "", nil, err
	}

	var packageEcosystem string
	var directories []string

	for _, content := range directoryContent {
		contentType := content.GetType()
		contentName := content.GetName()

		if contentType == "dir" {
			wg.Add(1)
			go func(contentPath string) {
				defer wg.Done()

				subEcosystem, subDirectories, err := c.GetRepositoryContents(ctx, contentPath, repository, supportedEcosystems)
				if err != nil {
					return
				}
				if subEcosystem != "" {
					mu.Lock()
					packageEcosystem = subEcosystem
					directories = append(directories, subDirectories...)
					mu.Unlock()
				}
			}(content.GetPath())
		} else {
			// Check if the file belongs to any supported ecosystem
			ecosystem := supportedEcosystems.GetPackageEcosystem(contentName)
			if ecosystem != "" && packageEcosystem == "" {
				packageEcosystem = ecosystem
				mu.Lock()
				directories = append(directories, strings.TrimSuffix(content.GetPath(), fmt.Sprintf("/%s", contentName)))
				mu.Unlock()
			}
		}
	}

	wg.Wait()

	return packageEcosystem, directories, nil
}

func (c GitHubClient) GetDependabotTemplateData(tarball *os.File) (string, []string, error) {
	b, err := os.ReadFile(tarball.Name())
	if err != nil {
		return "", nil, fmt.Errorf("failed to read tarball: %w", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	var directories []string
	var packageEcosystem string
	var manifest string

	supportedFiles := make(map[string]string) // Map file names to package ecosystems
	for eco, manifests := range supportedEcosystems {
		for _, manifest := range manifests {
			supportedFiles[manifest] = eco
		}
	}

	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", nil, err
		}

		for manifest, ecosystem := range supportedFiles {
			if strings.Contains(header.Name, manifest) {
				packageEcosystem = ecosystem
				break
			}
		}

		if packageEcosystem != "" {
			break
		}
	}

	if packageEcosystem == "" {
		return "", nil, fmt.Errorf("no supported package ecosystem found in the repository")
	}

	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", nil, err
		}

		if strings.Contains(header.Name, manifest) {
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
		}
	}

	return packageEcosystem, directories, nil
}
