package main

import (
	"context"
	"os"
	"testing"
)

func TestWaitForMerge(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("skipping test, GITHUB_TOKEN not set")
	}
	client := newAuthenticatedClient()

	ctx := context.Background()

	pr, _, err := client.PullRequests.Get(ctx, "GSA", "grace-paas-baseline", 17)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error getting PR: %v\n", err)
	}

	err = waitForMerge(pr)
	if err == nil {
		t.Errorf("waitForMerge() failed. Expected error but got none.")
	} else if err.Error() != "pull request closed but not merged" {
		t.Errorf("unexpected error. Expected %q. Got %q", "pull request closed but not merged", err.Error())
	}

	pr, _, err = client.PullRequests.Get(ctx, "GSA", "grace-paas-baseline", 20)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error getting PR: %v\n", err)
	}

	err = waitForMerge(pr)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error: %v", err)
	}

	pr, _, err = client.PullRequests.Get(ctx, "GSA", "grace-paas-baseline", 21)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error getting PR: %v\n", err)
	}

	err = waitForMerge(pr)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error: %v", err)
	}
}
