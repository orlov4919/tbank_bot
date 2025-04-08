package tgnotifier_test

import (
	"linkTraccer/internal/application/scrapper/notifiers/mocks"
	"linkTraccer/internal/application/scrapper/notifiers/tgnotifier"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/scrapper"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	errRepo     = errors.New("ошибка в репозитории")
	errClient   = errors.New("ошибка в клиенте")
	linkInfo    = &scrapper.LinkInfo{ID: 1, URL: "github.com", LastUpdate: time.Now()}
	linkUpdates = scrapper.LinkUpdates{&scrapper.LinkUpdate{}}
)

func TestTgNotifier_SendUpdates(t *testing.T) {
	repoWithErr := mocks.NewUserRepo(t)
	repoWithoutErr := mocks.NewUserRepo(t)
	botClientWithErr := mocks.NewBotClient(t)
	botClientWithoutErr := mocks.NewBotClient(t)

	repoWithoutErr.On("UsersWhoTrackLink", mock.Anything).Return(nil, nil)
	repoWithErr.On("UsersWhoTrackLink", mock.Anything).Return(nil, errRepo)
	botClientWithErr.On("SendLinkUpdates", mock.Anything).Return(errClient)
	botClientWithoutErr.On("SendLinkUpdates", mock.Anything).Return(nil)

	type TestCase struct {
		name      string
		botClient tgnotifier.BotClient
		userRepo  scrapservice.UserRepo
		correct   bool
	}

	tests := []TestCase{
		{
			name:      "ошибка при получении всех пользователей отслеживающих ссылку",
			botClient: botClientWithErr,
			userRepo:  repoWithErr,
			correct:   false,
		},
		{
			name:      "ошибка при отправке апдейтов",
			botClient: botClientWithErr,
			userRepo:  repoWithoutErr,
			correct:   false,
		},
		{
			name:      "апдейты успешно отправлены",
			botClient: botClientWithoutErr,
			userRepo:  repoWithoutErr,
			correct:   true,
		},
	}

	for _, test := range tests {
		notifier := tgnotifier.New(test.userRepo, test.botClient)

		err := notifier.SendUpdates(linkInfo, linkUpdates)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
