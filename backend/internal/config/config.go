// Package config loads runtime configuration from environment variables and,
// for secrets, from SSM Parameter Store (SecureString) with in-process caching.
package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Config holds all values needed by the application at runtime.
type Config struct {
	DynamoTableName    string
	AllowedCORSOrigins []string

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	JWTSigningKey      string

	OAuthRedirectBaseURL string
}

// SSMClient is the subset of the SSM API the config package depends on.
type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

var (
	loadOnce   sync.Once
	cachedCfg  *Config
	cachedErr  error
	newSSMFunc = newSSMClient
)

// Load returns the process-wide configuration, fetching SSM secrets once per
// cold start (Lambda execution environments reuse this across invocations).
func Load(ctx context.Context) (*Config, error) {
	loadOnce.Do(func() {
		cachedCfg, cachedErr = load(ctx)
	})
	return cachedCfg, cachedErr
}

func load(ctx context.Context) (*Config, error) {
	cfg := &Config{
		DynamoTableName:      mustEnv("DYNAMODB_TABLE_NAME"),
		AllowedCORSOrigins:   splitCSV(os.Getenv("ALLOWED_CORS_ORIGINS")),
		OAuthRedirectBaseURL: mustEnv("OAUTH_REDIRECT_BASE_URL"),
	}

	prefix := os.Getenv("OAUTH_SSM_PREFIX")
	if prefix == "" {
		// No SSM prefix configured: fall back to plain env vars (useful for local/dev).
		cfg.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
		cfg.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
		cfg.GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
		cfg.GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
		cfg.JWTSigningKey = os.Getenv("JWT_SIGNING_KEY")
		return cfg, nil
	}

	client, err := newSSMFunc(ctx)
	if err != nil {
		return nil, fmt.Errorf("config: creating ssm client: %w", err)
	}

	params := map[string]*string{
		prefix + "/google_client_id":     &cfg.GoogleClientID,
		prefix + "/google_client_secret": &cfg.GoogleClientSecret,
		prefix + "/github_client_id":     &cfg.GitHubClientID,
		prefix + "/github_client_secret": &cfg.GitHubClientSecret,
		prefix + "/jwt_signing_key":      &cfg.JWTSigningKey,
	}
	for name, dest := range params {
		value, err := getParameter(ctx, client, name)
		if err != nil {
			return nil, fmt.Errorf("config: fetching parameter %s: %w", name, err)
		}
		*dest = value
	}

	return cfg, nil
}

func getParameter(ctx context.Context, client SSMClient, name string) (string, error) {
	out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.Parameter.Value), nil
}

func newSSMClient(ctx context.Context) (SSMClient, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return ssm.NewFromConfig(awsCfg), nil
}

func mustEnv(key string) string {
	return os.Getenv(key)
}

func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
