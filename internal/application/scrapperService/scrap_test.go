package scrapperService_test

import (
	"github.com/stretchr/testify/mock"
	"linkTraccer/internal/application/scrapperService"
	"linkTraccer/internal/application/scrapperService/mocks"
	"linkTraccer/internal/domain/scrapper"
	"testing"
)

type Link = scrapper.Link
type User = scrapper.User
type LinkState = scrapper.LinkState

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

func TestScrapper_CheckLinkUpdates(t *testing.T) {

	gitClient := mocks.NewSiteClient(t)
	userRepo := mocks.NewUserRepo(t)
	botClient := mocks.NewBotClient(t)

	gitClient.On("CanTrack", firstSaveLink).Return(true).Times(1)
	gitClient.On("CanTrack", secondSaveLink).Return(true).Times(1)
	gitClient.On("CanTrack", thirdSaveLink).Return(false).Times(1)

	gitClient.On("LinkState", firstSaveLink).Return(firstLinkNewState, nil).Times(1)
	gitClient.On("LinkState", secondSaveLink).Return(secondLinkNewState, nil).Times(1)

	userRepo.On("AllLinks").Return([]Link{firstSaveLink, secondSaveLink, thirdSaveLink}).Times(1)
	userRepo.On("LinkState", firstSaveLink).Return(firstLinkOldState, nil).Times(1)
	userRepo.On("LinkState", secondSaveLink).Return(secondLinkOldState, nil).Times(1)

	userRepo.On("ChangeLinkState", secondSaveLink, secondLinkNewState).Return(nil).Times(1)
	userRepo.On("ChangeLinkState", firstSaveLink, firstLinkNewState).Return(nil).Times(1)

	userRepo.On("UsersWhoTrackLink", firstSaveLink).Return(userWhoTrackFirstLink).Times(1)

	botClient.On("SendLinkUpdates", mock.Anything).Return(nil).Times(1)

	scrapper := scrapperService.New(userRepo, botClient, gitClient)

	scrapper.CheckLinkUpdates()
}
