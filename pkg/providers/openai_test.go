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

	// With invalid/mock token, we expect a real API error (no mock data)
	_, err := provider.Fetch(context.Background(), &models.OpenCodeAuthConfig{
		RawKeys: map[string]interface{}{
			"openai": map[string]interface{}{
				"access": "mock-key",
			},
		},
	})
	if err == nil {
		t.Fatal("Expected an error for an invalid token, but got nil (mock data may still be active)")
	}
	// error is expected and acceptable — provider is honest
}
