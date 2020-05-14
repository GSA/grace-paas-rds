package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

const serviceNowURL = "https://gsasandbox.servicenowservices.com/nav_to.do?uri=sc_req_item.do%3Fsys_id%3D"

func newAuthenticatedClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func pullRequest(r *ritm, repo string) (*github.PullRequest, error) {
	fmt.Println("Creating Pull request")
	ctx := context.Background()
	baseBranch := "master"
	commitBranch := r.Number
	owner := "GSA"
	prBody := fmt.Sprintf("[%s](%s%s)\n- %s %s RDS in %s account", r.Number, serviceNowURL, r.SysID, r.Size, r.Engine, r.Account)
	client := newAuthenticatedClient()
	newPR := &github.NewPullRequest{
		Title: &r.Number,
		Head:  &commitBranch,
		Base:  &baseBranch,
		Body:  &prBody,
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return pr, err
	}

	req := github.ReviewersRequest{
		TeamReviewers: []string{"grace-developers"},
	}

	_, _, err = client.PullRequests.RequestReviewers(ctx, owner, repo, *pr.Number, req)
	if err != nil {
		return pr, err
	}

	return pr, nil
}

func waitForMerge(pr *github.PullRequest) error {
	client := newAuthenticatedClient()
	ctx := context.Background()
	owner := *pr.Base.Repo.Owner.Login
	repo := *pr.Base.Repo.Name
	var err error

	fmt.Print("Waiting for Pull Request to be merged")
	for *pr.State != "closed" {
		fmt.Print(".")
		time.Sleep(10 * time.Second)
		pr, _, err = client.PullRequests.Get(ctx, owner, repo, *pr.Number)
		if err != nil {
			return err
		}
	}
	fmt.Println()

	if !*pr.Merged {
		return fmt.Errorf("pull request %s but not merged", *pr.State)
	}

	return nil
}
