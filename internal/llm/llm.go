// Package llm provides a provider-agnostic LLM client interface with an
// OpenAI implementation.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SummaryRequest is sent to the LLM to summarise a hunk.
type SummaryRequest struct {
	// HunkBody is the raw diff hunk body.
	HunkBody string
	// File is the name of the file being modified.
	File string
}

// SummaryResponse holds the result of a summary call.
type SummaryResponse struct {
	// Summary is a short (≤80 chars) one-line summary.
	Summary string
	// Details is a longer Markdown explanation.
	Details string
}

// Client is the provider-agnostic interface for generating hunk summaries.
type Client interface {
	Summarise(ctx context.Context, req SummaryRequest) (SummaryResponse, error)
}

// ---- OpenAI implementation --------------------------------------------------

const openAIEndpoint = "https://api.openai.com/v1/chat/completions"

// OpenAIClient implements Client using the OpenAI chat completions API.
type OpenAIClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
	BaseURL    string // overrideable for testing
}

// NewOpenAI creates a new OpenAIClient with sensible defaults.
func NewOpenAI(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		APIKey:     apiKey,
		Model:      model,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    openAIEndpoint,
	}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
}

// Summarise calls the OpenAI chat completions endpoint and extracts a
// summary + details for the given hunk.
func (c *OpenAIClient) Summarise(ctx context.Context, req SummaryRequest) (SummaryResponse, error) {
	if c.APIKey == "" {
		return SummaryResponse{}, errors.New("llm: API key is required")
	}

	prompt := fmt.Sprintf(`You are a code reviewer. Summarise the following diff hunk from file %q.
Respond with JSON: {"summary":"<one-line summary ≤80 chars>","details":"<markdown explanation>"}

Hunk:
%s`, req.File, req.HunkBody)

	body, err := json.Marshal(openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You summarise git diff hunks concisely."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return SummaryResponse{}, fmt.Errorf("llm: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return SummaryResponse{}, fmt.Errorf("llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return SummaryResponse{}, fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return SummaryResponse{}, fmt.Errorf("llm: read response: %w", err)
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return SummaryResponse{}, fmt.Errorf("llm: decode response: %w", err)
	}

	if apiResp.Error != nil {
		return SummaryResponse{}, fmt.Errorf("llm: api error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return SummaryResponse{}, errors.New("llm: no choices in response")
	}

	var result SummaryResponse
	content := apiResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fall back to treating the content as plain text summary.
		result.Summary = content
	}
	return result, nil
}
