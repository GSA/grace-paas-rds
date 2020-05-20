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

	"github.com/andrewstuart/servicenow"
	git "github.com/go-git/go-git/v5"
	"github.com/google/go-github/v28/github"
	"github.com/jszwedko/go-circleci"
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

// req is a provisioning request object
type req struct {
	circleClient *circleci.Client
	email        string
	fullPath     string
	githubClient *github.Client
	githubURL    string
	inFile       string
	ritm         *ritm
	relPath      string
	repo         *git.Repository
	repoName     string
	snowClient   *servicenow.Client
	tempDir      string
}

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

func newReq() (*req, error) {
	var r req
	err := check()
	if err != nil {
		return &r, err
	}

	r.inFile = os.Args[1]
	r.repoName = os.Args[2]
	r.email = "grace-staff@gsa.gov"
	r.githubURL = "https://github.com/GSA/"
	r.circleClient = newCircleClient(os.Getenv("CIRCLE_TOKEN"))
	r.githubClient = newAuthenticatedClient()
	r.snowClient = newSnowClient()

	r.ritm, err = parseRITM(r.inFile)
	if err != nil {
		return &r, err
	}

	r.relPath = filepath.Join("terraform", "rds_"+r.ritm.Number+".tf.json")
	r.tempDir = filepath.Join(os.TempDir(), r.repoName)
	r.fullPath = filepath.Join(r.tempDir, r.relPath)

	return &r, nil
}

func (r *req) checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		if r.ritm != nil {
			err := r.updateRITM(err)
			if err != nil {
				fmt.Println(err)
			}
		}
		os.Exit(1)
	}
}

func handleRITM(opt ...*req) {
	var r *req
	var err error
	if len(opt) > 0 {
		r = opt[0]
	} else {
		r, err = newReq()
		r.checkErr(err)
	}

	repo, err := r.cloneRepo()
	r.checkErr(err)

	r.repo = repo
	tf := r.ritm.generateTerraform()

	err = r.newBranch()
	r.checkErr(err)

	err = tf.writeFile(r.fullPath)
	r.checkErr(err)

	_, err = r.addPassword()
	r.checkErr(err)

	err = r.commit()
	r.checkErr(err)

	err = os.RemoveAll(r.tempDir) // Remove the cloned repo after pushing
	r.checkErr(err)

	pr, err := r.pullRequest()
	r.checkErr(err)

	err = waitForMerge(pr)
	r.checkErr(err)

	err = waitForApply(pr)
	r.checkErr(err)

	err = r.updateRITM(nil)
	r.checkErr(err)

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

func check() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: %s <inFile> <repoName>", filepath.Base(os.Args[0]))
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("environment variable GITHUB_TOKEN must be set")
	}

	if os.Getenv("CIRCLE_TOKEN") == "" {
		return fmt.Errorf("environment variable CIRCLE_TOKEN must be set")
	}

	if os.Getenv("SN_INSTANCE") == "" {
		return fmt.Errorf("environment variable SN_INSTANCE must be set")
	}

	if os.Getenv("SN_PASSWORD") == "" {
		return fmt.Errorf("environment variable SN_PASSWORD must be set")
	}

	if os.Getenv("SN_USER") == "" {
		return fmt.Errorf("environment variable SN_USER must be set")
	}
	return nil
}

func main() {
	handleRITM()
}
