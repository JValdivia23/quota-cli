package opencodebar

import "fmt"

// QuotaInfo represents the structure of quota data
type QuotaInfo struct {
	ProjectCode string
	Used        float64
	Limit       float64
	Type        string // e.g., "storage", "compute"
}

// FetchQuota simulates fetching quota data (to be replaced with actual HPC API logic)
func FetchQuota(project string) ([]QuotaInfo, error) {
	// TODO: Implement actual quota retrieval logic for Derecho/Casper
	// For now, return mock data
	mockData := []QuotaInfo{
		{ProjectCode: project, Used: 1.5, Limit: 10.0, Type: "storage (TB)"},
		{ProjectCode: project, Used: 50000, Limit: 100000, Type: "compute (core-hours)"},
	}
	return mockData, nil
}

// FormatQuota returns a clear string representation of the quota
func FormatQuota(data []QuotaInfo) string {
	result := "Quota Usage:\n"
	for _, q := range data {
		result += fmt.Sprintf("  [%s] %s: %.2f / %.2f\n", q.ProjectCode, q.Type, q.Used, q.Limit)
	}
	return result
}
