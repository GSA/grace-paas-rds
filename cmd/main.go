// Command to generate Terrafrom JSON Configuration Syntax to create an AWS
// Relational Database Service (RDS) resource from a ServiceNow ticket (json)
//
// The command takes two arguments: the input file containing the JSON from the
// ServiceNow requested item, and the path/name of the output file for the JSON
// syntax configuration file. The JSON syntax file should have a `.tf.json`
// extension.

// For example:
//
// ```
// $ grace-paas-rds RITM.json rds.tf.json
// ```
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	maxPort          = 65535
	minPort          = 1150
	backupStartHour  = 3     // 0300 UTC 11:00PM ET
	backupEndHour    = 9     // 0900 UTC 5:00AM ET
	backupWindowSize = 30    // minutes
	maintenanceDay   = "Thu" // Thursday...assuming we aren't crossing a day boundary
	yes              = "Yes" // ServiceNow uses "Yes"/"No" instead of booleans
)

// ritm type for the parsed ServiceNow RITM results JSON
type ritm struct {
	Account         string `json:"account"`       // "grace-paas-developent",
	CatalogItemName string `json:"cat_item_name"` // "GRACE-PaaS AWS RDS Provisioning Request",
	Comments        string `json:"comments"`      // "",
	Engine          string `json:"engine"`        // "mysql8.0",
	Identifier      string `json:"identifier"`    // "test-rds",
	MultiAZ         string `json:"multi_az"`      // false,
	Name            string `json:"name"`          // "TestDB",
	Number          string `json:"number"`        // "RITM0001001",
	OpenedBy        string `json:"opened_by"`     // "by@email.com",
	Password        string `json:"password"`      // not actually in RITM, but randomly generated
	RequestedFor    string `json:"requested_for"` // "for@email.com",
	Size            string `json:"size"`          // "small",
	Supervisor      string `json:"supervisor"`    // "supervisor@email.com",
	SysID           string `json:"sys_id"`        // "99aa00000aa9aa00a9a99999a99aaa99",
	Username        string `json:"username"`      // "TestUser"
}

func handleRITM(inFile, repoName string) {
	ritm, err := parseRITM(inFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	tf := ritm.generateTerraform()

	repo, err := cloneRepo(repoName)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = newBranch(repo, ritm.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	relPath := "terraform/rds_" + ritm.Number + ".tf.json"
	fullPath := "/tmp/" + repoName + "/" + relPath

	err = tf.writeFile(fullPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = addPassword(repoName, ritm)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = commit(repo, relPath, ritm.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	pr, err := pullRequest(ritm, repoName)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.RemoveAll("/tmp/" + repoName + "/") // Remove the cloned repo after pushing
	if err != nil {
		fmt.Println(err)
		return
	}

	err = waitForMerge(pr)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = waitForApply(pr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Processing complete")
}

func parseRITM(inFile string) (*ritm, error) {
	fmt.Printf("Parsing RITM from: %s\n", inFile)
	jsonFile, err := os.Open(inFile) // #nosec G304
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close() // #nosec G307
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var ritm ritm
	err = json.Unmarshal(byteValue, &ritm)
	if err != nil {
		return nil, err
	}

	return &ritm, nil
}

func randStart() int {
	min := backupStartHour * 60
	max := int(math.Abs(float64(backupEndHour-backupStartHour)))*60 - backupWindowSize
	return rand.Intn(max-min) + min
}

func backupWindow(m int) string {
	e := m + backupWindowSize
	return fmt.Sprintf("%02d:%02d-%02d:%02d", m/60, m%60, e/60, e%60)
}

func maintenanceWindow(m int) string {
	// Does not support spanning multiple days
	m = m + backupWindowSize + 1
	e := m + backupWindowSize
	return fmt.Sprintf("%s:%02d:%02d-%s:%02d:%02d", maintenanceDay, m/60, m%60, maintenanceDay, e/60, e%60)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <inFile> <repoName>\n", filepath.Base(os.Args[0]))
		return
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Printf("GITHUB_TOKEN environment variable must be set.")
		return
	}

	if os.Getenv("CIRCLE_TOKEN") == "" {
		fmt.Printf("CIRCLE_TOKEN environment variable must be set.")
		return
	}

	inFile := os.Args[1]
	repoName := os.Args[2]

	handleRITM(inFile, repoName)
}
