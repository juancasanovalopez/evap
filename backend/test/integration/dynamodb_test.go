//go:build integration

// Package integration runs the store package against a real DynamoDB Local
// instance (see docker-compose/GitHub Actions service in the README).
package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"

	"evap-backend/internal/store"
)

func newTestClient(t *testing.T) *dynamodb.Client {
	t.Helper()
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	require.NoError(t, err)

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func ensureTable(t *testing.T, client *dynamodb.Client, tableName string) {
	t.Helper()
	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
	})
	if err != nil {
		var inUse *types.ResourceInUseException
		if !errors.As(err, &inUse) {
			require.NoError(t, err)
		}
	}
}

func TestDynamoDBUserRepository_UpsertAndGet_AgainstDynamoDBLocal(t *testing.T) {
	tableName := "evap_users_integration"
	client := newTestClient(t)
	ensureTable(t, client, tableName)

	repo := store.NewDynamoDBUserRepository(client, tableName)
	ctx := context.Background()

	created, err := repo.Upsert(ctx, store.User{
		Provider:   "google",
		ProviderID: time.Now().Format("150405.000000000"),
		Email:      "integration@example.com",
		Name:       "Integration Test",
	})
	require.NoError(t, err)

	got, err := repo.Get(ctx, created.Provider, created.ProviderID)
	require.NoError(t, err)
	require.Equal(t, created.Email, got.Email)
	require.Equal(t, created.Name, got.Name)
}

func TestDynamoDBUserRepository_Get_NotFound_AgainstDynamoDBLocal(t *testing.T) {
	tableName := "evap_users_integration"
	client := newTestClient(t)
	ensureTable(t, client, tableName)

	repo := store.NewDynamoDBUserRepository(client, tableName)
	_, err := repo.Get(context.Background(), "google", "does-not-exist")
	require.ErrorIs(t, err, store.ErrNotFound)
}
