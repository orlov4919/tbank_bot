package scrapclient

import "fmt"

type ErrBadRequestStatus struct {
	msg  string
	code int
}

func NewErrBadRequestStatus(msg string, code int) *ErrBadRequestStatus {
	return &ErrBadRequestStatus{
		msg:  msg,
		code: code,
	}
}

func (err *ErrBadRequestStatus) Error() string {
	return fmt.Sprintf("%s код ответа сервера: %d", err.msg, err.code)
}
