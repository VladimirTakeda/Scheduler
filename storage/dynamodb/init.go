package dynamodb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
)

type DynamoStorage struct {
	client *dynamodb.Client
}

func NewDynamoStorage() *DynamoStorage {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &DynamoStorage{dynamodb.NewFromConfig(cfg)}
}

type ProcessedEventRepo struct {
	Client *dynamodb.Client
}

func NewProcessedEventRepo() *ProcessedEventRepo {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &ProcessedEventRepo{Client: dynamodb.NewFromConfig(cfg)}
}
