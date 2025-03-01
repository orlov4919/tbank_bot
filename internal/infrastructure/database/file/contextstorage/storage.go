package contextstorage

import (
	"linkTraccer/internal/domain/tgbot"
	"sync"
)

type ID = tgbot.ID
type ContextData = tgbot.ContextData

type ContextStorage struct {
	mu      sync.Mutex
	context map[ID]*ContextData
}

func New() *ContextStorage {
	return &ContextStorage{
		context: make(map[ID]*ContextData),
	}
}

func (c *ContextStorage) RegUser(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.context[id]; ok {
		return NewErrUserAlreadyReg(id)
	}

	c.context[id] = &ContextData{}

	return nil
}

func (c *ContextStorage) AddURL(id ID, url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.context[id]; !ok {
		return NewErrUserNotReg(id)
	}

	c.context[id].URL = url

	return nil
}

func (c *ContextStorage) AddFilters(id ID, filters []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.context[id]; !ok {
		return NewErrUserNotReg(id)
	}

	c.context[id].Filters = filters

	return nil
}

func (c *ContextStorage) AddTags(id ID, tags []string) error {
	if _, ok := c.context[id]; !ok {
		return NewErrUserNotReg(id)
	}

	c.context[id].Tags = tags

	return nil
}

func (c *ContextStorage) ResetCtx(id ID) error {
	if _, ok := c.context[id]; !ok {
		return NewErrUserNotReg(id)
	}

	c.context[id] = &ContextData{}

	return nil
}

func (c *ContextStorage) UserContext(id ID) (*ContextData, error) {
	if _, ok := c.context[id]; !ok {
		return nil, NewErrUserNotReg(id)
	}

	return c.context[id], nil
}
