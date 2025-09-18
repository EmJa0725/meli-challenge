package llm

import (
	"context"
	"fmt"
	"os"
)

// LLMClient defines the behavior for any LLM provider (OpenAI, Gemini, etc.)
type LLMClient interface {
	// ClassifySample receives a data row sample and a list of classification categories.
	// It should return the matched InfoType or "N/A".
	ClassifySample(ctx context.Context, sample string, rules []string) (string, error)
}

// NewLLMClientFromEnv selects provider and initializes it from environment variables.
//   - LLM_PROVIDER=openai|gemini
//   - LLM_MODEL=model-name
func NewLLMClientFromEnv() (LLMClient, error) {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "openai" // default provider
	}

	switch provider {
	case "openai":
		return NewOpenAIClient(), nil
	// case "gemini":
	//	return NewGeminiClient(), nil
	default:
		return nil, fmt.Errorf("unsupported LLM_PROVIDER: %s", provider)
	}
}
