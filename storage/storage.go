package storage

type Storage interface {
	SaveUser(userID int64, firstName, userName, timezone string) error
	GetUser(userID int64) (*User, error)
	DeleteUser(userID int64) error
	SaveUserTimezone(userID int64, timezone string) error
	GetUserState(userID int64) (*UserState, error)
	ClearUserState(userID int64) error
	GetReminderByID(reminderID, userID string) (*Reminder, error)
	SaveUserState(userState *UserState) error
	SaveReminder(reminder *Reminder) error
	GetUserReminders(userID int64) ([]Reminder, error)
	DeleteReminder(reminderID, userID string) error
}

type ProcessedEventStorage interface {
	IsEventProcessed(tgUpdateID string) (bool, error)
	MarkEventProcessed(tgUpdateID string) error
}
