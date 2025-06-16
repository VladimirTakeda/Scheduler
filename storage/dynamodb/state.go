package dynamodb

import (
	"InterviewScheduler/storage"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

const UserStateTableName = "SchedulerBotUserStates"

// We save user date with local timezone, so we can use it later
func (ds *DynamoStorage) SaveUserState(userState *storage.UserState) error {
	item, err := attributevalue.MarshalMap(userState)
	if err != nil {
		return fmt.Errorf("failed to marshal user state: %v", err)
	}

	_, err = ds.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(UserStateTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save user state to DynamoDB: %v", err)
	}

	return nil
}

func (ds *DynamoStorage) ClearUserState(userID int64) error {
	_, err := ds.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(UserStateTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: strconv.FormatInt(userID, 10)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user state from DynamoDB: %v", err)
	}

	return nil
}

func (ds *DynamoStorage) GetUserState(userID int64) (*storage.UserState, error) {
	result, err := ds.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(UserStateTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: strconv.FormatInt(userID, 10)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user state from DynamoDB: %v", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var userState storage.UserState
	err = attributevalue.UnmarshalMap(result.Item, &userState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user state: %v", err)
	}

	return &userState, nil
}
