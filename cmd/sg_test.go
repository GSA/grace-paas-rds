package main

import "testing"

func TestSecurityGroup(t *testing.T) {
	var tf *terraform
	sg := tf.securityGroup("test_rds", "test-rds", 5700)
	eName := "test-rds-SG"
	ePort := 5700
	ecidrBlock := "${var.test_rds_mgmt_cidr_blocks}"
	if sg["name"] != eName {
		t.Errorf("*terraform.rdsEngineDefaults() failed: incorrect name. Expected: %s\n Got: %s\n", eName, sg["name"])
	}
	port := sg["ingress"].([2]map[string]interface{})[0]["from_port"]
	if port != ePort {
		t.Errorf("*terraform.rdsEngineDefaults() failed: incorrect port. Expected: %d\n Got: %s\n", ePort, port)
	}
	cidrBlock := sg["ingress"].([2]map[string]interface{})[1]["cidr_blocks"]
	if cidrBlock != ecidrBlock {
		t.Errorf("*terraform.rdsEngineDefaults() failed: incorrect management cidr_blocks. Expected: %s\n Got: %s\n", ecidrBlock, cidrBlock)
	}
}
