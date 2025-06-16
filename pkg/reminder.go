package pkg

import (
	"InterviewScheduler/storage"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

var commands = map[string]struct{}{
	"/start":    {},
	"/remind":   {},
	"/cancel":   {},
	"/list":     {},
	"/timezone": {},
}

func HandleReminderTextInputWithSchedule(bot *tgbotapi.BotAPI, dbStorage storage.Storage, schedulerClient *scheduler.Client, lambdaArn, roleArn string, update tgbotapi.Update, userState *storage.UserState) error {
	reminderText := strings.TrimSpace(update.Message.Text)

	if reminderText == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Текст напоминания не может быть пустым. Попробуйте еще раз:")
		bot.Send(msg)
		return nil
	}

	if _, isCommand := commands[reminderText]; isCommand {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нельзя использовать служебную команду в качестве текста напоминания. Введите другой текст:")
		bot.Send(msg)
		return nil
	}

	// The time is already in user time zone, so to translate to UTC we need to subtract the offset
	timeParts := strings.Split(userState.Data["time"], ":")
	hour, _ := strconv.Atoi(timeParts[0])
	minute, _ := strconv.Atoi(timeParts[1])

	// Получаем часовой пояс из профиля пользователя
	userProfile, err := dbStorage.GetUser(update.Message.From.ID)
	userTimezone := "UTC"
	if err == nil && userProfile != nil && userProfile.Timezone != "" {
		log.Printf("Using user timezone from profile: %s", userProfile.Timezone)
		userTimezone = userProfile.Timezone
	} else {
		log.Printf("User timezone not set in profile, using UTC")
	}
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		log.Printf("Failed to load location for timezone '%s': %v", userTimezone, err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	log.Printf("Now: %v", now)
	year, month := now.Year(), now.Month()
	if userState.Data["week"] != "0" || userState.Data["week"] != "" {
		// Если выбрана не текущая неделя, корректируем месяц и год
		// (оставляем как есть, если логика weekOffset не используется для месяца)
	}
	day, _ := strconv.Atoi(userState.Data["day"])
	reminderDateLocal := time.Date(year, month, day, hour, minute, 0, 0, loc)
	reminderDateUTC := reminderDateLocal.UTC()
	dateTimeStr := reminderDateLocal.Format("02.01.2006 15:04 MST")

	// Создаем напоминание
	reminder := &storage.Reminder{
		UserID:    strconv.FormatInt(update.Message.From.ID, 10),
		Text:      reminderText,
		Week:      userState.Data["week"],
		Day:       userState.Data["day"],
		Time:      userState.Data["time"],
		DateTime:  dateTimeStr,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Timezone:  userTimezone,
	}

	err = dbStorage.SaveReminder(reminder)
	if err != nil {
		return fmt.Errorf("failed to save reminder: %v", err)
	}

	err = dbStorage.ClearUserState(update.Message.From.ID)
	if err != nil {
		log.Printf("Failed to clear user state: %v", err)
	}

	// Интеграция с AWS Scheduler
	if schedulerClient != nil && lambdaArn != "" && roleArn != "" {
		payload := map[string]string{"reminder_id": reminder.ID, "user_id": reminder.UserID}
		err = CreateReminderSchedules(context.TODO(), schedulerClient, lambdaArn, roleArn, reminder.ID, reminderDateUTC, payload)
		if err != nil {
			log.Printf("Failed to create AWS Scheduler jobs: %v", err)
		}
	}

	confirmationText := fmt.Sprintf("✅ Напоминание создано!\n\n📅 Дата и время: %s\n📝 Текст: %s\n\nБот напомнит вам за 24 часа, за 1 час и за 10 минут до события.", dateTimeStr, reminderText)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, confirmationText)
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send confirmation: %v", err)
	}

	return nil
}

func HandleListReminders(bot *tgbotapi.BotAPI, dbStorage storage.Storage, chatID int64, userID int64) error {
	reminders, err := dbStorage.GetUserReminders(userID)
	if err != nil {
		return fmt.Errorf("failed to get user reminders: %v", err)
	}

	if len(reminders) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📝 У вас пока нет напоминаний.\n\nИспользуйте /remind для создания нового напоминания.")
		_, err = bot.Send(msg)
		return err
	}

	for i, reminder := range reminders {
		// Используем точную дату и время из поля DateTime
		reminderTime := reminder.DateTime
		if reminderTime == "" {
			// fallback: если по каким-то причинам поле пустое, используем week/day/time
			return fmt.Errorf("failed to get reminder time for reminder ID %s", reminder.ID)
		}

		text := fmt.Sprintf("✅ Напоминание %d\n\n📅 Дата и время: %s\n📝 Текст: %s", i+1, reminderTime, reminder.Text)

		deleteButton := tgbotapi.NewInlineKeyboardButtonData(
			"🗑 Удалить",
			fmt.Sprintf("delete_%s", reminder.ID),
		)
		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{deleteButton})

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send reminder: %v", err)
		}
	}

	return nil
}

func HandleDeleteReminder(bot *tgbotapi.BotAPI, dbStorage storage.Storage, update tgbotapi.Update, data string, chatID int64) error {
	reminderID := strings.TrimPrefix(data, "delete_")

	var deletedReminder *storage.Reminder

	deletedReminder, err := dbStorage.GetReminderByID(reminderID, strconv.FormatInt(update.CallbackQuery.From.ID, 10))
	if err != nil {
		return fmt.Errorf("failed to get reminder by ID: %v", err)
	}

	err = dbStorage.DeleteReminder(reminderID, strconv.FormatInt(update.CallbackQuery.From.ID, 10))
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %v", err)
	}

	confirmation := "Напоминание удалено ✅"
	if deletedReminder != nil {
		confirmation = fmt.Sprintf("Удалено напоминание:\n\n📅 %s\n📝 %s", deletedReminder.DateTime, deletedReminder.Text)
	}

	// Сначала отправляем уведомление о том, что напоминание удалено
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, confirmation)
	_, err = bot.Request(callback)
	if err != nil {
		log.Printf("Failed to send callback response: %v", err)
	}

	// Затем отправляем сообщение с новым списком напоминаний
	msg := tgbotapi.NewMessage(chatID, "Новый список напоминаний после удаления:")
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send new list header: %v", err)
	}

	err = HandleListReminders(bot, dbStorage, chatID, update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("failed to refresh reminders list: %v", err)
	}

	return nil
}
