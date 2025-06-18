package message

import (
	"InterviewScheduler/pkg"
	"InterviewScheduler/storage"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

var schedulerClient *scheduler.Client

func initScheduler() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	schedulerClient = scheduler.NewFromConfig(cfg)
}

const (
	StateIdle                = "idle"
	StateWaitingTimezone     = "waiting_timezone"
	StateWaitingWeek         = "waiting_week"
	StateWaitingDay          = "waiting_day"
	StateWaitingTime         = "waiting_time"
	StateWaitingReminderText = "waiting_reminder_text"
)

type StateHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error)

var stateHandlers = map[string]StateHandler{
	StateIdle:                handleIdleState,
	StateWaitingReminderText: handleWaitingReminderText,
	StateWaitingTimezone:     handleWaitingTimezone,
	StateWaitingWeek:         handleWaitingWeek,
	StateWaitingDay:          handleWaitingDay,
	StateWaitingTime:         handleWaitingTime,
}

func ProcessMessageUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage) (events.APIGatewayProxyResponse, error) {
	if update.Message == nil || update.Message.From == nil {
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "No message"}, nil
	}
	// we can try to get user stage from message or if it's unclear - from db
	userState, err := dbStorage.GetUserState(update.Message.From.ID)
	if err != nil {
		log.Printf("Failed to get user state: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения состояния пользователя. Попробуйте позже.")
		_, err := bot.Send(msg)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "tg error"}, err
		}
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Error"}, nil
	}
	if userState == nil {
		userState = &storage.UserState{UserID: strconv.FormatInt(update.Message.From.ID, 10), State: StateIdle}
	}
	handler, ok := stateHandlers[userState.State]
	if !ok {
		handler = handleIdleState
	}
	newState, err := handler(bot, update, dbStorage, userState)
	if err != nil {
		log.Printf("State handler error: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка. Попробуйте еще раз.")
		bot.Send(msg)
	}
	if newState != userState.State {
		err := dbStorage.SaveUserState(&storage.UserState{UserID: strconv.FormatInt(update.Message.From.ID, 10), State: newState})
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "tg error"}, err
		}
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
}

func handleIdleState(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	switch update.Message.Text {
	case "/start":
		err := dbStorage.SaveUser(
			update.Message.From.ID,
			update.Message.From.FirstName,
			update.Message.From.UserName,
			"Europe/Bratislava",
		)
		if err != nil {
			log.Printf("Failed to save user: %v", err)
		}

		preview := "Привет! Я твой Telegram бот на AWS Lambda. Я могу создавать напоминания и присылать их тебе в нужное время.\n\n" +
			"Доступные команды:\n" +
			"/remind — создать новое напоминание\n" +
			"/list — показать все твои напоминания\n" +
			"/cancel — отменить создание напоминания\n" +
			"/timezone — выбрать или изменить часовой пояс\n"

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/start"),
				tgbotapi.NewKeyboardButton("/remind"),
				tgbotapi.NewKeyboardButton("/list"),
				tgbotapi.NewKeyboardButton("/cancel"),
			),
		)
		keyboard.ResizeKeyboard = true

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, preview)
		msg.ReplyMarkup = keyboard
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return StateWaitingTimezone, nil
	case "/remind":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите текст напоминания:")
		bot.Send(msg)
		return StateWaitingReminderText, nil
	case "/cancel":
		err := dbStorage.ClearUserState(update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to clear user state: %v", err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return StateIdle, nil
	case "/list":
		err := pkg.HandleListReminders(bot, dbStorage, update.Message.Chat.ID, update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to handle list reminders: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка напоминаний.")
			bot.Send(msg)
		}
		return StateIdle, nil
	case "/timezone":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите ваш часовой пояс:")
		msg.ReplyMarkup = pkg.TimezoneKeyboard(0)
		bot.Send(msg)
		return StateIdle, nil
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /remind /timezone /list")
		bot.Send(msg)
		return StateIdle, nil
	}
}

func handleWaitingReminderText(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/cancel" {
		err := dbStorage.ClearUserState(update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to clear user state: %v", err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return StateIdle, nil
	}
	lambdaArn := os.Getenv("LAMBDA_ARN")
	roleArn := os.Getenv("SCHEDULER_ROLE_ARN")
	if schedulerClient == nil {
		initScheduler()
	}
	err := pkg.HandleReminderTextInputWithSchedule(bot, dbStorage, schedulerClient, lambdaArn, roleArn, update, userState)
	if err != nil {
		log.Printf("Failed to handle reminder text input: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при сохранении напоминания. Попробуйте еще раз.")
		bot.Send(msg)
	}
	return StateIdle, nil
}

func handleWaitingTimezone(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	// Здесь логика выбора и сохранения таймзоны
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Таймзона установлена!")
	bot.Send(msg)
	return StateIdle, nil
}

func handleWaitingWeek(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	// Здесь логика выбора и сохранения недели
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неделя установлена!")
	bot.Send(msg)
	return StateIdle, nil
}

func handleWaitingDay(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	// Здесь логика выбора и сохранения дня
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День установлен!")
	bot.Send(msg)
	return StateIdle, nil
}

func handleWaitingTime(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	// Здесь логика выбора и сохранения времени
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Время установлено!")
	bot.Send(msg)
	return StateIdle, nil
}
