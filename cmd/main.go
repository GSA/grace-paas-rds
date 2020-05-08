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
	"strconv"
	"time"
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
	AccountSysID                       string `json:"account_sys_id"`                        // "99aa00000aa9aa00a9a99999a99aaa99",
	Identifier                         string `json:"identifier"`                            // "test-rds",
	Comments                           string `json:"comments"`                              // "",
	CatalogItemName                    string `json:"cat_item_name"`                         // "GRACE-PaaS AWS RDS Provisioning Request",
	MultiAZ                            string `json:"multi_az"`                              // false,
	PerformanceInsightsEnabled         string `json:"performance_insights_enabled"`          // false,
	RequestedFor                       string `json:"requested_for"`                         // "for@email.com",
	PerformanceInsightsRetentionPeriod string `json:"performance_insights_retention_period"` // 7,
	Number                             string `json:"number"`                                // "RITM0001001",
	SysID                              string `json:"sys_id"`                                // "99aa00000aa9aa00a9a99999a99aaa99",
	AccountID                          string `json:"account_id"`                            // "123456789012",
	Size                               string `json:"size"`                                  // "small",
	OpenedBy                           string `json:"opened_by"`                             // "by@email.com",
	Engine                             string `json:"engine"`                                // "mysql8.0",
	Name                               string `json:"name"`                                  // "TestDB",
	MonitoringInterval                 string `json:"monitoring_interval"`                   // 5,
	Supervisor                         string `json:"supervisor"`                            // "supervisor@email.com",
	Username                           string `json:"username"`                              // "TestUser"
}

type terraform struct {
	Map map[string]interface{}
}

func handleRITM(inFile, outFile string) {
	ritm, err := parseRITM(inFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = generateTerraform(ritm, outFile)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func parseRITM(inFile string) (*ritm, error) {
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

// nolint: funlen
func generateTerraform(ritm *ritm, outFile string) error {
	var tf terraform
	defaults := tf.rdsModuleDefaults()
	engines := tf.rdsEngineDefaults()
	family := ritm.Engine
	options := engines[family].(map[string]interface{})
	engine := options["engine"]
	rand.Seed(time.Now().UnixNano())
	backupStartTime := randStart()

	// Override and add to defaults
	defaults["identifier"] = ritm.Identifier
	defaults["engine"] = engine
	defaults["engine_version"] = options["engine_version"]
	defaults["enabled_cloudwatch_logs_exports"] = options["enabled_cloudwatch_logs_exports"]
	defaults["instance_class"] = options[ritm.Size].(map[string]interface{})["instance_class"]
	defaults["kms_key_id"] = "${aws_kms_key." + ritm.Identifier + ".id}"
	defaults["allocated_storage"] = options[ritm.Size].(map[string]interface{})["allocated_storage"]
	defaults["name"] = ritm.Name
	defaults["username"] = ritm.Username
	defaults["password"] = "${var.db_password}"
	defaults["port"] = rand.Intn(maxPort-minPort) + minPort
	defaults["backup_window"] = backupWindow(backupStartTime)
	defaults["maintenance_window"] = maintenanceWindow(backupStartTime)
	defaults["final_snapshot_identifier"] = ritm.Identifier + "-final-shapshot"
	defaults["major_engine_version"] = options["major_engine_version"]
	defaults["max_allocated_storage"] = 3 * defaults["allocated_storage"].(int)
	mi, err := strconv.Atoi(ritm.MonitoringInterval)
	if err != nil {
		fmt.Printf("WARNING: Error converting monitoring_interval to integer: %v\n", err)
	}
	defaults["monitoring_interval"] = mi
	defaults["monitoring_role_name"] = ritm.Identifier + "-monitoring-role"
	/* Enable once custom property/option groups are defined
	if engine == "mysql" {
		defaults["option_group_name"] = "grace.paas." + engine + "-" + options["major_engine_version"]
	}
	if engine == "postgres" {
		defaults["parameter_group_name"] = "grace.paas." + engine + "-" + options["major_engine_version"]
	}
	defaults["use_parameter_group_name_prefix"] = false
	*/
	if ritm.PerformanceInsightsEnabled == yes {
		defaults["performance_insights_enabled"] = true
		defaults["performance_insights_retention_period"], err = strconv.Atoi(ritm.PerformanceInsightsRetentionPeriod)
		if err != nil {
			fmt.Printf("WARNING: Error converting to performance_insights_retention_period integer: %v\n", err)
		}
	}
	if ritm.MultiAZ == yes {
		defaults["multi_az"] = true
		defaults["subnet_ids"] = "${module.network.back_vpc_subnet_ids}"
	}
	defaults["vpc_security_group_ids"] = [...]string{"${aws_security_group." + ritm.Identifier + ".id}"}

	tf.Map = map[string]interface{}{
		"variable": [...]map[string]interface{}{{
			"db_password": map[string]interface{}{
				"type":        "string",
				"description": "(required) RDS user password",
			},
			"rds_mgmt_cidr_blocks": map[string]interface{}{
				"type":        "list(string)",
				"description": "(optional) List of CIDR blocks from which to manage RDS",
				"default":     [...]string{},
			}},
		},
		"module": map[string]interface{}{
			ritm.Identifier: defaults,
		},
		"resource": [...]map[string]interface{}{{
			"aws_security_group": map[string]interface{}{
				ritm.Identifier: map[string]interface{}{
					"name":        ritm.Identifier + "-SG",
					"description": "Allow RDS inboud traffic",
					"vpc_id":      "${module.network.back_vpc_id}",
					"ingress": [...]map[string]interface{}{
						{
							"description":      "Mid VPC",
							"from_port":        defaults["port"],
							"to_port":          defaults["port"],
							"protocol":         "TCP",
							"cidr_blocks":      [...]string{"${module.network.mid_vpc_cidr}"},
							"ipv6_cidr_blocks": [...]string{},
							"prefix_list_ids":  [...]string{},
							"security_groups":  [...]string{},
							"self":             false,
						},
						{
							"description":      "DBMW Mgmt",
							"from_port":        defaults["port"],
							"to_port":          defaults["port"],
							"protocol":         "TCP",
							"cidr_blocks":      "${var.rds_mgmt_cidr_blocks}",
							"ipv6_cidr_blocks": [...]string{},
							"prefix_list_ids":  [...]string{},
							"security_groups":  [...]string{},
							"self":             false,
						},
					},
				},
			},
			"aws_kms_key": map[string]interface{}{
				ritm.Identifier: map[string]interface{}{
					"description":         ritm.Identifier + " RDS KMS Key",
					"enable_key_rotation": true,
				},
			},
			"aws_kms_alias": map[string]interface{}{
				ritm.Identifier: map[string]interface{}{
					"name":          "alias/" + ritm.Identifier,
					"target_key_id": "${aws_kms_key." + ritm.Identifier + ".key_id}",
				},
			},
		},
		},
	}

	b, err := json.MarshalIndent(tf.Map, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Terraform: \n%v\n", string(b))
	err = ioutil.WriteFile(outFile, b, 0600)
	return err
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <inFile> <outFile>\n", filepath.Base(os.Args[0]))
		return
	}

	inFile := os.Args[1]
	outFile := os.Args[2]

	handleRITM(inFile, outFile)
}
