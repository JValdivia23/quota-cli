package display

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// PrintTable renders a clean, adaptive table from provider reports.
// Works well for any number of providers or account sub-rows.
func PrintTable(reports []*models.ProviderReport) {
	if len(reports) == 0 {
		return
	}

	// Sort: multi-account providers first, then alphabetical
	sort.Slice(reports, func(i, j int) bool {
		ai, aj := len(reports[i].Accounts) > 0, len(reports[j].Accounts) > 0
		if ai != aj {
			return ai // multi-account first
		}
		return reports[i].Name < reports[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	fmt.Fprintln(w, "Provider\tRefresh\tUse\tKey Metrics")
	printDivider(w, reports)

	for _, rep := range reports {
		// Error row (provider reached but API call failed)
		if rep.ErrorMsg != "" {
			fmt.Fprintf(w, "%s\t%s\t-\t⚠  %s\n",
				rep.Name, truncate("(unavailable)", 20), rep.ErrorMsg)
			continue
		}

		switch rep.Type {

		case models.TypeQuotaBased:
			if len(rep.Accounts) > 0 {
				// Multi-account: provider name row + indented sub-rows
				fmt.Fprintf(w, "%s\t\t\t\n", rep.Name)
				for _, acc := range rep.Accounts {
					used := acc.Entitlement - acc.Remaining
					pct := 0
					if acc.Entitlement > 0 {
						pct = (used * 100) / acc.Entitlement
					}
					metricStr := ""
					if acc.Entitlement > 0 {
						metricStr = fmt.Sprintf("%d/%d remaining", acc.Remaining, acc.Entitlement)
					} else {
						metricStr = "unlimited"
					}
					fmt.Fprintf(w, "  ↳\t%s\t%d%%\t%s\n", acc.Email, pct, metricStr)
				}
			} else {
				// Single-account quota
				used := rep.Entitlement - rep.Remaining
				pct := 0
				if rep.Entitlement > 0 {
					pct = (used * 100) / rep.Entitlement
				}
				refresh := dashIfEmpty(rep.RefreshTime)
				metricStr := ""
				if rep.Entitlement > 0 {
					metricStr = fmt.Sprintf("%d/%d remaining", rep.Remaining, rep.Entitlement)
				} else {
					metricStr = "unlimited"
				}
				fmt.Fprintf(w, "%s\t%s\t%d%%\t%s\n", rep.Name, refresh, pct, metricStr)
			}

		case models.TypeTokensBased:
			refresh := dashIfEmpty(rep.RefreshTime)
			fmt.Fprintf(w, "%s\t%s\t-\t%s tokens used\n",
				rep.Name, refresh, formatTokens(rep.TokensUsed))

		case models.TypePayAsYouGo:
			fmt.Fprintf(w, "%s\t-\t-\t$%.2f spent\n", rep.Name, rep.Cost)

		default:
			fmt.Fprintf(w, "%s\t-\t-\t-\n", rep.Name)
		}
	}

	w.Flush()
}

// printDivider prints a separator row that adapts to the terminal.
func printDivider(w *tabwriter.Writer, reports []*models.ProviderReport) {
	// Find the longest provider name to size the first column
	maxName := 8 // minimum "Provider" width
	for _, r := range reports {
		if len(r.Name) > maxName {
			maxName = len(r.Name)
		}
	}
	seg := func(n int) string { return strings.Repeat("─", n) }
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		seg(maxName+2), seg(22), seg(6), seg(20))
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func formatTokens(n int64) string {
	if n == 0 {
		return "0"
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// PrintJSON exports provider reports as structured JSON.
func PrintJSON(reports []*models.ProviderReport) {
	output := make(map[string]*models.ProviderReport, len(reports))
	for _, rep := range reports {
		output[rep.Name] = rep
	}
	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(b))
}
