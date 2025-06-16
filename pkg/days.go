package pkg

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

func timeSelectionButtons(chatID int64, selectedDate string, weekOffset int, dayNumber int, timezone string) tgbotapi.MessageConfig {
	buttons := [][]tgbotapi.InlineKeyboardButton{}
	row := []tgbotapi.InlineKeyboardButton{}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[timeSelectionButtons] Failed to load location for timezone '%s': %v", timezone, err)
		return tgbotapi.NewMessage(chatID, "Ошибка: не удалось определить часовой пояс. Попробуйте позже.")
	}

	parsedDate, err := time.ParseInLocation("02.01.2006", selectedDate, loc)
	if err != nil {
		log.Printf("[timeSelectionButtons] Failed to parse selectedDate '%s' in location '%s': %v", selectedDate, timezone, err)
		return tgbotapi.NewMessage(chatID, "Ошибка: не удалось обработать выбранную дату. Попробуйте позже.")
	}
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	log.Printf("today: %v, parsedDate: %v", today, parsedDate)
	isToday := parsedDate.Equal(today)
	now = time.Now().In(loc)

	for hour := 7; hour <= 23; hour++ {
		for minute := 0; minute < 60; minute += 30 {
			timeStr := fmt.Sprintf("%02d:%02d", hour, minute)
			// Пропускаем время, если выбран сегодня и оно уже прошло
			if isToday {
				log.Printf("[timeSelectionButtons] hour=%d, now.Hour()=%d, minute=%d, now.Minute()=%d", hour, now.Hour(), minute, now.Minute())
				if hour < now.Hour() || (hour == now.Hour() && minute <= now.Minute()) {
					log.Printf("[timeSelectionButtons] SKIP: hour=%d, minute=%d (already passed)", hour, minute)
					continue
				}
			}
			callbackData := fmt.Sprintf("time_%d_%d_%s", weekOffset, dayNumber, timeStr)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(timeStr, callbackData))
			if len(row) == 4 {
				buttons = append(buttons, row)
				row = []tgbotapi.InlineKeyboardButton{}
			}
		}
		if hour == 23 {
			break
		}
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	backButton := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("← Назад к выбору дня", fmt.Sprintf("back_to_days_%d", weekOffset)),
	}
	buttons = append(buttons, backButton)

	parsedDate, err = time.ParseInLocation("2006-01-02", selectedDate, loc)
	dateStr := selectedDate
	if err == nil {
		dateStr = parsedDate.Format("02.01.2006")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Choose the time for for day: %s:", dateStr))
	msg.ReplyMarkup = keyboard
	return msg
}

func ProcessDays(bot *tgbotapi.BotAPI, update tgbotapi.Update, data string, chatID int64, messageID int) error {
	// we got only date here
	// callbackData := fmt.Sprintf("day_%d_%d", weekOffset, day)

	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return fmt.Errorf(fmt.Sprintf("invalid time data format, %s", data))
	}

	week, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])

	now := time.Now()
	selectedDate := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location()).Format("02.01.2006")

	msg := timeSelectionButtons(chatID, selectedDate, week, day, "Europe/Bratislava") // Convert char to int
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send time selection: %v", err)
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err = bot.Request(deleteMsg)
	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Выберите время")
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}
	return nil
}

func BackToDays(bot *tgbotapi.BotAPI, update tgbotapi.Update, data string, chatID int64, messageID int) {
	weekOffsetStr := data[13:]
	var weekOffset int
	fmt.Sscanf(weekOffsetStr, "%d", &weekOffset)

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

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}
}
