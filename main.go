package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func main() {
	var destination string
	flag.StringVar(&destination, "destination", "./repos", "The destination directory for cloned repositories")
	flag.Parse()

	if _, err := os.Stat(destination); os.IsNotExist(err) {
		if err := os.Mkdir(destination, 0755); err != nil {
			fmt.Printf("Error creating destination directory: %v\n", err)
			os.Exit(1)
		}
	}

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
	opt := &github.RepositoryListOptions{Type: "public", ListOptions: github.ListOptions{PerPage: 100}}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List(context.Background(), "", opt)
		if err != nil {
			fmt.Printf("💀 error listing repositories: %v\n", err)
			os.Exit(1)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		fmt.Printf("💀 error getting user info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cloning all repos for user: github.com/%s\n", *user.Login)
	fmt.Printf("Found: %d repos\n", len(allRepos))

	// for each repo we find, clone or pull it
	for i, repo := range allRepos {
		cloneURL := *repo.CloneURL
		repoName := *repo.Name
		destPath := filepath.Join(destination, repoName)

		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			// announce each repo as we clone it
			fmt.Printf("[ %0*d / %d ] ✅ 📦 cloned: %s ", len(fmt.Sprint(len(allRepos))), i+1, len(allRepos), repoName)

			cmd := exec.Command("git", "clone", cloneURL, destPath)
			var output bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &output
			err := cmd.Run()
			if err != nil {
				fmt.Printf("❌\n")
				fmt.Printf("Error cloning repository %s: %v\n", repoName, output.String())
				continue
			}
			fmt.Printf("\n")
		} else {
			// announce each repo as we pull it
			fmt.Printf("[ %0*d / %d ] ✅ ⬇️  pulled: %s ", len(fmt.Sprint(len(allRepos))), i+1, len(allRepos), repoName)

			cmd := exec.Command("git", "-C", destPath, "pull", "--rebase")
			var output bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &output
			err := cmd.Run()
			if err != nil {
				fmt.Printf("❌\n")
				fmt.Printf("Error pulling repository %s: %v\n", repoName, output.String())
				continue
			}
			fmt.Printf("\n")
		}
	}
}
