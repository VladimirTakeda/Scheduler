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

func HandleTelegramRequest(req events.APIGatewayProxyRequest, storage storage.Storage) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received message %s", req.Body)
	var update tgbotapi.Update
	err := json.Unmarshal([]byte(req.Body), &update)
	if err != nil {
		log.Printf("Failed to parse update: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
	}

	if update.MyChatMember != nil && update.MyChatMember.NewChatMember.Status == "kicked" {
		kickedUpdate, err := kicked.ProcessKickedUpdate(bot, update, storage)
		return kickedUpdate, err
	}

	if update.Message != nil {
		messageUpdate, err := message.ProcessMessageUpdate(bot, update, storage)
		return messageUpdate, err
	}

	if update.CallbackQuery != nil {
		query, err := callback.ProcessCallbackQuery(bot, update, storage)
		return query, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}
