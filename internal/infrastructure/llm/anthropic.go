package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	anthropicBaseURL    = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	anthropicModelSonnet = "claude-sonnet-4-6"
	anthropicModelHaiku  = "claude-haiku-4-5-20251001"
)

// AnthropicProvider implements LLMProvider for Anthropic Claude.
type AnthropicProvider struct {
	apiKey  string
	model   string
	client  *http.Client
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = anthropicModelSonnet
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string  { return "anthropic" }
func (p *AnthropicProvider) Model() string { return p.model }

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

func (p *AnthropicProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	start := time.Now()
	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: req.MaxTokens,
		System:    req.SystemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: req.UserPrompt}},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicBaseURL, bytes.NewReader(b))
	if err != nil {
		return LLMResponse{}, err
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return LLMResponse{}, fmt.Errorf("anthropic: status %d: %s", resp.StatusCode, string(body))
	}

	var ar anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return LLMResponse{}, fmt.Errorf("anthropic: decode response: %w", err)
	}

	content := ""
	if len(ar.Content) > 0 {
		content = ar.Content[0].Text
	}

	return LLMResponse{
		Content: content,
		Usage: TokenUsage{
			PromptTokens:     ar.Usage.InputTokens,
			CompletionTokens: ar.Usage.OutputTokens,
			TotalTokens:      ar.Usage.InputTokens + ar.Usage.OutputTokens,
		},
		Model:     ar.Model,
		LatencyMs: time.Since(start).Milliseconds(),
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req LLMRequest) (<-chan LLMChunk, error) {
	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: req.MaxTokens,
		System:    req.SystemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: req.UserPrompt}},
		Stream:    true,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicBaseURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic: stream status %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan LLMChunk, 32)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- LLMChunk{Done: true}
				return
			}

			var event struct {
				Type  string `json:"type"`
				Delta *struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			if event.Type == "content_block_delta" && event.Delta != nil && event.Delta.Type == "text_delta" {
				ch <- LLMChunk{Delta: event.Delta.Text}
			}
			if event.Type == "message_stop" {
				ch <- LLMChunk{Done: true}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- LLMChunk{Error: fmt.Errorf("anthropic: stream read error: %w", err)}
		}
	}()

	return ch, nil
}
