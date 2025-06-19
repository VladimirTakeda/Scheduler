# InterviewScheduler Telegram Bot

InterviewScheduler — это Telegram-бот для создания и управления напоминаниями с использованием AWS Lambda, DynamoDB и AWS Scheduler.

## Возможности
- Создание напоминаний с выбором недели, дня, времени и текста
- Управление таймзоной пользователя
- Просмотр и удаление напоминаний
- Хранение данных в DynamoDB
- Масштабируемое выполнение через AWS Lambda

## Архитектура
- **Go** — основной язык
- **tgbotapi** — работа с Telegram Bot API
- **AWS SDK** — интеграция с AWS сервисами
- **DynamoDB** — хранение пользователей и напоминаний
- **State Machine** — управление этапами взаимодействия с пользователем

## Диаграмма состояний (State Machine)

```mermaid
stateDiagram-v2
    [*] --> idle: /start
    idle --> waiting_timezone: /timezone
    waiting_timezone --> idle2
    idle2 --> remind: /remind
    idle2 --> list: /list
    idle2 --> timezone: /timezone
    remind --> waiting_week
    waiting_week --> waiting_day: "Неделя выбрана"
    waiting_week --> idle2: /cancel
    waiting_day --> waiting_time: "День выбран"
    waiting_day --> idle2: /cancel
    waiting_time --> waiting_reminder_text: "Время выбрано"
    waiting_time --> idle2: /cancel
    waiting_reminder_text --> idle2: "Текст напоминания введён, Напоминание создано"
    waiting_reminder_text --> idle2: /cancel
```

## Быстрый старт
1. Склонируйте репозиторий
2. Установите зависимости: `go mod tidy`
3. Настройте переменные окружения для AWS и Telegram Bot Token
4. Запустите локально или задеплойте в AWS Lambda

## Структура проекта
- `internal/telegram/message.go` — основная логика state machine
- `pkg/` — вспомогательные функции (клавиатуры, обработка дат и времени)
- `storage/` — работа с базой данных

## Контакты
- Автор: @yourusername
- Вопросы и предложения: issues/pull requests приветствуются!

## TODO
1) Если напоминание закончилось - удалить его из базы данных
