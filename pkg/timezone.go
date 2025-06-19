package pkg

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Добавим структуру для отображения города, зоны и смещения
var timezonesData = []struct {
	City   string
	Zone   string
	Offset int // в часах относительно UTC
}{
	{"Лос-Анджелес", "America/Los_Angeles", -7},
	{"Мехико", "America/Mexico_City", -6},
	{"Чикаго", "America/Chicago", -5},
	{"Нью-Йорк", "America/New_York", -4},
	{"Торонто", "America/Toronto", -4},
	{"Сан-Паулу", "America/Sao_Paulo", -3},
	{"Лондон", "Europe/London", 1},
	{"Братислава", "Europe/Bratislava", 2},
	{"Берлин", "Europe/Berlin", 2},
	{"Париж", "Europe/Paris", 2},
	{"Рим", "Europe/Rome", 2},
	{"Мадрид", "Europe/Madrid", 2},
	{"Варшава", "Europe/Warsaw", 2},
	{"Прага", "Europe/Prague", 2},
	{"Москва", "Europe/Moscow", 3},
	{"Минск", "Europe/Minsk", 3},
	{"Киев", "Europe/Kiev", 3},
	{"Афины", "Europe/Athens", 3},
	{"Дубай", "Asia/Dubai", 4},
	{"Ташкент", "Asia/Tashkent", 5},
	{"Екатеринбург", "Asia/Yekaterinburg", 5},
	{"Алматы", "Asia/Almaty", 6},
	{"Бишкек", "Asia/Bishkek", 6},
	{"Бангкок", "Asia/Bangkok", 7},
	{"Новосибирск", "Asia/Novosibirsk", 7},
	{"Шанхай", "Asia/Shanghai", 8},
	{"Сингапур", "Asia/Singapore", 8},
	{"Токио", "Asia/Tokyo", 9},
	{"Владивосток", "Asia/Vladivostok", 10},
	{"Сидней", "Australia/Sydney", 10},
}

const timezonesPerPage = 8

func TimezoneKeyboard(page int) tgbotapi.InlineKeyboardMarkup {
	start := page * timezonesPerPage
	end := start + timezonesPerPage
	if end > len(timezonesData) {
		end = len(timezonesData)
	}
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for i := start; i < end; i++ {
		tz := timezonesData[i]
		label := fmt.Sprintf("%s (%s, UTC%+d)", tz.City, tz.Zone, tz.Offset)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "tz_"+tz.Zone),
		))
	}
	pagRow := []tgbotapi.InlineKeyboardButton{}
	if page > 0 {
		pagRow = append(pagRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("tzpage_%d", page-1)))
	}
	if end < len(timezonesData) {
		pagRow = append(pagRow, tgbotapi.NewInlineKeyboardButtonData("Вперёд ➡️", fmt.Sprintf("tzpage_%d", page+1)))
	}
	if len(pagRow) > 0 {
		rows = append(rows, pagRow)
	}
	// Кнопка "Моего города нет в списке"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Моего города нет в списке", "tz_not_found"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
