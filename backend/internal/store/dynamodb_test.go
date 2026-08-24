package store

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDynamoDBClient struct {
	mock.Mock
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*dynamodb.GetItemOutput)
	return out, args.Error(1)
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	out, _ := args.Get(0).(*dynamodb.PutItemOutput)
	return out, args.Error(1)
}

func TestDynamoDBUserRepository_Get_NotFound(t *testing.T) {
	client := &mockDynamoDBClient{}
	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

	repo := NewDynamoDBUserRepository(client, "users")
	_, err := repo.Get(context.Background(), "google", "1")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDynamoDBUserRepository_Get_Found(t *testing.T) {
	client := &mockDynamoDBClient{}
	item, err := attributevalue.MarshalMap(userItem{
		PK: "USER#google#1", SK: "PROFILE", Provider: "google", ProviderID: "1", Email: "a@example.com", Name: "Ada",
	})
	require.NoError(t, err)
	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	repo := NewDynamoDBUserRepository(client, "users")
	user, err := repo.Get(context.Background(), "google", "1")
	require.NoError(t, err)
	require.Equal(t, "a@example.com", user.Email)
}

func TestDynamoDBUserRepository_Upsert_NewUserSetsCreatedAt(t *testing.T) {
	client := &mockDynamoDBClient{}
	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)
	client.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

	repo := NewDynamoDBUserRepository(client, "users")
	user, err := repo.Upsert(context.Background(), User{Provider: "google", ProviderID: "1", Email: "a@example.com"})
	require.NoError(t, err)
	require.False(t, user.CreatedAt.IsZero())

	client.AssertExpectations(t)
}
