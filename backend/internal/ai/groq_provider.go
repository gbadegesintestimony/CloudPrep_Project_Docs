package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloudprep-backend/internal/models"
)

type GroqProvider struct {
	APIKey  string
	Model   string
	Client  *http.Client
	BaseURL string
}

func NewGroqProvider(apiKey, modelID string) *GroqProvider {
	if modelID == "" {
		modelID = "openai/gpt-oss-120b"
	}
	return &GroqProvider{
		APIKey:  apiKey,
		Model:   modelID,
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: "https://api.groq.com/openai/v1",
	}
}

func (p *GroqProvider) ProviderName() string {
	return "groq"
}

func (p *GroqProvider) ModelID() string {
	return p.Model
}

func (p *GroqProvider) IsConfigured() bool {
	return strings.TrimSpace(p.APIKey) != ""
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqRequestBody struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
	Temperature    float64             `json:"temperature,omitempty"`
}

type groqResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (p *GroqProvider) GenerateQuestions(
	ctx context.Context,
	req GenerationRequest,
	topic models.CurriculumTopic,
	excerpts []string,
) (*GenerationResult, error) {
	if !p.IsConfigured() {
		return nil, ErrUnconfigured
	}

	start := time.Now()
	systemPrompt := BuildSystemPrompt()
	userPrompt := BuildUserPrompt(req, topic, excerpts)

	reqBody := groqRequestBody{
		Model: p.Model,
		Messages: []groqMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
		Temperature:    0.4,
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode groq request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create groq request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return nil, fmt.Errorf("%w: %v", ErrProviderOffline, sanitizeError(err.Error(), p.APIKey))
		}
		return nil, errors.New(sanitizeError(err.Error(), p.APIKey))
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading groq response: %w", err)
	}

	durationMs := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("%w: groq status %d", ErrRateLimited, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var groqResp groqResponseBody
	if err := json.Unmarshal(bodyBytes, &groqResp); err != nil {
		return nil, fmt.Errorf("failed parsing groq wrapper JSON: %w", err)
	}

	if groqResp.Error != nil {
		return nil, fmt.Errorf("groq API returned error: %s (code %s)", groqResp.Error.Message, groqResp.Error.Code)
	}

	if len(groqResp.Choices) == 0 {
		return nil, errors.New("groq returned 0 choices")
	}

	text := groqResp.Choices[0].Message.Content
	cleanedText := cleanJSONFences(text)

	var parsed struct {
		Questions []GeneratedQuestion `json:"questions"`
	}
	if err := json.Unmarshal([]byte(cleanedText), &parsed); err != nil {
		return nil, fmt.Errorf("failed parsing generated questions JSON from groq: %w (raw: %s)", err, cleanedText)
	}

	return &GenerationResult{
		Provider:     p.ProviderName(),
		ModelID:      p.ModelID(),
		FailoverUsed: false,
		Questions:    parsed.Questions,
		DurationMs:   durationMs,
		RawResponse:  cleanedText,
	}, nil
}

