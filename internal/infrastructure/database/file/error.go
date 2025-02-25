package file

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
