package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"careergps/pkg/hash"
)

// LLMProvider is the port for all LLM interactions.
// Implementations live in infrastructure/llm/{anthropic,openai,gemini}.
type LLMProvider interface {
	Generate(ctx context.Context, req LLMRequest) (LLMResponse, error)
	Stream(ctx context.Context, req LLMRequest) (<-chan LLMChunk, error)
	Name() string
	Model() string
}

// JSONSchema enforces structured output for non-streaming calls.
type JSONSchema struct {
	Schema json.RawMessage
}

// LLMRequest is the unified input to any LLM provider.
type LLMRequest struct {
	SystemPrompt   string
	UserPrompt     string
	OutputSchema   *JSONSchema // nil for streaming or free-form
	MaxTokens      int
	Temperature    float32
	IdempotencyKey string // used as cache key prefix
}

// CacheKey produces a deterministic cache key for this request.
func (r LLMRequest) CacheKey(model, engineVersion string) string {
	parts := []string{r.SystemPrompt, r.UserPrompt, model, engineVersion}
	if r.OutputSchema != nil {
		parts = append(parts, string(r.OutputSchema.Schema))
	}
	return hash.SHA256Multi(parts...)
}

// TokenUsage tracks prompt and completion token counts.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMResponse is the unified output from any LLM provider.
type LLMResponse struct {
	Content   string
	Usage     TokenUsage
	Model     string
	LatencyMs int64
	CacheHit  bool
}

// LLMChunk is a single streaming delta from the LLM.
type LLMChunk struct {
	Delta string
	Done  bool
	Error error
}

// Cache is the interface for caching LLM responses.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
}

// CachingProvider wraps any LLMProvider with response caching.
// Only caches Generate (non-streaming) calls.
type CachingProvider struct {
	inner         LLMProvider
	cache         Cache
	ttl           time.Duration
	engineVersion string
}

func NewCachingProvider(inner LLMProvider, cache Cache, ttl time.Duration, engineVersion string) *CachingProvider {
	return &CachingProvider{inner: inner, cache: cache, ttl: ttl, engineVersion: engineVersion}
}

func (p *CachingProvider) Name() string  { return p.inner.Name() }
func (p *CachingProvider) Model() string { return p.inner.Model() }

func (p *CachingProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	key := req.CacheKey(p.inner.Model(), p.engineVersion)
	if cached, ok := p.cache.Get(ctx, key); ok {
		var resp LLMResponse
		if err := json.Unmarshal(cached, &resp); err == nil {
			resp.CacheHit = true
			return resp, nil
		}
	}

	resp, err := p.inner.Generate(ctx, req)
	if err != nil {
		return resp, err
	}

	if b, err := json.Marshal(resp); err == nil {
		p.cache.Set(ctx, key, b, p.ttl)
	}
	return resp, nil
}

// Stream is never cached — conversational context changes every turn.
func (p *CachingProvider) Stream(ctx context.Context, req LLMRequest) (<-chan LLMChunk, error) {
	return p.inner.Stream(ctx, req)
}

// FallbackProvider chains providers: primary → fallbacks in order.
// For streaming, only the primary is attempted — fallback on mid-stream error
// would require complex state reconstruction, deferred to V2.
type FallbackProvider struct {
	providers []LLMProvider
}

func NewFallbackProvider(providers ...LLMProvider) *FallbackProvider {
	if len(providers) == 0 {
		panic("llm: FallbackProvider requires at least one provider")
	}
	return &FallbackProvider{providers: providers}
}

func (f *FallbackProvider) Name() string  { return f.providers[0].Name() }
func (f *FallbackProvider) Model() string { return f.providers[0].Model() }

func (f *FallbackProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	var lastErr error
	for _, p := range f.providers {
		resp, err := p.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return LLMResponse{}, fmt.Errorf("all LLM providers failed: %w", lastErr)
}

func (f *FallbackProvider) Stream(ctx context.Context, req LLMRequest) (<-chan LLMChunk, error) {
	return f.providers[0].Stream(ctx, req)
}
