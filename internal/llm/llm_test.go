package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/oscarhfnorris/git-hunk/internal/llm"
)

// ---- mock client ----

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Summarise(ctx context.Context, req llm.SummaryRequest) (llm.SummaryResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(llm.SummaryResponse), args.Error(1)
}

func TestMockClient_Summarise(t *testing.T) {
	mc := new(MockClient)
	req := llm.SummaryRequest{HunkBody: "@@ -1 +1 @@ ...", File: "main.go"}
	want := llm.SummaryResponse{Summary: "test summary", Details: "some details"}

	mc.On("Summarise", mock.Anything, req).Return(want, nil)

	got, err := mc.Summarise(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	mc.AssertExpectations(t)
}

// ---- OpenAI implementation tests (fake HTTP server) ----

func fakeOpenAIServer(t *testing.T, responseBody interface{}, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(responseBody)
	}))
}

func TestOpenAIClient_Summarise_Success(t *testing.T) {
	innerJSON, _ := json.Marshal(map[string]string{
		"summary": "Add error handling",
		"details": "Returns early on validation failure.",
	})
	respBody := map[string]interface{}{
		"choices": []map[string]interface{}{
			{"message": map[string]string{"role": "assistant", "content": string(innerJSON)}},
		},
	}
	srv := fakeOpenAIServer(t, respBody, http.StatusOK)
	defer srv.Close()

	c := llm.NewOpenAI("sk-test", "gpt-4o-mini")
	c.BaseURL = srv.URL

	got, err := c.Summarise(context.Background(), llm.SummaryRequest{
		HunkBody: "@@ -1,3 +1,5 @@ ...",
		File:     "foo.go",
	})
	require.NoError(t, err)
	assert.Equal(t, "Add error handling", got.Summary)
	assert.Equal(t, "Returns early on validation failure.", got.Details)
}

func TestOpenAIClient_Summarise_APIError(t *testing.T) {
	respBody := map[string]interface{}{
		"error": map[string]string{"message": "invalid api key"},
	}
	srv := fakeOpenAIServer(t, respBody, http.StatusUnauthorized)
	defer srv.Close()

	c := llm.NewOpenAI("bad-key", "gpt-4o-mini")
	c.BaseURL = srv.URL

	_, err := c.Summarise(context.Background(), llm.SummaryRequest{HunkBody: "...", File: "f.go"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestOpenAIClient_Summarise_NoAPIKey(t *testing.T) {
	c := llm.NewOpenAI("", "gpt-4o-mini")
	_, err := c.Summarise(context.Background(), llm.SummaryRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestOpenAIClient_Summarise_NoChoices(t *testing.T) {
	respBody := map[string]interface{}{
		"choices": []interface{}{},
	}
	srv := fakeOpenAIServer(t, respBody, http.StatusOK)
	defer srv.Close()

	c := llm.NewOpenAI("sk-test", "gpt-4o-mini")
	c.BaseURL = srv.URL

	_, err := c.Summarise(context.Background(), llm.SummaryRequest{HunkBody: "...", File: "f.go"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestOpenAIClient_Summarise_PlainTextFallback(t *testing.T) {
	respBody := map[string]interface{}{
		"choices": []map[string]interface{}{
			{"message": map[string]string{"role": "assistant", "content": "plain text summary"}},
		},
	}
	srv := fakeOpenAIServer(t, respBody, http.StatusOK)
	defer srv.Close()

	c := llm.NewOpenAI("sk-test", "gpt-4o-mini")
	c.BaseURL = srv.URL

	got, err := c.Summarise(context.Background(), llm.SummaryRequest{HunkBody: "...", File: "f.go"})
	require.NoError(t, err)
	assert.Equal(t, "plain text summary", got.Summary)
}
