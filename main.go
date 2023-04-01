package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN environment variable not set.")
		os.Exit(1)
	}

	// make a new client with the token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	// list all repositories, public in this case
	opt := &github.RepositoryListOptions{Type: "public"}
	repos, _, err := client.Repositories.List(context.Background(), "", opt)
	if err != nil {
		fmt.Printf("💀 error listing repositories: %v\n", err)
		os.Exit(1)
	}

	// for each repo we find, clone it
	for _, repo := range repos {
		cloneURL := *repo.CloneURL
		repoName := *repo.Name
		destPath := filepath.Join(".", repoName)

		// announce each repo as we clone it
		fmt.Printf("👷 cloning %s...\n", cloneURL)

		cmd := exec.Command("git", "clone", cloneURL, destPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("💀 error cloning repository %s: %v\n", repoName, err)
			continue
		}
	}
}
