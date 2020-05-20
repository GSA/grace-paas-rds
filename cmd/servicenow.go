package main

import (
	"fmt"
	"os"

	"github.com/andrewstuart/servicenow"
)

func newSnowClient() *servicenow.Client {
	return &servicenow.Client{
		Username: os.Getenv("SN_USER"),
		Password: os.Getenv("SN_PASSWORD"),
		Instance: os.Getenv("SN_INSTANCE"),
	}
}

func (r *req) updateRITM(e error) error {
	fmt.Printf("Updating %s (%s)\n", r.ritm.Number, r.ritm.SysID)
	table := "sc_req_item"
	var out map[string]interface{}
	var state = 2 // Work in Progress
	var comment = "RDS Provisioned via GRACE-PaaS CI/CD Pipeline"
	if e != nil {
		state = 8 // Reopened
		comment = fmt.Sprintf("Error provisioning RDS: %v", e)
	}
	body := map[string]interface{}{
		"state":    state,
		"comments": comment,
	}

	return r.snowClient.PerformFor(table, "update", r.ritm.SysID, nil, body, &out)
}
