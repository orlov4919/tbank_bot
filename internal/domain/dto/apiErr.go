package dto

// HTTP codes.
const (
	httpStatusBadRequest = "400"
	httpStatusNotFound   = "404"
)

// api err descriptions.
const (
	wrongRequestArg = "Некорректные параметры запроса"
)

// exceptions name.
const (
	errID   = "id error"
	errBody = "body error"
	errLink = "link erroe"
)

// exceptions message.
const (
	idNotNum             = "id не соответствует числу"
	negativeID           = "полученное id < 0, должно быть id >=0"
	idRegistered         = "id уже зарегистрирован"
	idNotRegistered      = "id не зарегистрирован"
	cantReadBody         = "не возможно прочитать body"
	badBodyJSON          = "JSON имеет не правильный формат"
	cantTrackLink        = "переданная ссылка не поддерживается"
	userAlreadyTrackLink = "пользователь уже отслеживает эту ссылку"
	userNotTrackLink     = "пользователь не отслеживает эту ссылку"
)

// api errors chat handler.
var (
	APIErrIDNotNum          = newAPIErrResponse(errID, idNotNum, httpStatusBadRequest)
	APIErrNegativeID        = newAPIErrResponse(errID, negativeID, httpStatusBadRequest)
	APIErrUserRegistered    = newAPIErrResponse(errID, idRegistered, httpStatusBadRequest)
	APIErrUserNotRegistered = newAPIErrResponse(errID, idNotRegistered, httpStatusNotFound)
	APIErrCantReadBody      = newAPIErrResponse(errBody, cantReadBody, httpStatusBadRequest)
	APIErrBadJSON           = newAPIErrResponse(errBody, badBodyJSON, httpStatusBadRequest)
	APIErrBadLink           = newAPIErrResponse(errLink, cantTrackLink, httpStatusBadRequest)
	APIErrDuplicateLink     = newAPIErrResponse(errLink, userAlreadyTrackLink, httpStatusBadRequest)
	APIErrNotTrackLink      = newAPIErrResponse(errLink, userNotTrackLink, httpStatusNotFound)
)

type APIErrResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

func newAPIErrResponse(exceptionName, exceptionMessage, code string) *APIErrResponse {
	return &APIErrResponse{
		Description:      wrongRequestArg,
		Code:             code,
		ExceptionName:    exceptionName,
		ExceptionMessage: exceptionMessage,
		Stacktrace:       []string{},
	}
}
