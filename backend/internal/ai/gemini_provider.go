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

var (
	ErrRateLimited     = errors.New("provider rate limit exceeded (HTTP 429)")
	ErrProviderOffline = errors.New("provider service unavailable or timed out")
	ErrUnconfigured    = errors.New("provider API key is not configured")
)

type GeminiProvider struct {
	APIKey  string
	Model   string
	Client  *http.Client
	BaseURL string
}

func NewGeminiProvider(apiKey, modelID string) *GeminiProvider {
	if modelID == "" {
		modelID = "gemini-3.8-flash"
	}
	return &GeminiProvider{
		APIKey:  apiKey,
		Model:   modelID,
		Client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
	}
}

func (p *GeminiProvider) ProviderName() string {
	return "gemini"
}

func (p *GeminiProvider) ModelID() string {
	return p.Model
}

func (p *GeminiProvider) IsConfigured() bool {
	return strings.TrimSpace(p.APIKey) != ""
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenConfig struct {
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
}

type geminiRequestBody struct {
	SystemInstruction *geminiContent   `json:"system_instruction,omitempty"`
	Contents          []geminiContent  `json:"contents"`
	GenerationConfig  *geminiGenConfig `json:"generationConfig,omitempty"`
}

type geminiResponseBody struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) GenerateQuestions(
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

	reqBody := geminiRequestBody{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userPrompt}},
			},
		},
		GenerationConfig: &geminiGenConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.4,
		},
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode gemini request: %w", err)
	}

	// Use standard REST endpoint without embedding API keys in query parameters (Hardened V-1)
	url := fmt.Sprintf("%s/models/%s:generateContent", p.BaseURL, p.Model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", p.APIKey)

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
		return nil, fmt.Errorf("failed reading gemini response: %w", err)
	}

	durationMs := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("%w: gemini status %d", ErrRateLimited, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponseBody
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed parsing gemini wrapper JSON: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini API returned error: %s (code %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gemini returned 0 candidates")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
	cleanedText := cleanJSONFences(text)

	var parsed struct {
		Questions []GeneratedQuestion `json:"questions"`
	}
	if err := json.Unmarshal([]byte(cleanedText), &parsed); err != nil {
		return nil, fmt.Errorf("failed parsing generated questions JSON: %w (raw: %s)", err, cleanedText)
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

func cleanJSONFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
	} else if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
	}
	if strings.HasSuffix(trimmed, "```") {
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}

func sanitizeError(msg, secret string) string {
	if secret != "" && len(secret) > 6 {
		msg = strings.ReplaceAll(msg, secret, "[REDACTED_API_KEY]")
	}
	return msg
}

