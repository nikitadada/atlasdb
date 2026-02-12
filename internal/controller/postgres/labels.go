package postgres

func Labels(clusterName string) map[string]string {
	return map[string]string{
		"app":             "pg-test",
		"postgrescluster": clusterName,
	}
}
