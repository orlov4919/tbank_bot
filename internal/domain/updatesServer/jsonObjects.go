package updatesServer

type LinkUpdate struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	TgChatIds   []int  `json:"tgChatIds"`
}

type ApiErrorResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

func NewApiErrorResponse(exeptionName, exeptionMessage string, stackTrace []string) *ApiErrorResponse {
	return &ApiErrorResponse{
		Description:      "Некорректные параметры запроса",
		Code:             "400",
		ExceptionName:    exeptionName,
		ExceptionMessage: exeptionMessage,
		Stacktrace:       stackTrace,
	}
}
