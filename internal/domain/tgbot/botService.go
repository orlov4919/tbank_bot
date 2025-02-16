package tgbot

type BotService interface {
	SendMessage(userID int, text string)
	HandleUsersUpdates() Updates
	ChangeDialogState(update Update)
	ChangeOffset(offset int)
}
