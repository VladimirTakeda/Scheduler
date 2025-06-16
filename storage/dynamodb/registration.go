package dynamodb

import (
	"InterviewScheduler/storage"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

const UserTableName = "SchedulerBotUsers"

func (ds *DynamoStorage) SaveUser(userID int64, firstName, username string, timezone string) error {
	user := storage.User{
		UserID:       fmt.Sprintf("%d", userID),
		FirstName:    firstName,
		Username:     username,
		RegisteredAt: time.Now().Format(time.RFC3339),
		Timezone:     timezone,
	}

	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %v", err)
	}

	_, err = ds.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(UserTableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %v", err)
	}

	log.Printf("Saved user %s to DynamoDB", user.UserID)
	return nil
}

func (ds *DynamoStorage) IsUserRegistered(chatID int64) (bool, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(UserTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", chatID)},
		},
	}

	out, err := ds.client.GetItem(context.TODO(), input)
	if err != nil {
		return false, err
	}
	return out.Item != nil && len(out.Item) > 0, nil
}

// SaveUserTimezone сохраняет выбранный пользователем часовой пояс в профиле пользователя
func (ds *DynamoStorage) SaveUserTimezone(userID int64, timezone string) error {
	key := map[string]types.AttributeValue{
		"user_id": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", userID)},
	}
	update := &dynamodb.UpdateItemInput{
		TableName:        aws.String(UserTableName),
		Key:              key,
		UpdateExpression: aws.String("SET #tz = :tz"),
		ExpressionAttributeNames: map[string]string{
			"#tz": "timezone",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tz": &types.AttributeValueMemberS{Value: timezone},
		},
	}
	_, err := ds.client.UpdateItem(context.TODO(), update)
	if err != nil {
		return fmt.Errorf("failed to update user timezone: %v", err)
	}
	return nil
}

// GetUser возвращает профиль пользователя по userID
func (ds *DynamoStorage) GetUser(userID int64) (*storage.User, error) {
	result, err := ds.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(UserTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", userID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user from DynamoDB: %v", err)
	}
	if result.Item == nil {
		return nil, nil
	}
	var user storage.User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %v", err)
	}
	return &user, nil
}

func (ds *DynamoStorage) DeleteUser(userID int64) error {
	_, err := ds.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(UserTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", userID)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user from DynamoDB: %v", err)
	}
	log.Printf("Deleted user %d from DynamoDB", userID)
	return nil
}
