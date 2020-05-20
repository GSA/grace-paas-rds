package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

type terraform struct {
	Map map[string]interface{}
}

func (ritm *ritm) generateTerraform() terraform {
	fmt.Println("Generating terraform")
	var tf terraform
	rand.Seed(time.Now().UnixNano())
	module := tf.rdsModule(ritm)
	resourceID := strings.ReplaceAll(ritm.Identifier, "-", "_") // Conforms to our naming standard

	tf.Map = map[string]interface{}{
		"variable": [...]map[string]interface{}{{
			resourceID + "_db_password": map[string]interface{}{
				"type":        "string",
				"description": "(required) RDS user password",
			}},
			{
				resourceID + "_mgmt_cidr_blocks": map[string]interface{}{
					"type":        "list(string)",
					"description": "(optional) List of CIDR blocks from which to manage RDS",
					"default":     [...]string{},
				}},
		},
		"module": map[string]interface{}{
			resourceID: module,
		},
		"resource": [...]map[string]interface{}{{
			"aws_security_group": map[string]interface{}{
				resourceID: tf.securityGroup(resourceID, ritm.Identifier, module["port"].(int)),
			},
			"aws_kms_key": map[string]interface{}{
				resourceID: map[string]interface{}{
					"description":         ritm.Identifier + " RDS KMS Key",
					"enable_key_rotation": true,
				},
			},
			"aws_kms_alias": map[string]interface{}{
				resourceID: map[string]interface{}{
					"name":          "alias/" + resourceID,
					"target_key_id": "${aws_kms_key." + resourceID + ".key_id}",
				},
			},
			"aws_ssm_parameter": map[string]interface{}{
				resourceID + "_password": map[string]interface{}{
					"name":        "/database/password/" + ritm.Identifier,
					"description": ritm.Identifier + " RDS Master Password",
					"type":        "SecureString",
					"value":       "${var." + resourceID + "_db_password}",
					"key_id":      "${aws_kms_key." + resourceID + ".arn}",
				},
			},
		},
		},
	}

	return tf
}

func (tf *terraform) rdsModule(ritm *ritm) map[string]interface{} {
	defaults := tf.rdsModuleDefaults()
	engines := tf.rdsEngineDefaults()
	family := ritm.Engine
	options := engines[family].(map[string]interface{})
	engine := options["engine"]
	backupStartTime := randStart()                              // Number of minutes after start of backupwindow start hour
	resourceID := strings.ReplaceAll(ritm.Identifier, "-", "_") // Conforms to our naming standard

	// Override and add to defaults
	defaults["identifier"] = ritm.Identifier
	defaults["engine"] = engine
	defaults["engine_version"] = options["engine_version"]
	defaults["enabled_cloudwatch_logs_exports"] = options["enabled_cloudwatch_logs_exports"]
	defaults["instance_class"] = options[ritm.Size].(map[string]interface{})["instance_class"]
	defaults["kms_key_id"] = "${aws_kms_key." + resourceID + ".arn}"
	defaults["allocated_storage"] = options[ritm.Size].(map[string]interface{})["allocated_storage"]
	defaults["name"] = ritm.Name
	defaults["username"] = ritm.Username
	defaults["password"] = "${var." + resourceID + "_db_password}"
	defaults["port"] = rand.Intn(maxPort-minPort) + minPort
	defaults["backup_window"] = backupWindow(backupStartTime)
	defaults["maintenance_window"] = maintenanceWindow(backupStartTime)
	defaults["final_snapshot_identifier"] = ritm.Identifier + "-final-shapshot"
	defaults["major_engine_version"] = options["major_engine_version"]
	defaults["max_allocated_storage"] = 3 * defaults["allocated_storage"].(int)
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
	if ritm.MultiAZ == yes {
		defaults["multi_az"] = true
		defaults["subnet_ids"] = "${module.network.back_vpc_subnet_ids}"
	}
	defaults["vpc_security_group_ids"] = [...]string{"${aws_security_group." + resourceID + ".id}"}

	return defaults
}

func (tf *terraform) writeFile(outFile string) error {
	fmt.Printf("Writing terraform to file: %s\n", outFile)
	b, err := json.MarshalIndent(tf.Map, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outFile, b, 0600)
	return err
}
