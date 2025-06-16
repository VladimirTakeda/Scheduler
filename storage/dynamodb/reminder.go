package dynamodb

import (
	"InterviewScheduler/storage"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"strconv"
)

const UserReminderTableName = "SchedulerBotReminders"

func (ds *DynamoStorage) SaveReminder(reminder *storage.Reminder) error {
	// Генерируем UUID для напоминания
	reminder.ID = uuid.NewString()

	item, err := attributevalue.MarshalMap(reminder)
	if err != nil {
		return fmt.Errorf("failed to marshal reminder: %v", err)
	}

	_, err = ds.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(UserReminderTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to save reminder to DynamoDB: %v", err)
	}

	return nil
}

func (ds *DynamoStorage) GetUserReminders(userID int64) ([]storage.Reminder, error) {
	result, err := ds.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(UserReminderTableName),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: strconv.FormatInt(userID, 10)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders from DynamoDB: %v", err)
	}

	var reminders []storage.Reminder
	for _, item := range result.Items {
		var reminder storage.Reminder
		err = attributevalue.UnmarshalMap(item, &reminder)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal reminder: %v", err)
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

func (ds *DynamoStorage) GetReminderByID(reminderID, userID string) (*storage.Reminder, error) {
	result, err := ds.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(UserReminderTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID}, // необходимо указывать как partition key так и sort ket
			"id":      &types.AttributeValueMemberS{Value: reminderID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder from DynamoDB: %v", err)
	}
	if result.Item == nil || len(result.Item) == 0 {
		return nil, fmt.Errorf("reminder not found")
	}
	var reminder storage.Reminder
	err = attributevalue.UnmarshalMap(result.Item, &reminder)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reminder: %v", err)
	}
	return &reminder, nil
}

// doesn't work so far
func (ds *DynamoStorage) DeleteReminder(reminderID, userID string) error {
	_, err := ds.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(UserReminderTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"id":      &types.AttributeValueMemberS{Value: reminderID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete reminder from DynamoDB: %v", err)
	}

	return nil
}
