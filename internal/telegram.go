package internal

import (
	"InterviewScheduler/internal/telegram/callback"
	"InterviewScheduler/internal/telegram/kicked"
	"InterviewScheduler/internal/telegram/message"
	"InterviewScheduler/storage"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

var bot *tgbotapi.BotAPI

// called automatically when the Lambda function is initialized
func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN env var is not set")
	}

	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
}

func HandleTelegramRequest(req events.APIGatewayProxyRequest, storage storage.Storage, processedRepo storage.ProcessedEventStorage) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received message %s", req.Body)
	var update tgbotapi.Update
	err := json.Unmarshal([]byte(req.Body), &update)
	if err != nil {
		log.Printf("Failed to parse update: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Bad Request"}, nil
	}

	processed, err := processedRepo.IsEventProcessed(strconv.Itoa(update.UpdateID))
	if err != nil {
		log.Printf("Failed to check processed event: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Internal error"}, nil
	}
	if processed {
		log.Printf("Event already processed: %s", update.UpdateID)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Event already processed"}, nil
	}

	if update.MyChatMember != nil && update.MyChatMember.NewChatMember.Status == "kicked" {
		kickedUpdate, err := kicked.ProcessKickedUpdate(bot, update, storage)
		err = processedRepo.MarkEventProcessed(strconv.Itoa(update.UpdateID))
		if err != nil {
			log.Printf("Failed to mark event processed: %v", err)
		}
		return kickedUpdate, err
	}

	if update.Message != nil {
		messageUpdate, err := message.ProcessMessageUpdate(bot, update, storage)
		err = processedRepo.MarkEventProcessed(strconv.Itoa(update.UpdateID))
		if err != nil {
			log.Printf("Failed to mark event processed: %v", err)
		}
		return messageUpdate, err
	}

	if update.CallbackQuery != nil {
		query, err := callback.ProcessCallbackQuery(bot, update, storage)
		err = processedRepo.MarkEventProcessed(strconv.Itoa(update.UpdateID))
		if err != nil {
			log.Printf("Failed to mark event processed: %v", err)
		}
		return query, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}
