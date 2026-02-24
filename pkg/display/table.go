package display

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// PrintTable formats the provider reports into a uniform CLI table.
func PrintTable(reports []*models.ProviderReport) {
	// Sort alphabetical
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Name < reports[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintln(w, "Provider\tRefresh\tUse\tKey Metrics")
	fmt.Fprintln(w, "───────────\t───────────────\t────────\t────────────────")

	for _, req := range reports {
		if req.Type == models.TypeQuotaBased {
			used := req.Entitlement - req.Remaining
			usagePct := 0
			if req.Entitlement > 0 {
				usagePct = (used * 100) / req.Entitlement
			}
			usageStr := fmt.Sprintf("%d%%", usagePct)

			refreshStr := req.RefreshTime
			if refreshStr == "" {
				refreshStr = "-"
			}

			metricStr := fmt.Sprintf("%d/%d remaining", req.Remaining, req.Entitlement)

			if len(req.Accounts) > 0 {
				for _, acc := range req.Accounts {
					accUsed := acc.Entitlement - acc.Remaining
					accUsagePct := 0
					if acc.Entitlement > 0 {
						accUsagePct = (accUsed * 100) / acc.Entitlement
					}
					accUsageStr := fmt.Sprintf("%d%%", accUsagePct)
					accMetricStr := fmt.Sprintf("%d/%d remaining", acc.Remaining, acc.Entitlement)
					fmt.Fprintf(w, "%s (%s)\t%s\t%s\t%s\n", req.Name, acc.Email, refreshStr, accUsageStr, accMetricStr)
				}
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", req.Name, refreshStr, usageStr, metricStr)
			}
		} else {
			// Pay as you go
			fmt.Fprintf(w, "%s\t-\t-\t$%.2f spent\n", req.Name, req.Cost)
		}
	}
	w.Flush()
}

// PrintJSON exports the raw structs as JSON
func PrintJSON(reports []*models.ProviderReport) {
	output := make(map[string]*models.ProviderReport)
	for _, req := range reports {
		// Mock logic; using name as key
		key := req.Name
		output[key] = req
	}

	bytes, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(bytes))
}
