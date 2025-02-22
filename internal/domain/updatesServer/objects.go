package updatesServer

type LinkUpdate struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	TgChatIds   []int  `json:"tgChatIds"`
}

type ApiErrResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

func NewApiErrResponse(exeptionName, exeptionMessage string, stackTrace []string) *ApiErrResponse {
	return &ApiErrResponse{
		Description:      "Некорректные параметры запроса",
		Code:             "400",
		ExceptionName:    exeptionName,
		ExceptionMessage: exeptionMessage,
		Stacktrace:       stackTrace,
	}
}
