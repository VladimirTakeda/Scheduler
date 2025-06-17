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
    [*] --> idle
    idle --> waiting_timezone: /start, /timezone
    idle --> waiting_reminder_text: /remind
    idle --> idle: /list, /cancel, unknown
    waiting_timezone --> idle: "Таймзона выбрана"
    waiting_week --> waiting_day: "Неделя выбрана"
    waiting_week --> idle: /cancel
    waiting_day --> waiting_time: "День выбран"
    waiting_day --> idle: /cancel
    waiting_time --> waiting_reminder_text: "Время выбрано"
    waiting_time --> idle: /cancel
    waiting_reminder_text --> idle: "Текст напоминания введён, Напоминание создано"
    waiting_reminder_text --> idle: /cancel
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
