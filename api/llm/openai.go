package llm

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"

	"meli-challenge/logger"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient loads credentials and model name from environment variables.
// TLS certificate verification is disabled (INSECURE).
func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY not set")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // default model
	}

	// Custom HTTP client with TLS verification disabled
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ⚠️ TODO: Configure properly for production
	}
	httpClient := &http.Client{Transport: insecureTransport}

	cfg := openai.DefaultConfig(apiKey)
	cfg.HTTPClient = httpClient

	return &OpenAIClient{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

// ClassifySample sends the row sample to OpenAI and asks it to classify based on rules.
func (c *OpenAIClient) ClassifySample(ctx context.Context, sample string, rules []string) (string, error) {
	prompt := "You are a strict data classifier. Given a row sample, identify if it contains any of these categories: "
	prompt += strings.Join(rules, ", ")
	prompt += ". If none apply, return 'N/A'.\nSample: " + sample
	// Log the prompt for debugging (note: may contain sensitive sample data)
	logger.Debugf("LLM prompt: %s", prompt)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: "You are a strict data classifier. Only respond with the matching category or 'N/A'."},
				{Role: "user", Content: prompt},
			},
		},
	)
	if err != nil {
		return "", err
	}

	// Log the raw LLM response for debugging/inspection
	if len(resp.Choices) > 0 {
		logger.Debugf("LLM response (raw): %s", resp.Choices[0].Message.Content)
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
