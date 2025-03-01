package contextstorage

import "fmt"

type ErrUserAlreadyReg struct {
	msg ID
}

func (err *ErrUserAlreadyReg) Error() string {
	return fmt.Sprintf("пользователь с id = %d уже регистрировался", err.msg)
}

func NewErrUserAlreadyReg(id ID) *ErrUserAlreadyReg {
	return &ErrUserAlreadyReg{
		msg: id,
	}
}

type ErrUserNotReg struct {
	msg ID
}

func (err *ErrUserNotReg) Error() string {
	return fmt.Sprintf("пользователь с id = %d не регистрировался", err.msg)
}

func NewErrUserNotReg(id ID) *ErrUserNotReg {
	return &ErrUserNotReg{
		msg: id,
	}
}
