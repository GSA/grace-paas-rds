package main

import "testing"

func TestRdsModuleDefaults(t *testing.T) {
	var tf *terraform
	defaults := tf.rdsModuleDefaults()
	expected := "terraform-aws-modules/rds/aws"
	if defaults["source"] != expected {
		t.Errorf("*terraform.rdsEngineDefaults() failed: incorrect module source. Expected: %s\n Got: %s\n", expected, defaults["source"])
	}
}
