package tgnotifier

import (
	"fmt"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
)

const (
	descriptionFormat = "Пришло новое уведомление 🔥\n\nСобытие: %s\nПользователь: %s\nВремя создания: %s\nПревью: %s\n\n"
)

type LinkInfo = scrapper.LinkInfo
type LinkUpdates = scrapper.LinkUpdates
type LinkUpdate = scrapper.LinkUpdate

type BotClient interface {
	SendLinkUpdates(update *dto.LinkUpdate) error
}

type TgNotifier struct {
	botClient BotClient
	userRepo  scrapservice.UserRepo
}

func New(userRepo scrapservice.UserRepo, botClient BotClient) *TgNotifier {
	return &TgNotifier{
		botClient: botClient,
		userRepo:  userRepo,
	}
}

func (t *TgNotifier) SendUpdates(linkInfo *LinkInfo, linkUpdates LinkUpdates) error {
	users, err := t.userRepo.UsersWhoTrackLink(linkInfo.ID)

	if err != nil {
		return fmt.Errorf("при получении всех пользователей произошла ошибка : %w", err)
	}

	for _, update := range linkUpdates {

		err := t.botClient.SendLinkUpdates(&dto.LinkUpdate{
			ID:          linkInfo.ID,
			URL:         linkInfo.URL,
			Description: t.notifyMsg(update),
			TgChatIDs:   users})

		if err != nil {
			return fmt.Errorf("не удалось отправить обновление ссылки  : %w", err)
		}
	}

	return nil
}

func (t *TgNotifier) notifyMsg(linkUpdate *LinkUpdate) string {
	return fmt.Sprintf(descriptionFormat, linkUpdate.Header, linkUpdate.UserName,
		linkUpdate.CreateTime, linkUpdate.Preview)
}
