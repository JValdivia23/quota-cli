package providers

import (
	"context"
	"testing"

	"github.com/JValdivia23/quota-cli/pkg/models"
)

func TestOpenAIProvider_Fetch(t *testing.T) {
	provider := &OpenAIProvider{}

	if provider.Name() != "OpenAI" {
		t.Errorf("Expected OpenAI, got %s", provider.Name())
	}

	if provider.Type() != models.TypeQuotaBased {
		t.Errorf("Expected TypeQuotaBased, got %v", provider.Type())
	}

	report, err := provider.Fetch(context.Background(), &models.OpenCodeAuthConfig{
		RawKeys: map[string]interface{}{
			"openai": map[string]interface{}{
				"access": "mock-key",
			},
		},
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if report.Name != "OpenAI" {
		t.Errorf("Expected report name OpenAI, got %s", report.Name)
	}

	if report.Entitlement != 100 {
		t.Errorf("Expected entitlement 100, got %d", report.Entitlement)
	}
}
