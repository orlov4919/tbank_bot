package dto

type LinkUpdate struct {
	ID          int64   `json:"id"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	TgChatIDs   []int64 `json:"tgChatIds"`
}

type APIErrResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}

func NewAPIErrResponse(exeptionName, exeptionMessage string, stackTrace []string) *APIErrResponse {
	return &APIErrResponse{
		Description:      "Некорректные параметры запроса",
		Code:             "400",
		ExceptionName:    exeptionName,
		ExceptionMessage: exeptionMessage,
		Stacktrace:       stackTrace,
	}
}
