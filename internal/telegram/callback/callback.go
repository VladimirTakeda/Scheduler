package callback

import (
	"InterviewScheduler/internal/telegram/message"
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
		page, _ := strconv.Atoi(data[7:])
		edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, pkg.TimezoneKeyboard(page))
		_, _ = bot.Send(edit)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}
	if len(data) > 3 && data[:3] == "tz_" {
		tz := data[3:]
		err := dbStorage.SaveUserTimezone(update.CallbackQuery.From.ID, tz)
		if err != nil {
			log.Printf("Failed to save timezone: %v", err)
		}
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/start"),
				tgbotapi.NewKeyboardButton("/list"),
				tgbotapi.NewKeyboardButton("/remind"),
				tgbotapi.NewKeyboardButton("/timezone"),
				tgbotapi.NewKeyboardButton("/cancel"),
			),
		)
		keyboard.ResizeKeyboard = true

		msg := tgbotapi.NewMessage(chatID, "Часовой пояс установлен: "+tz)
		msg.ReplyMarkup = keyboard
		_, _ = bot.Send(msg)
		// Update user state to idle2 after setting timezone
		err = dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10), State: message.StateIdle2})
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}

	if data == "week_0" || data == "week_1" || data == "week_2" {
		pkg.ProcessWeeks(bot, update, data, chatID, messageID)
		dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10), State: message.StateWaitingDay})
	}

	if data == "back_to_weeks" {
		pkg.BackToWeeks(bot, update, chatID, messageID)
		dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10), State: message.StateWaitingWeek})
	}

	if len(data) > 4 && data[:4] == "day_" {
		err := pkg.ProcessDays(bot, update, data, chatID, messageID)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10), State: message.StateWaitingTime})
	}

	if len(data) > 13 && data[:13] == "back_to_days_" {
		pkg.BackToDays(bot, update, data, chatID, messageID)
		dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10), State: message.StateWaitingDay})
	}

	if len(data) > 5 && data[:5] == "time_" {
		userState, err := pkg.HandleTimeSelection(bot, dbStorage, update, data, chatID, messageID)
		if err != nil {
			log.Printf("Failed to handle time selection: %v", err)
		}
		dbStorage.SaveUserState(userState)
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
