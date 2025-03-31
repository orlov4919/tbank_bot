package dto

// HTTP codes
const (
	httpStatusBadRequest = "400"
	httpStatusNotFound   = "404"
)

// api err descriptions
const (
	wrongRequestArg = "Некорректные параметры запроса"
	// methodNotAllowed  = "Метод не поддерживается"
	// chatNotRegistered = "Чат не существует"
	// linkNotFound  = "Ссылка не найдена"
	// internalError = "Произошла внутренняя ошибка"
)

// exceptions name
const (
	errId = "id error"
)

// exceptions message
const (
	idNotNum        = "id не соответствует числу"
	negativeID      = "полученное id < 0, должно быть id >=0"
	idRegistered    = "id уже зарегистрирован"
	idNotRegistered = "id не зарегистрирован"
)

// api errors
var (
	ApiErrIDNotNum          = newAPIErrResponse(wrongRequestArg, errId, idNotNum, httpStatusBadRequest)
	ApiErrNegativeID        = newAPIErrResponse(wrongRequestArg, errId, negativeID, httpStatusBadRequest)
	ApiErrUserRegistered    = newAPIErrResponse(wrongRequestArg, errId, idRegistered, httpStatusBadRequest)
	ApiErrUserNotRegistered = newAPIErrResponse(wrongRequestArg, errId, idNotRegistered, httpStatusNotFound)
)

type APIErrResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

func newAPIErrResponse(description, exceptionName, exceptionMessage, code string) *APIErrResponse {
	return &APIErrResponse{
		Description:      description,
		Code:             code,
		ExceptionName:    exceptionName,
		ExceptionMessage: exceptionMessage,
		Stacktrace:       []string{},
	}
}
