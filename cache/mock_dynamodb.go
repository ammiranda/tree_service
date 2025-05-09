package cache

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// MockDynamoDBClient implements DynamoDBAPI for testing
type MockDynamoDBClient struct {
	mu    sync.RWMutex
	items map[string]map[string]types.AttributeValue
}

// NewMockDynamoDBClient creates a new mock DynamoDB client
func NewMockDynamoDBClient() *MockDynamoDBClient {
	return &MockDynamoDBClient{
		items: make(map[string]map[string]types.AttributeValue),
	}
}

// CreateTable mocks the CreateTable operation
func (m *MockDynamoDBClient) CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tableName := *params.TableName
	if _, ok := m.items[tableName]; !ok {
		m.items[tableName] = make(map[string]types.AttributeValue)
	}
	return &dynamodb.CreateTableOutput{}, nil
}

// DescribeTable mocks the DescribeTable operation
func (m *MockDynamoDBClient) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return &dynamodb.DescribeTableOutput{}, nil
}

// GetItem mocks the GetItem operation
func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tableName := *params.TableName
	if items, ok := m.items[tableName]; ok {
		key := params.Key["key"].(*types.AttributeValueMemberS).Value
		if item, ok := items[key]; ok {
			// Return a properly formatted item
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"key":       &types.AttributeValueMemberS{Value: key},
					"data":      item,
					"ttl":       &types.AttributeValueMemberN{Value: "9999999999"},
					"timestamp": &types.AttributeValueMemberN{Value: "1"},
				},
			}, nil
		}
	}
	// Return empty response when item doesn't exist
	return &dynamodb.GetItemOutput{
		Item: nil,
	}, nil
}

// PutItem mocks the PutItem operation
func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tableName := *params.TableName
	if _, ok := m.items[tableName]; !ok {
		m.items[tableName] = make(map[string]types.AttributeValue)
	}

	// Store the data field
	key := params.Item["key"].(*types.AttributeValueMemberS).Value
	m.items[tableName][key] = params.Item["data"]

	return &dynamodb.PutItemOutput{}, nil
}

// DeleteItem mocks the DeleteItem operation
func (m *MockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tableName := *params.TableName
	key := params.Key["key"].(*types.AttributeValueMemberS).Value

	// Delete the item from the table
	if items, ok := m.items[tableName]; ok {
		delete(items, key)
		// If the table is empty, remove it
		if len(items) == 0 {
			delete(m.items, tableName)
		}
	}

	return &dynamodb.DeleteItemOutput{}, nil
}
