package internal

import (
	"InterviewScheduler/storage"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func HandleSchedulerEvent(storage storage.Storage, reminderID, userID string) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received Scheduler event: reminder_id=%s, user_id=%s", reminderID, userID)

	if reminderID == "" || userID == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Missing reminder_id or user_id"}, nil
	}

	reminder, err := storage.GetReminderByID(reminderID, userID)
	if err != nil {
		log.Printf("Failed to get reminder: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to fetch reminder"}, nil
	}

	uid, _ := strconv.ParseInt(reminder.UserID, 10, 64)
	msg := tgbotapi.NewMessage(uid, fmt.Sprintf("🔔 Напоминание!\n\n📅 %s\n📝 %s", reminder.DateTime, reminder.Text))
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send reminder: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to send Telegram message"}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Reminder sent"}, nil
}
