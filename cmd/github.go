package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

func newAuthenticatedClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func (r *req) pullRequest() (*github.PullRequest, error) {
	fmt.Println("Creating Pull request")
	ctx := context.Background()
	baseBranch := "master"
	commitBranch := r.ritm.Number
	owner := "GSA"
	serviceNowURL := fmt.Sprintf("https://%s/nav_to.do?uri=sc_req_item.do%%3Fsys_id%%3D", os.Getenv("SN_INSTANCE"))
	prBody := fmt.Sprintf("[%s](%s%s)\n- %s %s RDS in %s account",
		r.ritm.Number, serviceNowURL, r.ritm.SysID, r.ritm.Size, r.ritm.Engine, r.ritm.Account)
	newPR := &github.NewPullRequest{
		Title: &r.ritm.Number,
		Head:  &commitBranch,
		Base:  &baseBranch,
		Body:  &prBody,
	}

	pr, _, err := r.githubClient.PullRequests.Create(ctx, owner, r.repoName, newPR)
	if err != nil {
		return pr, err
	}

	revReq := github.ReviewersRequest{
		TeamReviewers: []string{"grace-developers"},
	}

	_, _, err = r.githubClient.PullRequests.RequestReviewers(ctx, owner, r.repoName, *pr.Number, revReq)
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
