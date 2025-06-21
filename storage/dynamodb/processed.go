package dynamodb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
)

const ProcessesEvents = "ProcessedEvents"

func (r *ProcessedEventRepo) IsEventProcessed(tgUpdateID string) (bool, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(ProcessesEvents),
		Key: map[string]types.AttributeValue{
			"tg_update_id": &types.AttributeValueMemberS{Value: tgUpdateID},
		},
	}
	result, err := r.Client.GetItem(context.TODO(), input)
	if err != nil {
		return false, err
	}
	return result.Item != nil && len(result.Item) > 0, nil
}

func (r *ProcessedEventRepo) MarkEventProcessed(tgUpdateID string) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(ProcessesEvents),
		Item: map[string]types.AttributeValue{
			"tg_update_id": &types.AttributeValueMemberS{Value: tgUpdateID},
		},
	}
	_, err := r.Client.PutItem(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to mark event processed: %v", err)
	}
	return err
}
