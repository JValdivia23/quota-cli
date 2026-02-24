package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

// VertexProvider implements the Vertex AI token fetcher via Google Cloud Monitoring
type VertexProvider struct{}

func (c *VertexProvider) Name() string {
	return "Vertex AI"
}

func (c *VertexProvider) Type() models.ProviderType {
	return models.TypeTokensBased
}

func (c *VertexProvider) Fetch(ctx context.Context, cfg *models.OpenCodeAuthConfig) (*models.ProviderReport, error) {
	// Discover standard cloud-platform credentials
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("could not find default GCP credentials: %w", err)
	}

	projectID := creds.ProjectID
	if projectID == "" {
		// Fallback to quota_project_id if using gcloud User Credentials
		var payload struct {
			QuotaProjectID string `json:"quota_project_id"`
		}
		if err := json.Unmarshal(creds.JSON, &payload); err == nil && payload.QuotaProjectID != "" {
			projectID = payload.QuotaProjectID
		}
	}

	if projectID == "" {
		return nil, fmt.Errorf("could not determine GCP Project ID from credentials")
	}

	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring client: %w", err)
	}
	defer client.Close()

	// Calculate start of current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Query for tokens used
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: `metric.type="aiplatform.googleapis.com/generate_content/total_token_count"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startOfMonth),
			EndTime:   timestamppb.New(now),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(2592000 * time.Second), // 30 days
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_SUM,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
		},
	}

	it := client.ListTimeSeries(ctx, req)
	var totalTokens int64
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Don't error out hard, vertex API might simply not have data or permissions could be restrictive
			break
		}
		for _, point := range resp.Points {
			totalTokens += point.GetValue().GetInt64Value()
		}
	}

	history, _ := fetchVertexHistory(ctx, client, projectID)

	return &models.ProviderReport{
		Name:        c.Name(),
		Type:        c.Type(),
		TokensUsed:  totalTokens,
		RefreshTime: "Monthly",
		History:     history,
	}, nil
}

func (c *VertexProvider) FetchHistory(ctx context.Context, cfg *models.OpenCodeAuthConfig) ([]models.DailyUsage, error) {
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	projectID := creds.ProjectID
	if projectID == "" {
		var payload struct {
			QuotaProjectID string `json:"quota_project_id"`
		}
		if err := json.Unmarshal(creds.JSON, &payload); err == nil && payload.QuotaProjectID != "" {
			projectID = payload.QuotaProjectID
		}
	}

	if projectID == "" {
		return nil, fmt.Errorf("no project ID found in credentials")
	}

	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return fetchVertexHistory(ctx, client, projectID)
}

func fetchVertexHistory(ctx context.Context, client *monitoring.MetricClient, projectID string) ([]models.DailyUsage, error) {
	now := time.Now()
	start := now.AddDate(0, 0, -7)

	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: `metric.type="aiplatform.googleapis.com/generate_content/total_token_count"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(start),
			EndTime:   timestamppb.New(now),
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(86400 * time.Second), // 1 day
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_SUM,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
		},
	}

	it := client.ListTimeSeries(ctx, req)

	// Create map of days
	dayMap := make(map[string]float64)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, point := range resp.Points {
			ts := point.GetInterval().GetEndTime().AsTime()
			dateStr := ts.Format("2006-01-02")
			dayMap[dateStr] += float64(point.GetValue().GetInt64Value())
		}
	}

	var history []models.DailyUsage
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		val := dayMap[d]
		history = append(history, models.DailyUsage{
			Date:             d,
			IncludedRequests: val, // We store token amount here to show activity
		})
	}
	return history, nil
}
