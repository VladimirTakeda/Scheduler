package pkg

import (
	"InterviewScheduler/storage"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

func ProcessTimes(bot *tgbotapi.BotAPI, update tgbotapi.Update, data string, chatID int64, messageID int) {
	parts := data[5:]                // Убираем "time_"
	lastUnderscore := len(parts) - 6 // "_HH:MM" = 6 символов
	selectedDate := parts[:lastUnderscore]
	selectedTime := parts[lastUnderscore+1:]

	parsedDate, err := time.Parse("2006-01-02", selectedDate)
	dateStr := selectedDate
	if err == nil {
		dateStr = parsedDate.Format("02.01.2006")
	}

	confirmMsg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("✅ Напоминание создано!\n\n📅 Дата: %s\n🕐 Время: %s\n\nТеперь можно добавить текст напоминания или создать новое напоминание командой /remind",
			dateStr, selectedTime))
	_, err = bot.Send(confirmMsg)
	if err != nil {
		log.Printf("Failed to send confirmation: %v", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err = bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Напоминание создано!")
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}
}

func HandleTimeSelection(bot *tgbotapi.BotAPI, dbStorage storage.Storage, update tgbotapi.Update, data string, chatID int64, messageID int) (*storage.UserState, error) {
	// callbackData := fmt.Sprintf("time_%d_%d_%s", weekOffset, dayNumber, timeStr)
	parts := strings.Split(data, "_")
	if len(parts) != 4 {
		return nil, fmt.Errorf(fmt.Sprintf("invalid time data format, %s", data))
	}

	week := parts[1]
	day := parts[2]
	time := parts[3]

	userProfile, err := dbStorage.GetUser(update.CallbackQuery.From.ID)
	if err != nil || userProfile == nil || userProfile.Timezone == "" {
		return nil, fmt.Errorf("timezone not set for user")
	}

	userState := &storage.UserState{
		UserID: strconv.FormatInt(update.CallbackQuery.From.ID, 10),
		State:  "waiting_reminder_text",
		Data: map[string]string{
			"week": week,
			"day":  day,
			"time": time,
		},
	}

	err = dbStorage.SaveUserState(userState)
	if err != nil {
		return userState, fmt.Errorf("failed to save user state: %v", err)
	}

	edit := tgbotapi.NewEditMessageText(chatID, messageID,
		fmt.Sprintf("Выбрано время: %s\n\nТеперь введите текст напоминания:", time))
	edit.ReplyMarkup = nil

	_, err = bot.Send(edit)
	if err != nil {
		return userState, fmt.Errorf("failed to edit message: %v", err)
	}

	msg := tgbotapi.NewMessage(chatID, "📝 Введите текст вашего напоминания:\n\n(или /cancel для отмены)")
	_, err = bot.Send(msg)
	if err != nil {
		return userState, fmt.Errorf("failed to send instruction message: %v", err)
	}

	return userState, nil
}
