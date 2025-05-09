package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ammiranda/tree_service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBAPI defines the interface for DynamoDB operations
type DynamoDBAPI interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// DynamoDBCache implements CacheProvider using DynamoDB
type DynamoDBCache struct {
	client   DynamoDBAPI
	cacheTTL time.Duration
}

// NewDynamoDBCache creates a new DynamoDB cache provider
func NewDynamoDBCache() (*DynamoDBCache, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &DynamoDBCache{
		client:   dynamodb.NewFromConfig(cfg),
		cacheTTL: 5 * time.Minute,
	}, nil
}

// NewDynamoDBCacheWithClient creates a new DynamoDB cache provider with a custom client
func NewDynamoDBCacheWithClient(client DynamoDBAPI) *DynamoDBCache {
	return &DynamoDBCache{
		client:   client,
		cacheTTL: 5 * time.Minute,
	}
}

// Initialize creates the DynamoDB table if it doesn't exist
func (c *DynamoDBCache) Initialize() error {
	ctx := context.TODO()

	// Check if table exists
	_, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		// Table exists
		return nil
	}

	// Create table
	_, err = c.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("key"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	return err
}

// GetTree retrieves the tree from DynamoDB cache if available
func (c *DynamoDBCache) GetTree() ([]*models.Node, bool) {
	ctx := context.TODO()

	// Get item from DynamoDB
	result, err := c.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: cacheKey},
		},
	})
	if err != nil {
		return nil, false
	}

	if result.Item == nil {
		return nil, false
	}

	var item CacheItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, false
	}

	// Check if cache is still valid
	now := time.Now().Unix()
	if now > item.TTL {
		// Cache expired, delete it
		if _, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"key": &types.AttributeValueMemberS{Value: cacheKey},
			},
		}); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Error deleting expired cache item: %v\n", err)
		}
		return nil, false
	}

	return item.Data, true
}

// SetTree stores the tree in DynamoDB cache
func (c *DynamoDBCache) SetTree(tree []*models.Node) {
	ctx := context.TODO()
	now := time.Now()
	ttl := now.Add(c.cacheTTL).Unix()

	item := CacheItem{
		Key:       cacheKey,
		Data:      tree,
		Timestamp: now.Unix(),
		TTL:       ttl,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		// If we can't marshal the item, invalidate the cache
		if err := c.InvalidateCache(); err != nil {
			fmt.Printf("Warning: Error invalidating cache after marshal failure: %v\n", err)
		}
		return
	}

	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		// If we can't store the item, invalidate the cache
		if err := c.InvalidateCache(); err != nil {
			fmt.Printf("Warning: Error invalidating cache after put failure: %v\n", err)
		}
		return
	}
}

// InvalidateCache removes the tree from DynamoDB cache
func (c *DynamoDBCache) InvalidateCache() error {
	ctx := context.Background()
	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: cacheKey},
		},
	})
	return err
}

// SetCacheTTL sets the cache time-to-live duration
func (c *DynamoDBCache) SetCacheTTL(ttl time.Duration) {
	c.cacheTTL = ttl
}

const (
	tableName = "TreeCache"
	cacheKey  = "tree"
)

type CacheItem struct {
	Key       string         `dynamodbav:"key"`
	Data      []*models.Node `dynamodbav:"data"`
	Timestamp int64          `dynamodbav:"timestamp"`
	TTL       int64          `dynamodbav:"ttl"`
}
