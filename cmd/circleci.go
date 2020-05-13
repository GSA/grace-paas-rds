package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jszwedko/go-circleci"
)

func addPassword(repo string, ritm *ritm) (*circleci.EnvVar, error) {
	resourceID := strings.ReplaceAll(ritm.Identifier, "-", "_")
	name := "TF_VAR_" + resourceID + "_db_password"
	value := generatePassword()
	fmt.Printf("Creating CircleCI environment variable %s in %s project\n", name, repo)
	client := &circleci.Client{Token: os.Getenv("CIRCLE_TOKEN")}
	return client.AddEnvVar("GSA", repo, name, value)
}
