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
	StateIdle                = "idle"             // Initial state after /start command, need timezine to proceed
	StateIdle2               = "idle2"            // State after timezone is set, can create reminders
	StateWaitingTimezone     = "waiting_timezone" // send buttons to user to select timezone
	StateWaitingWeek         = "waiting_week"
	StateWaitingDay          = "waiting_day"
	StateWaitingTime         = "waiting_time"
	StateWaitingReminderText = "waiting_reminder_text"
)

type StateHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error)

var stateHandlers = map[string]StateHandler{
	StateIdle:                handleIdleState,
	StateIdle2:               handleIdle2State,
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
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: "tg error"}, err
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
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: "tg error"}, err
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
			"Давай установим твою таймзону, нажми /timezone:\n"

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/timezone"),
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
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /start")
		bot.Send(msg)
		return StateIdle, nil
	}
}

func handleIdle2State(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	switch update.Message.Text {
	case "/remind":
		msg := pkg.WeekSelectionButtons(update.Message.Chat.ID)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send week selection: %v", err)
		}
		return StateWaitingWeek, nil
	case "/list":
		err := pkg.HandleListReminders(bot, dbStorage, update.Message.Chat.ID, update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to handle list reminders: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка напоминаний.")
			bot.Send(msg)
		}
		return StateIdle2, nil
	case "/timezone":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите ваш часовой пояс:")
		msg.ReplyMarkup = pkg.TimezoneKeyboard(0)
		bot.Send(msg)
		return StateWaitingTimezone, nil
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
			"Давай установим твою таймзону, нажми /timezone:\n"

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/timezone"),
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
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /remind /timezone /list /start")
		bot.Send(msg)
		return StateIdle2, nil
	}
}

func handleWaitingTimezone(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/timezone" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите ваш часовой пояс:")
		msg.ReplyMarkup = pkg.TimezoneKeyboard(0)
		bot.Send(msg)
		return StateWaitingTimezone, nil
	}
	if update.Message.Text == "/cancel" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбор часового пояса отменён.")
		bot.Send(msg)
		return StateIdle2, nil
	}
	return StateWaitingTimezone, nil
}

func handleWaitingWeek(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/cancel" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		bot.Send(msg)
		return StateIdle2, nil
	}
	return StateWaitingWeek, nil
}

func handleWaitingDay(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/cancel" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		bot.Send(msg)
		return StateIdle2, nil
	}
	return StateWaitingDay, nil
}

func handleWaitingTime(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/cancel" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		bot.Send(msg)
		return StateIdle2, nil
	}
	return StateWaitingTime, nil
}

func handleWaitingReminderText(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage, userState *storage.UserState) (string, error) {
	if update.Message.Text == "/cancel" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		bot.Send(msg)
		return StateIdle2, nil
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
	return StateIdle2, nil
}
