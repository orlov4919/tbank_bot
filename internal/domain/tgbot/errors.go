package tgbot

import "fmt"

type ErrFieldOkIsFalse struct {
	method string
}

func (err ErrFieldOkIsFalse) Error() string {
	return fmt.Sprintf("При вызове метода %s, поле ответа OK = false", err.method)
}
