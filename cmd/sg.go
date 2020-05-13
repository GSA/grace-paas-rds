package main

func (tf *terraform) securityGroup(resourceID, id string, port int) map[string]interface{} {
	return map[string]interface{}{
		"name":        id + "-SG",
		"description": "Allow RDS inboud traffic",
		"vpc_id":      "${module.network.back_vpc_id}",
		"ingress": [...]map[string]interface{}{
			{
				"description":      "Mid VPC",
				"from_port":        port,
				"to_port":          port,
				"protocol":         "TCP",
				"cidr_blocks":      [...]string{"${module.network.mid_vpc_cidr}"},
				"ipv6_cidr_blocks": [...]string{},
				"prefix_list_ids":  [...]string{},
				"security_groups":  [...]string{},
				"self":             false,
			},
			{
				"description":      "DBMW Mgmt",
				"from_port":        port,
				"to_port":          port,
				"protocol":         "TCP",
				"cidr_blocks":      "${var." + resourceID + "_mgmt_cidr_blocks}",
				"ipv6_cidr_blocks": [...]string{},
				"prefix_list_ids":  [...]string{},
				"security_groups":  [...]string{},
				"self":             false,
			},
		},
	}
}
