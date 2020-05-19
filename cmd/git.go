package main

import (
	"fmt"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func cloneRepo(repo string) (*git.Repository, error) {
	fmt.Printf("Cloning repository: %s", repo)
	url := "https://github.com/GSA/" + repo
	directory := "/tmp/" + repo
	token := os.Getenv("GITHUB_TOKEN")

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth: &http.BasicAuth{
			Username: "access_token", // yes, this can be anything except an empty string
			Password: token,
		},
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return r, err
	}

	return r, nil
}

func (r *req) newBranch() error {
	fmt.Printf("Adding branch: %s", r.repoName)
	branch := plumbing.ReferenceName("refs/heads/" + r.repoName)

	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Force:  false,
		Branch: branch,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *req) commit() error {
	fmt.Println("Committing changes")
	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(r.relPath)
	if err != nil {
		return err
	}
	_, err = w.Commit(r.ritm.Number, &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.ritm.Number,
			Email: r.email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	fmt.Println("Pushing changes to GitHub")
	err = r.repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "access_token", // yes, this can be anything except an empty string
			Password: os.Getenv("GITHUB_TOKEN"),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
