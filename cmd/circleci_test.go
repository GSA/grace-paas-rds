package main

import (
	"context"
	"os"
	"testing"
)

func TestWaitForApply(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("skipping test, GITHUB_TOKEN not set")
	}

	if os.Getenv("CIRCLE_TOKEN") == "" {
		t.Skip("skipping test, CIRCLE_TOKEN not set")
	}

	client := newAuthenticatedClient()
	ctx := context.Background()
	pr, _, err := client.PullRequests.Get(ctx, "GSA", "grace-paas-baseline", 21)
	if err != nil {
		t.Errorf("waitForMerge() failed. Unexpected error getting PR: %v\n", err)
	}
	err = waitForApply(pr)

	t.Logf("Error: %v", err)
}
