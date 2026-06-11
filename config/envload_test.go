package config

import (
	"testing"
)

// Simulates production (Render): no .env file in cwd, config comes purely
// from environment variables.
func TestLoadFromEnvOnly(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("DATABASE_URL", "postgres://prod-host/db")
	t.Setenv("REDIS_ADDR", "prod-redis:6379")
	t.Setenv("JWT_PRIVATE_KEY_B64", "dGVzdA==")
	t.Setenv("JWT_PUBLIC_KEY_B64", "dGVzdA==")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("S3_BUCKET", "careergps-resumes")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	checks := map[string]string{
		"DATABASE_URL":        cfg.DatabaseURL,
		"REDIS_ADDR":          cfg.RedisAddr,
		"JWT_PRIVATE_KEY_B64": cfg.JWTPrivateKeyB64,
		"JWT_PUBLIC_KEY_B64":  cfg.JWTPublicKeyB64,
		"ANTHROPIC_API_KEY":   cfg.AnthropicAPIKey,
		"S3_BUCKET":           cfg.S3Bucket,
	}
	for name, got := range checks {
		if got == "" {
			t.Errorf("%s was set in env but unmarshalled empty", name)
		}
	}
}
