package kicked

import (
	"InterviewScheduler/storage"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ProcessKickedUpdate(_ *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage) (events.APIGatewayProxyResponse, error) {
	chatID := update.MyChatMember.Chat.ID
	userID := update.MyChatMember.From.ID
	newStatus := update.MyChatMember.NewChatMember.Status

	fmt.Printf("Bot was %s by user %d in chat %d\n", newStatus, userID, chatID)

	// TODO: delete user reminders
	err := dbStorage.DeleteUser(userID)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Can't delete user, try later"}, err
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "ok"}, nil
}
