package main

import "testing"

func TestRdsEngineDefaults(t *testing.T) {
	var tf *terraform
	defaults := tf.rdsEngineDefaults()
	if defaults["mysql5.7"].(map[string]interface{})["engine_version"] != "5.7.28" {
		t.Errorf("*terraform.rdsEngineDefaults() failed: incorrect MySQL5.7 version. Expected: %s Got: %s",
			"5.7.28", defaults["mysql5.7"].(map[string]interface{})["engine_version"])
	}
}
