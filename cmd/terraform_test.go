package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateTerraform(t *testing.T) {
	var r req
	r.inFile = filepath.Join("testdata", "test.json")

	err := r.parseRITM()
	if err != nil {
		t.Fatalf("generateTerraform() failed. Unable to parse test data: %v", err)
	}

	tf := r.ritm.generateTerraform()

	expected := "(required) RDS user password"
	got := tf.Map["variable"].([2]map[string]interface{})[0]["test_db_password"].(map[string]interface{})["description"]
	if expected != got {
		t.Errorf("generateTerraform() failed. Unable to parse test data. Expected: %s\nGot(%T): %v\n", expected, got, got)
	}

	expected = "(optional) List of CIDR blocks from which to manage RDS"
	got = tf.Map["variable"].([2]map[string]interface{})[1]["test_mgmt_cidr_blocks"].(map[string]interface{})["description"]
	if expected != got {
		t.Errorf("generateTerraform() failed. Unable to parse test data. Expected: %s\nGot: %s\n", expected, got)
	}
}

func TestWriteFile(t *testing.T) {
	var r req
	r.inFile = filepath.Join("testdata", "test.json")

	err := r.parseRITM()
	if err != nil {
		t.Fatalf("*terraform.writeFile() failed. Unable to parse test data: %v", err)
	}

	tf := r.ritm.generateTerraform()
	fileName := filepath.Join(os.TempDir(), "tf.json")

	err = tf.writeFile(fileName)
	if err != nil {
		t.Errorf("*terraform.writeFile(%s) failed: unexpected error: %v", fileName, err)
	}

	err = os.Remove(fileName)
	if err != nil {
		t.Fatalf("*terraform.writeFile(%s) failed. Unable to remove testFile: %v", fileName, err)
	}
}
