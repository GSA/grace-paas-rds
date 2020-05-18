package main

import (
	"os"
	"testing"
)

func resetEnv(oldArgs []string, oldEnv map[string]string) {
	os.Args = oldArgs
	for k, v := range oldEnv {
		os.Setenv(k, v)
	}
}

func TestCheck(t *testing.T) {
	oldArgs := os.Args
	oldEnv := map[string]string{
		"CIRCLE_TOKEN": os.Getenv("CIRCLE_TOKEN"),
		"GITHUB_TOKEN": os.Getenv("GITHUB_TOKEN"),
		"SN_INSTANCE":  os.Getenv("SN_INSTANCE"),
		"SN_PASSWORD":  os.Getenv("SN_PASSWORD"),
		"SN_USER":      os.Getenv("SN_USER"),
	}

	tt := map[string]struct {
		args []string
		env  map[string]string
		err  string
	}{
		"happy": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
		},
		"no arguments": {
			args: []string{"cmd"},
			err:  "usage: cmd <inFile> <repoName>",
		},
		"CIRCLE_TOKEN not set": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable CIRCLE_TOKEN must be set",
		},
		"GITHUB_TOKEN not set": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable GITHUB_TOKEN must be set",
		},
		"SN_INSTANCE not set": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "",
				"SN_PASSWORD":  "test",
				"SN_USER":      "test",
			},
			err: "environment variable SN_INSTANCE must be set",
		},
		"SN_PASSWORD not set": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "",
				"SN_USER":      "test",
			},
			err: "environment variable SN_PASSWORD must be set",
		},
		"SN_USER not set": {
			args: []string{"cmd", "test", "test"},
			env: map[string]string{
				"CIRCLE_TOKEN": "test",
				"GITHUB_TOKEN": "test",
				"SN_INSTANCE":  "test",
				"SN_PASSWORD":  "test",
				"SN_USER":      "",
			},
			err: "environment variable SN_USER must be set",
		},
	}
	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resetEnv(tc.args, tc.env)

			err := check()
			if tc.err == "" && err != nil {
				t.Errorf("check() failed: unexpected error: %v", err)
			} else if tc.err != "" && tc.err != err.Error() {
				t.Errorf("check() failed: expected error: %s\nGot: %v\n", tc.err, err)
			}
		})
	}

	resetEnv(oldArgs, oldEnv)
}
