package main

func (tf *terraform) rdsEngineDefaults() map[string]interface{} {
	m := map[string]interface{}{
		"mysql5.7": map[string]interface{}{
			"description":                     "MySQL Community Edition",
			"engine":                          "mysql",
			"engine_version":                  "5.7.30",
			"family":                          "mysql5.7",
			"major_engine_version":            "5.7",
			"port":                            3306,
			"enabled_cloudwatch_logs_exports": []string{"audit", "error", "general", "slowquery"},
			"small": map[string]interface{}{
				"instance_class":    "db.m5.large",
				"allocated_storage": 50,
			},
			"medium": map[string]interface{}{
				"instance_class":    "db.m5.xlarge",
				"allocated_storage": 100,
			},
			"large": map[string]interface{}{
				"instance_class":    "db.m5.2xlarge",
				"allocated_storage": 300,
			},
		},
		"mysql8.0": map[string]interface{}{
			"engine":                          "mysql",
			"engine_version":                  "8.0.20",
			"family":                          "mysql8.0",
			"major_engine_version":            "8.0",
			"port":                            3306,
			"description":                     "MySQL Community Edition",
			"enabled_cloudwatch_logs_exports": []string{"error", "general", "slowquery"},
			"small": map[string]interface{}{
				"instance_class":    "db.m5.large",
				"allocated_storage": 50,
			},
			"medium": map[string]interface{}{
				"instance_class":    "db.m5.xlarge",
				"allocated_storage": 100,
			},
			"large": map[string]interface{}{
				"instance_class":    "db.m5.2xlarge",
				"allocated_storage": 300,
			},
		},
		"postgres11": map[string]interface{}{
			"engine":                          "postgres",
			"engine_version":                  "11.8",
			"family":                          "postgres11",
			"major_engine_version":            "11",
			"port":                            5432,
			"description":                     "PostgreSQL",
			"enabled_cloudwatch_logs_exports": []string{"postgresql", "upgrade"},
			"small": map[string]interface{}{
				"instance_class":    "db.m5.large",
				"allocated_storage": 20,
			},
			"medium": map[string]interface{}{
				"instance_class":    "db.m5.xlarge",
				"allocated_storage": 40,
			},
			"large": map[string]interface{}{
				"instance_class":    "db.m5.2xlarge",
				"allocated_storage": 100,
			},
		},
		"postgres12": map[string]interface{}{
			"engine":                          "postgres",
			"engine_version":                  "12.3",
			"family":                          "postgres12",
			"major_engine_version":            "12",
			"port":                            5432,
			"description":                     "PostgreSQL",
			"enabled_cloudwatch_logs_exports": []string{"postgresql", "upgrade"},
			"small": map[string]interface{}{
				"instance_class":    "db.m5.large",
				"allocated_storage": 20,
			},
			"medium": map[string]interface{}{
				"instance_class":    "db.m5.xlarge",
				"allocated_storage": 40,
			},
			"large": map[string]interface{}{
				"instance_class":    "db.m5.2xlarge",
				"allocated_storage": 100,
			},
		},
	}
	return m
}
