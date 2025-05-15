package telegram

import "fmt"

func NewErrNegativeLimit(limit int) *ErrNegativeLimit {
	return &ErrNegativeLimit{
		limit: limit,
	}
}

type ErrNegativeLimit struct {
	limit int
}

func (err ErrNegativeLimit) Error() string {
	return fmt.Sprintf("Лимит не может быть %d, значение лимита должно быть 1 - 100", err.limit)
}

type ErrBotAPI struct {
	code int
}

func (err *ErrBotAPI) Error() string {
	return fmt.Sprintf("Запрос к BOT API закончился ошибкой %d", err.code)
}

func NewErrBotAPI(code int) *ErrBotAPI {
	return &ErrBotAPI{code: code}
}
