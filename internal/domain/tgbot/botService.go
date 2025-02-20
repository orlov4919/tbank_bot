package tgbot


//go:generate mockery

type BotService interface {
	SendMessage(userID int, text string) error
	HandleUsersUpdates() (Updates, error)
	ChangeOffset(offset int)
}

//ChangeDialogState(update Update)
