package botservice

import (
	"fmt"
	"linkTraccer/internal/domain/tgbot"
)

type ErrCommandNotFound struct {
	command tgbot.EventType
}

func NewErrCommandNotFound(command tgbot.EventType) *ErrCommandNotFound {
	return &ErrCommandNotFound{command: command}
}

func (err *ErrCommandNotFound) Error() string {
	return fmt.Sprintf("команда %s не найдена", err.command)
}
