package siteclients

import (
	"fmt"
	"linkTraccer/internal/domain/scrapper"
)

type Link = scrapper.Link

type ErrClientCantTrackLink struct {
	link   Link
	client string
}

func NewErrClientCantTrackLink(link Link, client string) *ErrClientCantTrackLink {
	return &ErrClientCantTrackLink{link: link, client: client}
}

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

func (err *ErrClientCantTrackLink) Error() string {
	return fmt.Sprintf("клиент %s не может отследить ссылку %s", err.client, err.link)
}
