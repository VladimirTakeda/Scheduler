package message

import (
	"InterviewScheduler/pkg"
	"InterviewScheduler/storage"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

var schedulerClient *scheduler.Client

func initScheduler() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	schedulerClient = scheduler.NewFromConfig(cfg)
}

func ProcessMessageUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, dbStorage storage.Storage) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received message from chat ID %d: %s", update.Message.Chat.ID, update.Message.Text)

	// Проверяем, ожидает ли пользователь ввода текста напоминания
	userState, err := dbStorage.GetUserState(update.Message.From.ID)
	if err != nil {
		log.Printf("Failed to get user state: %v", err)
	}

	// Если пользователь в состоянии ввода текста напоминания
	if userState != nil && userState.State == "waiting_reminder_text" {
		// Для интеграции с AWS Scheduler используйте HandleReminderTextInputWithSchedule
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
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	}

	switch {
	case update.Message.Text == "/start":
		// Сохраняем пользователя с таймзоной Bratislava
		err := dbStorage.SaveUser(
			update.Message.From.ID,
			update.Message.From.FirstName,
			update.Message.From.UserName,
			"Europe/Bratislava",
		)
		if err != nil {
			log.Printf("Failed to save user: %v", err)
		}

		// Описание возможностей бота
		preview := "Привет! Я твой Telegram бот на AWS Lambda. Я могу создавать напоминания и присылать их тебе в нужное время.\n\n" +
			"Доступные команды:\n" +
			"/remind — создать новое напоминание\n" +
			"/list — показать все твои напоминания\n" +
			"/cancel — отменить создание напоминания\n" +
			"/timezone — выбрать или изменить часовой пояс\n"

		// После /start показываем все команды
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
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	case update.Message.Text == "/remind":
		msg := pkg.WeekSelectionButtons(update.Message.Chat.ID)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send week selection: %v", err)
		}
	case update.Message.Text == "/cancel":
		err := dbStorage.ClearUserState(update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to clear user state: %v", err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создание напоминания отменено.")
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	case update.Message.Text == "/list":
		err := pkg.HandleListReminders(bot, dbStorage, update.Message.Chat.ID, update.Message.From.ID)
		if err != nil {
			log.Printf("Failed to handle list reminders: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении списка напоминаний.")
			bot.Send(msg)
		}
	case update.Message.Text == "/timezone":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите ваш часовой пояс:")
		msg.ReplyMarkup = pkg.TimezoneKeyboard(0)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send timezone keyboard: %v", err)
		}
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
	default:
		// Если пользователь только подключил бота (нет ��аписи в БД), показываем только /start
		userState, err := dbStorage.GetUserState(update.Message.From.ID)
		if err != nil || userState == nil {
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/start"),
				),
			)
			keyboard.ResizeKeyboard = true
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, нажмите /start для начала работы с ботом.")
			msg.ReplyMarkup = keyboard
			_, _ = bot.Send(msg)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
		}
		// Проверяем, зарегистрирован ли пользователь и установил ли он таймзону
		userProfile, _ := dbStorage.GetUser(update.Message.From.ID)
		if (userProfile == nil || userProfile.Timezone == "") && update.Message.Text != "/start" && update.Message.Text != "/timezone" {
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/start"),
					tgbotapi.NewKeyboardButton("/timezone"),
				),
			)
			keyboard.ResizeKeyboard = true
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, нажмите /start и выберите часовой пояс перед использованием других команд.")
			msg.ReplyMarkup = keyboard
			_, _ = bot.Send(msg)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Ты написал: %s\nИспользуй /remind для создания напоминания", update.Message.Text))
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "OK"}, nil
}
