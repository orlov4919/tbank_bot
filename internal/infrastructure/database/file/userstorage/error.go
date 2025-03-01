package userstorage

import "fmt"

type ErrWithStorage struct {
	msg string
}

func NewErrWithStorage(msg string) *ErrWithStorage {
	return &ErrWithStorage{msg: msg}
}

func (e *ErrWithStorage) Error() string {
	return fmt.Sprintf("Произошла ошибка при работе с хранилищем: %s", e.msg)
}

type ErrUserNotRegistered struct {
	msg int
}

func NewErrUserNotRegistered(id User) *ErrUserNotRegistered {
	return &ErrUserNotRegistered{
		msg: id,
	}
}

func (err *ErrUserNotRegistered) Error() string {
	return fmt.Sprintf("Пользователь с id = %d не зарегистрирован", err.msg)
}
