package storage

type User struct {
	UserID       string `dynamodbav:"user_id"`
	FirstName    string `dynamodbav:"first_mame"`
	Username     string `dynamodbav:"user_name"`
	RegisteredAt string `dynamodbav:"registered_at"`
	Timezone     string `dynamodbav:"timezone"`
}

type Reminder struct {
	UserID    string `json:"user_id" dynamodbav:"user_id"`
	Text      string `json:"text" dynamodbav:"text"`
	Week      string `json:"week" dynamodbav:"week"`
	Day       string `json:"day" dynamodbav:"day"`
	Time      string `json:"time" dynamodbav:"time"`
	DateTime  string `json:"datetime" dynamodbav:"datetime"`
	CreatedAt string `json:"created_at" dynamodbav:"created_at"`
	ID        string `json:"id" dynamodbav:"id"` // Будет генерироваться автоматически
	Timezone  string `json:"timezone" dynamodbav:"timezone"`
}

type UserState struct {
	UserID string            `json:"user_id" dynamodbav:"user_id"`
	State  string            `json:"state" dynamodbav:"state"`
	Data   map[string]string `json:"data" dynamodbav:"data"`
}
