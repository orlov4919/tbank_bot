package botclient

import "fmt"

type ErrBadAnswerFromServer struct {
	code int
}

func NewErrBadAnswerFromServer(code int) *ErrBadAnswerFromServer {
	return &ErrBadAnswerFromServer{
		code: code,
	}
}

func (err *ErrBadAnswerFromServer) Error() string {
	return fmt.Sprintf("сервер бота прислал ответ: %d ", err.code)
}
