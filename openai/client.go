package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"meli-challenge/logger"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// AnalyzeSamples sends sample values to OpenAI and expects a single-label response.
func AnalyzeSamples(columnName string, samples []string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Warnf("OPENAI_API_KEY not set; skipping content-based classification")
		return "N/A", nil
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	system := `You are a concise data classifier. Given a column name and up to 5 sample values, respond with exactly one of the following labels (no extra text): FIRST_NAME, LAST_NAME, DATE_OF_BIRTH, GENDER, SSN, EMAIL_ADDRESS, PHONE_NUMBER, ADDRESS, POSTAL_CODE, USERNAME, PASSWORD, API_KEY, CREDIT_CARD_NUMBER, BANK_ACCOUNT, IP_ADDRESS, MAC_ADDRESS, HOSTNAME, N/A.`

	userBuilder := strings.Builder{}
	userBuilder.WriteString("Column: ")
	userBuilder.WriteString(columnName)
	userBuilder.WriteString("\nSamples:\n")
	for i, s := range samples {
		userBuilder.WriteString("- ")
		userBuilder.WriteString(strings.TrimSpace(s))
		if i < len(samples)-1 {
			userBuilder.WriteString("\n")
		}
	}

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: userBuilder.String()},
		},
		MaxTokens: 10,
	}

	b, _ := json.Marshal(reqBody)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		logger.Errorf("OpenAI API error: status=%d body=%s", resp.StatusCode, string(body))
		return "", errors.New("openai API error")
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "N/A", nil
	}

	// Normalize output
	out := strings.TrimSpace(cr.Choices[0].Message.Content)
	out = strings.ToUpper(out)
	// Keep only first token
	if idx := strings.IndexAny(out, "\n \t"); idx > 0 {
		out = out[:idx]
	}
	// Validate label roughly
	valid := map[string]bool{
		"FIRST_NAME": true, "LAST_NAME": true, "DATE_OF_BIRTH": true, "GENDER": true,
		"SSN": true, "EMAIL_ADDRESS": true, "PHONE_NUMBER": true, "ADDRESS": true,
		"POSTAL_CODE": true, "USERNAME": true, "PASSWORD": true, "API_KEY": true,
		"CREDIT_CARD_NUMBER": true, "BANK_ACCOUNT": true, "IP_ADDRESS": true, "MAC_ADDRESS": true,
		"HOSTNAME": true, "N/A": true,
	}
	if !valid[out] {
		// try to map some common responses
		if strings.Contains(out, "EMAIL") {
			return "EMAIL_ADDRESS", nil
		}
		if strings.Contains(out, "IP") {
			return "IP_ADDRESS", nil
		}
		return "N/A", nil
	}

	return out, nil
}
