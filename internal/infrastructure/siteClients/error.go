package siteClients

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

func (err *ErrClientCantTrackLink) Error() string {
	return fmt.Sprintf("клиент %s не может отследить ссылку %s", err.client, err.link)
}
