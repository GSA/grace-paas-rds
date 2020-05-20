package main

import (
	"os"
	"testing"
)

func TestGenerateTerraform(t *testing.T) {
	ritm, err := parseRITM("testdata/test.json")
	if err != nil {
		t.Fatalf("generateTerraform() failed. Unable to parse test data: %v", err)
	}

	tf := ritm.generateTerraform()

	expected := "(required) RDS user password"
	got := tf.Map["variable"].([2]map[string]interface{})[0]["test_rds_db_password"].(map[string]interface{})["description"]
	if expected != got {
		t.Errorf("generateTerraform() failed. Unable to parse test data. Expected: %s\nGot(%T): %s\n", expected, got, got)
	}

	expected = "(optional) List of CIDR blocks from which to manage RDS"
	got = tf.Map["variable"].([2]map[string]interface{})[1]["test_rds_mgmt_cidr_blocks"].(map[string]interface{})["description"]
	if expected != got {
		t.Errorf("generateTerraform() failed. Unable to parse test data. Expected: %s\nGot: %s\n", expected, got)
	}
}

func TestWriteFile(t *testing.T) {
	ritm, err := parseRITM("testdata/test.json")
	if err != nil {
		t.Fatalf("*terraform.writeFile() failed. Unable to parse test data: %v", err)
	}

	tf := ritm.generateTerraform()
	fileName := os.TempDir() + "tf.json"

	err = tf.writeFile(fileName)
	if err != nil {
		t.Errorf("*terraform.writeFile(%s) failed: unexpected error: %v", fileName, err)
	}

	err = os.Remove(fileName)
	if err != nil {
		t.Fatalf("*terraform.writeFile(%s) failed. Unable to remove testFile: %v", fileName, err)
	}
}
