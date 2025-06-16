package callback

import (
	"InterviewScheduler/pkg"
	"InterviewScheduler/storage"
	"github.com/aws/aws-lambda-go/events"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func ProcessCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage) (events.APIGatewayProxyResponse, error) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID

	if len(data) > 7 && data[:7] == "tzpage_" {
		// Пагинация по часовым поясам
		page, _ := strconv.Atoi(data[7:])
		edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, pkg.TimezoneKeyboard(page))
		_, _ = bot.Send(edit)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}
	if len(data) > 3 && data[:3] == "tz_" {
		// Выбор часового пояса
		tz := data[3:]
		err := dbStorage.SaveUserTimezone(update.CallbackQuery.From.ID, tz)
		if err != nil {
			log.Printf("Failed to save timezone: %v", err)
		}
		msg := tgbotapi.NewMessage(chatID, "Часовой пояс установлен: "+tz)
		_, _ = bot.Send(msg)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}

	if data == "week_0" || data == "week_1" || data == "week_2" {
		pkg.ProcessWeeks(bot, update, data, chatID, messageID)
	}

	if data == "back_to_weeks" {
		pkg.BackToWeeks(bot, update, chatID, messageID)
	}

	if len(data) > 4 && data[:4] == "day_" {
		err := pkg.ProcessDays(bot, update, data, chatID, messageID)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
	}

	if len(data) > 13 && data[:13] == "back_to_days_" {
		pkg.BackToDays(bot, update, data, chatID, messageID)
	}

	if len(data) > 5 && data[:5] == "time_" {
		err := pkg.HandleTimeSelection(bot, dbStorage, update, data, chatID, messageID)
		if err != nil {
			log.Printf("Failed to handle time selection: %v", err)
		}
	}
	if len(data) > 7 && data[:7] == "delete_" {
		err := pkg.HandleDeleteReminder(bot, dbStorage, update, data, chatID)
		if err != nil {
			log.Printf("Failed to handle delete reminder: %v", err)
		}
	}
	if data == "tz_not_found" {
		msg := tgbotapi.NewMessage(chatID, "Если вашего города нет в списке, выберите ближайший город с таким же смещением UTC или напишите свой город в чат — мы поможем подобрать подходящий часовой пояс.")
		_, _ = bot.Send(msg)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
}
