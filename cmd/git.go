package main

import (
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func cloneRepo(repo string) (*git.Repository, error) {
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

func newBranch(r *git.Repository, b string) error {
	branch := plumbing.ReferenceName("refs/heads/" + b)

	w, err := r.Worktree()
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

func commit(r *git.Repository, f, id string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(f)
	if err != nil {
		return err
	}
	_, err = w.Commit(id, &git.CommitOptions{
		Author: &object.Signature{
			Name:  id,
			Email: "grace-staff@gsa.gov",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
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
