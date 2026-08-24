package config

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/require"
)

type fakeSSMClient struct {
	values map[string]string
}

func (f *fakeSSMClient) GetParameter(_ context.Context, params *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	name := aws.ToString(params.Name)
	return &ssm.GetParameterOutput{
		Parameter: &ssmtypes.Parameter{Value: aws.String(f.values[name])},
	}, nil
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() { _ = os.Unsetenv(key) })
}

func TestLoad_FallsBackToPlainEnvVarsWithoutSSMPrefix(t *testing.T) {
	setEnv(t, "DYNAMODB_TABLE_NAME", "users")
	setEnv(t, "OAUTH_REDIRECT_BASE_URL", "https://api.example.com")
	setEnv(t, "ALLOWED_CORS_ORIGINS", "https://a.example.com, https://b.example.com")
	setEnv(t, "GOOGLE_CLIENT_ID", "gid")
	setEnv(t, "JWT_SIGNING_KEY", "secret")

	cfg, err := load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "users", cfg.DynamoTableName)
	require.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, cfg.AllowedCORSOrigins)
	require.Equal(t, "gid", cfg.GoogleClientID)
	require.Equal(t, "secret", cfg.JWTSigningKey)
}

func TestLoad_ReadsSecretsFromSSMWhenPrefixSet(t *testing.T) {
	setEnv(t, "DYNAMODB_TABLE_NAME", "users")
	setEnv(t, "OAUTH_REDIRECT_BASE_URL", "https://api.example.com")
	setEnv(t, "OAUTH_SSM_PREFIX", "/evap/prod")

	original := newSSMFunc
	newSSMFunc = func(ctx context.Context) (SSMClient, error) {
		return &fakeSSMClient{values: map[string]string{
			"/evap/prod/google_client_id":     "gid",
			"/evap/prod/google_client_secret": "gsecret",
			"/evap/prod/github_client_id":     "hid",
			"/evap/prod/github_client_secret": "hsecret",
			"/evap/prod/jwt_signing_key":      "jwt-secret",
		}}, nil
	}
	t.Cleanup(func() { newSSMFunc = original })

	cfg, err := load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "gid", cfg.GoogleClientID)
	require.Equal(t, "gsecret", cfg.GoogleClientSecret)
	require.Equal(t, "jwt-secret", cfg.JWTSigningKey)
}
