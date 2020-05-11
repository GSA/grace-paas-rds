package main

// rdsModuleDefaults sets the default RDS Module parameters
func (t *terraform) rdsModuleDefaults() map[string]interface{} {
	m := map[string]interface{}{
		"source":                                "terraform-aws-modules/rds/aws",
		"version":                               "~> 2.0",
		"backup_retention_period":               31, // days
		"create_db_option_group":                false,
		"create_db_parameter_group":             false,
		"create_monitoring_role":                true,
		"deletion_protection":                   true,
		"monitoring_interval":                   5, // minutes
		"performance_insights_enabled":          true,
		"performance_insights_retention_period": 7, // days
		"publicly_accessible":                   false,
		"storage_encrypted":                     true,
	}
	return m
}
