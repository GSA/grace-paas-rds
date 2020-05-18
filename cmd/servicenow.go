package main

import (
	"fmt"
	"os"

	"github.com/andrewstuart/servicenow"
)

func updateRITM(ritm *ritm, e error) error {
	fmt.Printf("Updating %s (%s)\n", ritm.Number, ritm.SysID)
	client := servicenow.Client{
		Username: os.Getenv("SN_USER"),
		Password: os.Getenv("SN_PASSWORD"),
		Instance: os.Getenv("SN_INSTANCE"),
	}
	table := "sc_req_item"
	var out map[string]interface{}
	var state = 2 // Work in Progress
	var comment = "RDS Provisioning complete via GRACE-PaaS CI/CD Pipeline"
	if e != nil {
		state = 8 // Reopened
		comment = fmt.Sprintf("Error provisioning RDS: %v", e)
	}
	body := map[string]interface{}{
		"state":    state,
		"comments": comment,
	}

	return client.PerformFor(table, "update", ritm.SysID, nil, body, &out)
}
