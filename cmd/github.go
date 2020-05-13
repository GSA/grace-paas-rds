package main

import (
	"context"
	"fmt"
	"os"

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

func pullRequest(r *ritm, repo string) error {
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
		return err
	}

	req := github.ReviewersRequest{
		TeamReviewers: []string{"grace-developers"},
	}

	_, _, err = client.PullRequests.RequestReviewers(ctx, owner, repo, *pr.Number, req)
	if err != nil {
		return err
	}

	return nil
}
