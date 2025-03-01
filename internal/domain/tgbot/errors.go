package tgbot

import "fmt"

type ErrMachineCreationFailed struct {
}

func NewErrMachineCreationFailed() *ErrMachineCreationFailed {
	return &ErrMachineCreationFailed{}
}

func (err *ErrMachineCreationFailed) Error() string {
	return "При создании FSM, не указан переход из зачального состояния"
}

type ErrEventDeclined struct {
}

func NewErrEventDeclined() *ErrEventDeclined {
	return &ErrEventDeclined{}
}

func (err *ErrEventDeclined) Error() string {
	return "В запрашиваемом состоянии, не существует такого перехода"
}

type ErrTransitionFailed struct {
	id    ID
	event EventType
}

func NewErrTransitionFailed(id ID, event EventType) *ErrTransitionFailed {
	return &ErrTransitionFailed{
		id:    id,
		event: event,
	}
}

func (err *ErrTransitionFailed) Error() string {
	return fmt.Sprintf("Не получилось сделать переход по событию %s у user = %d", err.event, err.id)
}
