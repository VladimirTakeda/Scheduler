package pkg

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

func WeekSelectionButtons(chatID int64) tgbotapi.MessageConfig {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("this week          ", "week_0"),
			tgbotapi.NewInlineKeyboardButtonData("next week          ", "week_1"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("week after next", "week_2"),
		},
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, "Choose the week for the interview reminder:")
	msg.ReplyMarkup = keyboard
	return msg
}

func workDaysButtons(chatID int64, weekOffset int) tgbotapi.MessageConfig {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		// in go time package, Sunday is 0, but we want Monday to be the first day of the week
		weekday = 7
	}
	mondayThisWeek := now.AddDate(0, 0, -weekday+1)

	targetMonday := mondayThisWeek.AddDate(0, 0, 7*weekOffset)

	buttons := [][]tgbotapi.InlineKeyboardButton{}
	row := []tgbotapi.InlineKeyboardButton{}

	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}

	for i := 0; i < 5; i++ {
		day := targetMonday.AddDate(0, 0, i)
		label := fmt.Sprintf("%s %s", weekdays[i], day.Format("02"))
		callbackData := fmt.Sprintf("day_%d_%d", weekOffset, day.Day())

		if weekOffset == 0 && day.Before(now.Truncate(24*time.Hour)) {
			continue
		}

		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, callbackData))
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	backButton := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("← Назад к выбору недели", "back_to_weeks"),
	}
	buttons = append(buttons, backButton)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, "Choose the day for the interview reminder:")
	msg.ReplyMarkup = keyboard
	return msg
}

func ProcessWeeks(bot *tgbotapi.BotAPI, update tgbotapi.Update, data string, chatID int64, messageID int) {
	var weekOffset int
	fmt.Sscanf(data, "week_%d", &weekOffset)

	msg := workDaysButtons(chatID, weekOffset)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send work days buttons: %v", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err = bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Выберите день")
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}
}

func BackToWeeks(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, messageID int) {
	msg := WeekSelectionButtons(chatID)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send week selection: %v", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err = bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}
}
