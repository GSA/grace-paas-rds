package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func resetEnv(oldArgs []string, oldEnv map[string]string) {
	os.Args = oldArgs
	for k, v := range oldEnv {
		os.Setenv(k, v)
	}
}

func captureEnv() (oldArgs []string, oldEnv map[string]string) {
	oldArgs = os.Args
	oldEnv = map[string]string{
		"CIRCLE_TOKEN": os.Getenv("CIRCLE_TOKEN"),
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
		"SN_INSTANCE":  os.Getenv("SN_INSTANCE"),
		"SN_PASSWORD":  os.Getenv("SN_PASSWORD"),
		"SN_USER":      os.Getenv("SN_USER"),
	}
	return oldArgs, oldEnv
}

// nolint: funlen
func TestNewReq(t *testing.T) {
	oldArgs, oldEnv := captureEnv()
	tt := map[string]struct {
		args []string
		env  map[string]string
		err  string
		req  *req
	}{
		"happy": {
			args: []string{"cmd1", "-request", filepath.Join("testdata", "test.json"), "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			req: &req{
				email: "grace-staff@gsa.gov",
			},
		},
		"no arguments": {
			args: []string{"cmd2"},
			err:  "request must be set",
			req:  &req{},
		},
		"missing file": {
			args: []string{"cmd3", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "open test: no such file or directory",
			req: &req{},
		},
		"CIRCLE_TOKEN not set": {
			args: []string{"cmd4", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable CIRCLE_TOKEN must be set if format is 'terraform'",
			req: &req{},
		},
		"GITHUB_TOKEN not set": {
			args: []string{"cmd5", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable GITHUB_TOKEN must be set if format is 'terraform'",
			req: &req{},
		},
		"SN_INSTANCE not set": {
			args: []string{"cmd6", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable SN_INSTANCE must be set if format is 'terraform'",
			req: &req{},
		},
		"SN_PASSWORD not set": {
			args: []string{"cmd7", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "",
				"SN_USER":      "test",
			},
			err: "environment variable SN_PASSWORD must be set if format is 'terraform'",
			req: &req{},
		},
		"SN_USER not set": {
			args: []string{"cmd8", "-request", "test", "-format", "terraform", "-repo", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "",
			},
			err: "environment variable SN_USER must be set if format is 'terraform'",
			req: &req{},
		},
	}
	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resetEnv(tc.args, tc.env)
			req, err := newReq()
			if tc.err == "" && err != nil {
				t.Errorf("newReq() failed: unexpected error: %v", err)
			} else if tc.err != "" && (err == nil || tc.err != err.Error()) {
				t.Errorf("newReq() failed: expected error: %s\nGot: %v\n", tc.err, err)
			}
			if tc.req.email != req.email {
				t.Errorf("newReq() failed: expected: %v\ngot: %v\n", tc.req.email, req.email)
			}
			t.Logf("CircleCI Client: %v\n", req.circleClient)
		})
	}

	resetEnv(oldArgs, oldEnv)
}

// Tests that it has a non zero exit condition when there is an error
func TestCheckErr(t *testing.T) {
	if os.Getenv("BE_CHECKERR") == "1" {
		r := req{}
		r.checkErr(fmt.Errorf("testing"))
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckErr") // #nosec G204
	cmd.Env = append(os.Environ(), "BE_CHECKERR=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestRandStart(t *testing.T) {
	min := backupStartHour * 60
	max := int(math.Abs(float64(backupEndHour-backupStartHour)))*60 - backupWindowSize
	i := randStart()
	if i < min || i > max {
		t.Errorf("randStart() failed: value outside range %d - %d", min, max)
	}
}

func TestBackupWindow(t *testing.T) {
	expected := "05:11-05:41"
	w := backupWindow(311)
	if w != expected {
		t.Errorf("backupWindow() failed: expecting: %q got: %q", expected, w)
	}
}

func TestMaintenanceWindow(t *testing.T) {
	expected := "Thu:05:42-Thu:06:12"
	w := maintenanceWindow(311)
	if w != expected {
		t.Errorf("maintenanceWindow() failed: expecting: %q got: %q", expected, w)
	}
}

func TestHandleRITM(t *testing.T) {
	oldArgs, oldEnv := captureEnv()
	tt := map[string]struct {
		args []string
		env  map[string]string
		err  string
		req  *req
	}{
		"happy": {
			args: []string{"cmd1", "-request", filepath.Join("testdata", "test.json"),
				"-format", "json", "-outfile", filepath.Join(os.TempDir(), "test_out.json")},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
		},
	}
	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resetEnv(tc.args, tc.env)
			handleRITM()
			err := os.Remove(filepath.Join(os.TempDir(), "test_out.json"))
			if err != nil {
				t.Fatalf("handleRITM() failed. Unable to remove testFile: %v", err)
			}
		})
	}

	resetEnv(oldArgs, oldEnv)
}
