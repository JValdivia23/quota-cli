package predictor

import (
	"time"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// CalculatePrediction implements the weighted-average forecasting logic.
func CalculatePrediction(history []models.DailyUsage, currentUsage *models.ProviderReport) *models.PredictionReport {
	if len(history) < 2 {
		return &models.PredictionReport{Confidence: "Low (Insufficient Data)"}
	}

	// 1. Calculate Weighted Daily Average (Last 7 days)
	// Weights: [1.5, 1.5, 1.2, 1.2, 1.2, 1.0, 1.0] (newest to oldest)
	weights := []float64{1.5, 1.5, 1.2, 1.2, 1.2, 1.0, 1.0}
	var weightedSum float64
	var weightSum float64

	for i := 0; i < len(history) && i < len(weights); i++ {
		weightedSum += history[i].IncludedRequests * weights[i]
		weightSum += weights[i]
	}

	weightedAvg := weightedSum / weightSum

	// 2. Weekend Compensation
	// Simple heuristic: If weekends are historically lower, we adjust the forecast for remaining days.
	weekdayAvg, weekendAvg := calculateDayTypeAverages(history)
	weekendRatio := 1.0
	if weekdayAvg > 0 {
		weekendRatio = weekendAvg / weekdayAvg
	}
	if weekendRatio < 0.1 {
		weekendRatio = 0.1 // Minimum fallback
	}

	// 3. Calculate Remaining Days in Month
	now := time.Now().UTC()
	currentDay := now.Day()
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	remainingDays := lastDay - currentDay

	// Calculate remaining weekdays vs weekends
	remainingWeekdays, remainingWeekends := countRemainingDays(now, remainingDays)

	// 4. Forecast
	projectedFutureUsage := (weightedAvg * float64(remainingWeekdays)) + (weightedAvg * weekendRatio * float64(remainingWeekends))

	// Total = current usage from report + projected future
	// Current total usage for the month is usedRequests
	currentTotal := float64(currentUsage.Entitlement - currentUsage.Remaining)
	predictedTotal := currentTotal + projectedFutureUsage

	// 5. Cost Prediction
	extraCost := 0.0
	if predictedTotal > float64(currentUsage.Entitlement) {
		overage := predictedTotal - float64(currentUsage.Entitlement)
		extraCost = overage * 0.04 // GitHub Copilot $0.04/request default
	}

	confidence := "High"
	if len(history) < 4 {
		confidence = "Medium"
	}
	if len(history) < 3 {
		confidence = "Low"
	}

	return &models.PredictionReport{
		PredictedMonthlyRequests: predictedTotal,
		PredictedExtraCost:       extraCost,
		Confidence:               confidence,
	}
}

func calculateDayTypeAverages(history []models.DailyUsage) (float64, float64) {
	var weekdaySum, weekendSum float64
	var weekdayCount, weekendCount float64

	for _, d := range history {
		t, err := time.Parse("2006-01-02", d.Date)
		if err != nil {
			continue
		}
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			weekendSum += d.IncludedRequests
			weekendCount++
		} else {
			weekdaySum += d.IncludedRequests
			weekdayCount++
		}
	}

	wdAvg := 0.0
	if weekdayCount > 0 {
		wdAvg = weekdaySum / weekdayCount
	}
	weAvg := 0.0
	if weekendCount > 0 {
		weAvg = weekendSum / weekendCount
	}

	return wdAvg, weAvg
}

func countRemainingDays(start time.Time, days int) (int, int) {
	weekdays, weekends := 0, 0
	for i := 1; i <= days; i++ {
		future := start.AddDate(0, 0, i)
		if future.Weekday() == time.Saturday || future.Weekday() == time.Sunday {
			weekends++
		} else {
			weekdays++
		}
	}
	return weekdays, weekends
}
