package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient is the subset of the DynamoDB API the repository depends on.
type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// DynamoDBUserRepository implements UserRepository against a single-table
// DynamoDB design: PK="USER#<provider>#<providerID>", SK="PROFILE".
type DynamoDBUserRepository struct {
	client    DynamoDBClient
	tableName string
	now       func() time.Time
}

// NewDynamoDBUserRepository builds a repository backed by the given table.
func NewDynamoDBUserRepository(client DynamoDBClient, tableName string) *DynamoDBUserRepository {
	return &DynamoDBUserRepository{client: client, tableName: tableName, now: time.Now}
}

type userItem struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	Provider   string    `dynamodbav:"provider"`
	ProviderID string    `dynamodbav:"provider_id"`
	Email      string    `dynamodbav:"email"`
	Name       string    `dynamodbav:"name"`
	AvatarURL  string    `dynamodbav:"avatar_url"`
	CreatedAt  time.Time `dynamodbav:"created_at"`
	UpdatedAt  time.Time `dynamodbav:"updated_at"`
}

func userKey(provider, providerID string) (string, string) {
	return fmt.Sprintf("USER#%s#%s", provider, providerID), "PROFILE"
}

// Get retrieves a user profile by provider and provider-specific ID.
func (r *DynamoDBUserRepository) Get(ctx context.Context, provider, providerID string) (User, error) {
	pk, sk := userKey(provider, providerID)
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return User{}, fmt.Errorf("dynamodb get item: %w", err)
	}
	if out.Item == nil {
		return User{}, ErrNotFound
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return User{}, fmt.Errorf("dynamodb unmarshal item: %w", err)
	}
	return itemToUser(item), nil
}

// Upsert creates or updates a user profile, preserving the original CreatedAt.
func (r *DynamoDBUserRepository) Upsert(ctx context.Context, u User) (User, error) {
	existing, err := r.Get(ctx, u.Provider, u.ProviderID)
	now := r.now().UTC()
	switch {
	case err == nil:
		u.CreatedAt = existing.CreatedAt
	case errors.Is(err, ErrNotFound):
		u.CreatedAt = now
	default:
		return User{}, err
	}
	u.UpdatedAt = now

	pk, sk := userKey(u.Provider, u.ProviderID)
	item := userItem{
		PK:         pk,
		SK:         sk,
		Provider:   u.Provider,
		ProviderID: u.ProviderID,
		Email:      u.Email,
		Name:       u.Name,
		AvatarURL:  u.AvatarURL,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return User{}, fmt.Errorf("dynamodb marshal item: %w", err)
	}

	if _, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}); err != nil {
		return User{}, fmt.Errorf("dynamodb put item: %w", err)
	}

	return u, nil
}

func itemToUser(item userItem) User {
	return User{
		Provider:   item.Provider,
		ProviderID: item.ProviderID,
		Email:      item.Email,
		Name:       item.Name,
		AvatarURL:  item.AvatarURL,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}

// DynamoDBQueryClient is the subset of the DynamoDB API needed to query readings.
type DynamoDBQueryClient interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDBReadingRepository implements ReadingRepository against the
// sensor_readings table, querying the owner_user_id-timestamp GSI so a user
// can only ever see their own readings.
type DynamoDBReadingRepository struct {
	client    DynamoDBQueryClient
	tableName string
	indexName string
}

// NewDynamoDBReadingRepository builds a repository backed by the given table and GSI.
func NewDynamoDBReadingRepository(client DynamoDBQueryClient, tableName, indexName string) *DynamoDBReadingRepository {
	return &DynamoDBReadingRepository{client: client, tableName: tableName, indexName: indexName}
}

type readingItem struct {
	DeviceID    string  `dynamodbav:"device_id"`
	OwnerUserID string  `dynamodbav:"owner_user_id"`
	Timestamp   string  `dynamodbav:"timestamp"`
	Temperature float64 `dynamodbav:"temperature"`
	IngestedAt  string  `dynamodbav:"ingested_at"`
}

// ListRecentByOwner returns up to limit readings for ownerUserID, most recent first.
func (r *DynamoDBReadingRepository) ListRecentByOwner(ctx context.Context, ownerUserID string, limit int32) ([]Reading, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(r.indexName),
		KeyConditionExpression: aws.String("owner_user_id = :owner"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":owner": &types.AttributeValueMemberS{Value: ownerUserID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb query readings: %w", err)
	}

	items := make([]readingItem, 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
		return nil, fmt.Errorf("dynamodb unmarshal readings: %w", err)
	}

	readings := make([]Reading, 0, len(items))
	for _, item := range items {
		readings = append(readings, Reading{
			DeviceID:    item.DeviceID,
			OwnerUserID: item.OwnerUserID,
			Timestamp:   item.Timestamp,
			Temperature: item.Temperature,
			IngestedAt:  item.IngestedAt,
		})
	}
	return readings, nil
}
