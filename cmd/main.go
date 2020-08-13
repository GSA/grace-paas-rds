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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

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
	tfConst          = "terraform"
)

// req is a provisioning request object
type req struct {
	circleClient *circleci.Client
	email        string
	fullPath     string
	format       string // json or terraform
	githubClient *github.Client
	githubURL    string
	inFile       string
	ritm         *ritm
	relPath      string
	repo         *git.Repository
	repoName     string
	snowClient   *servicenow.Client
	tempDir      string
	reqMap       map[string]interface{}
}

// ritm type for the parsed ServiceNow RITM results JSON
type ritm struct {
	Account         string `json:"account"`              // "grace-paas-developent",
	CatalogItemName string `json:"cat_item_name"`        // "GRACE-PaaS AWS RDS Provisioning Request",
	Comments        string `json:"comments"`             // "",
	Engine          string `json:"engine"`               // "mysql8.0",
	Identifier      string `json:"identifier"`           // "test-rds",
	DevCount        string `json:"development_count"`    // 1,
	DevMultiAZ      string `json:"development_multi_az"` // false,
	DevSize         string `json:"development_size"`     // "small",
	ProdCount       string `json:"production_count"`     // 1,
	ProdMultiAZ     string `json:"production_multi_az"`  // false,
	ProdSize        string `json:"production_size"`      // "small",
	TestCount       string `json:"test_count"`           // 1,
	TestMultiAZ     string `json:"test_multi_az"`        // false,
	TestSize        string `json:"test_size"`            // "small",
	Name            string `json:"name"`                 // "TestDB",
	Number          string `json:"number"`               // "RITM0001001",
	OpenedBy        string `json:"opened_by"`            // "by@email.com",
	Password        string `json:"password"`             // not actually in RITM, but randomly generated
	RequestedFor    string `json:"requested_for"`        // "for@email.com",
	Supervisor      string `json:"supervisor"`           // "supervisor@email.com",
	SysID           string `json:"sys_id"`               // "99aa00000aa9aa00a9a99999a99aaa99",
	Username        string `json:"username"`             // "TestUser"
}

func newReq() (*req, error) {
	var r req
	flags, output, err := r.parseFlags(os.Args[0], os.Args[1:])
	if err != nil {
		fmt.Println(output)
		return &r, err
	}

	err = r.check()
	if err != nil {
		flags.PrintDefaults()
		return &r, err
	}

	err = r.parseRITM()
	if err != nil {
		return &r, err
	}

	if r.format == tfConst {
		r.email = "grace-staff@gsa.gov"
		r.githubURL = "https://github.com/GSA/"
		r.circleClient = newCircleClient(os.Getenv("CIRCLE_TOKEN"))
		r.githubClient = newAuthenticatedClient()
		r.snowClient = newSnowClient()

		r.relPath = filepath.Join(tfConst, "rds_"+r.ritm.Number+".tf.json")
		r.tempDir = filepath.Join(os.TempDir(), r.repoName)
		r.fullPath = filepath.Join(r.tempDir, r.relPath)
	}

	return &r, nil
}

func (r *req) parseFlags(progName string, args []string) (*flag.FlagSet, string, error) {
	flags := flag.NewFlagSet(progName, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)
	flags.StringVar(&r.inFile, "request", "", "JSON input file")
	flags.StringVar(&r.relPath, "outfile", "", "JSON output file")
	flags.StringVar(&r.repoName, "repo", "", "Repo name")
	flags.StringVar(&r.format, "format", "json", "Output file format: json or terraform")
	err := flags.Parse(args)
	if err != nil {
		return flags, buf.String(), err
	}
	return flags, buf.String(), nil
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

	switch format := r.format; format {
	case tfConst:
		r.handleTerraform()
	default:
		r.handleJSON()
	}
}

func (r *req) handleTerraform() {
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

func (r *req) handleJSON() {
	var tf terraform
	engines := tf.rdsEngineDefaults()
	family := r.ritm.Engine
	options := engines[family].(map[string]interface{})
	engine := options["engine"]
	backupStartTime := randStart() // Number of minutes after start of backupwindow start hour

	// Complete request for grace-actions
	r.reqMap["action"] = "rds"
	r.reqMap["engine"] = engine
	r.reqMap["engine_major_version"] = options["major_engine_version"]
	r.reqMap["engine_version"] = options["engine_version"]
	r.reqMap["port"] = options["port"]
	r.reqMap["enabled_cloudwatch_logs_exports"] = strings.Join(options["enabled_cloudwatch_logs_exports"].([]string), ",")
	r.reqMap["backup_window"] = backupWindow(backupStartTime)
	r.reqMap["maintenance_window"] = maintenanceWindow(backupStartTime)
	r.reqMap["development_instance_class"] = options[r.ritm.DevSize].(map[string]interface{})["instance_class"]
	r.reqMap["test_instance_class"] = options[r.ritm.TestSize].(map[string]interface{})["instance_class"]
	r.reqMap["production_instance_class"] = options[r.ritm.ProdSize].(map[string]interface{})["instance_class"]
	r.reqMap["development_allocated_storage"] = options[r.ritm.DevSize].(map[string]interface{})["allocated_storage"]
	r.reqMap["test_allocated_storage"] = options[r.ritm.TestSize].(map[string]interface{})["allocated_storage"]
	r.reqMap["production_allocated_storage"] = options[r.ritm.ProdSize].(map[string]interface{})["allocated_storage"]

	err := r.writeFile()
	r.checkErr(err)

	fmt.Println("Processing complete")
}

func (r *req) writeFile() error {
	fmt.Printf("Writing json to file: %s\n", r.relPath)
	b, err := json.MarshalIndent(r.reqMap, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(r.relPath, b, 0600)
	return err
}

func (r *req) parseRITM() error {
	fmt.Printf("Parsing RITM from: %s\n", r.inFile)
	jsonFile, err := os.Open(r.inFile) // #nosec G304
	if err != nil {
		return err
	}

	defer jsonFile.Close() // #nosec G307
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var ritm ritm
	err = json.Unmarshal(byteValue, &ritm)
	if err != nil {
		return err
	}
	r.ritm = &ritm

	var myMap map[string]interface{}
	err = json.Unmarshal(byteValue, &myMap)
	if err != nil {
		return err
	}
	r.reqMap = myMap

	return nil
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

func (r *req) check() error {
	if r.inFile == "" {
		return fmt.Errorf("request must be set")
	}

	if r.format == "json" {
		if r.relPath == "" {
			return fmt.Errorf("outfile must be set if format is json")
		}
		return nil
	}

	if r.repoName == "" {
		return fmt.Errorf("reponame must be set if format is 'terraform'")
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("environment variable GITHUB_TOKEN must be set if format is 'terraform'")
	}

	if os.Getenv("CIRCLE_TOKEN") == "" {
		return fmt.Errorf("environment variable CIRCLE_TOKEN must be set if format is 'terraform'")
	}

	if os.Getenv("SN_INSTANCE") == "" {
		return fmt.Errorf("environment variable SN_INSTANCE must be set if format is 'terraform'")
	}

	if os.Getenv("SN_PASSWORD") == "" {
		return fmt.Errorf("environment variable SN_PASSWORD must be set if format is 'terraform'")
	}

	if os.Getenv("SN_USER") == "" {
		return fmt.Errorf("environment variable SN_USER must be set if format is 'terraform'")
	}
	return nil
}

func main() {
	handleRITM()
}
