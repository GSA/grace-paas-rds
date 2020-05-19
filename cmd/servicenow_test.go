package main

import (
	"fmt"
	"os"
	"testing"
)

func TestUpdateRITM(t *testing.T) {
	if os.Getenv("SN_INSTANCE") == "" {
		t.Skip("skipping test, SN_INSTANCE not set")
	}
	if os.Getenv("SN_PASSWORD") == "" {
		t.Skip("skipping test, SN_PASSWORD not set")
	}
	if os.Getenv("SN_USER") == "" {
		t.Skip("skipping test, SN_USER not set")
	}

	var r *req
	r.ritm = &ritm{SysID: "9744dd541bb450505fdaa82fe54bcb2b"}
	err := r.updateRITM(fmt.Errorf("%s", "testing"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
