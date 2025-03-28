package scrapservice_test

import (
	"io"
	"linkTraccer/internal/application/scrapper/scrapservice"
	mocks2 "linkTraccer/internal/application/scrapper/scrapservice/mocks"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
)

type Link = scrapservice.Link
type User = scrapservice.User
type LinkState = scrapservice.LinkState

const (
	firstSaveLink  Link = "https://github.com/orlov4919/test"
	secondSaveLink Link = "https://github.com/orlov4919/testt2"
	thirdSaveLink  Link = "https://stackoverflow.com/"

	firstLinkOldState  LinkState = "oldState"
	firstLinkNewState  LinkState = "newState"
	secondLinkOldState LinkState = ""
	secondLinkNewState LinkState = "newState"
)

var userWhoTrackFirstLink = []User{1}
var logLevel = slog.LevelInfo
var logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: logLevel}))

func TestScrapper_CheckLinkUpdates(t *testing.T) {
	gitClient := mocks2.NewSiteClient(t)
	userRepo := mocks2.NewUserRepo(t)
	botClient := mocks2.NewBotClient(t)

	gitClient.On("CanTrack", firstSaveLink).Return(true).Times(1)
	gitClient.On("CanTrack", secondSaveLink).Return(true).Times(1)
	gitClient.On("CanTrack", thirdSaveLink).Return(false).Times(1)

	gitClient.On("LinkUpdates", firstSaveLink).Return(firstLinkNewState, nil).Times(1)
	gitClient.On("LinkUpdates", secondSaveLink).Return(secondLinkNewState, nil).Times(1)

	userRepo.On("LinksBatch").Return([]Link{firstSaveLink, secondSaveLink, thirdSaveLink}).Times(1)
	userRepo.On("LinkUpdates", firstSaveLink).Return(firstLinkOldState, nil).Times(1)
	userRepo.On("LinkUpdates", secondSaveLink).Return(secondLinkOldState, nil).Times(1)

	userRepo.On("ChangeLinkLastCheck", secondSaveLink, secondLinkNewState).Return(nil).Times(1)
	userRepo.On("ChangeLinkLastCheck", firstSaveLink, firstLinkNewState).Return(nil).Times(1)

	userRepo.On("UsersWhoTrackLink", firstSaveLink).Return(userWhoTrackFirstLink).Times(1)

	botClient.On("SendLinkUpdates", mock.Anything).Return(nil).Times(1)

	scrapper := scrapservice.New(userRepo, botClient, logger, gitClient)

	scrapper.CheckLinksUpdates()
}
