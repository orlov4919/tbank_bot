package tgbot

import "sync"

type userID = int

type State func(text string)

type usersState struct {
	mu      sync.Mutex
	storage map[userID]State
}
