package cache

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/repository"

	"github.com/google/uuid"
)

const (
	cacheInitCapacity = 10000
)

type userDTOWithTTL struct {
	user       *model.User
	lastUsedAt time.Time
}

type CacheDecorator struct {
	userRepo  repository.UserProvider
	mu        sync.RWMutex
	user      map[uuid.UUID]*userDTOWithTTL
	userLogin map[string]uuid.UUID
	done      chan struct{}
}

func NewDecorator(userRepo repository.UserProvider,
	cleanupInterval time.Duration,
	ttl time.Duration) (*CacheDecorator, error) {

	if userRepo == nil {
		return nil, errors.New("userRepo cannot be nil")
	}

	cache := &CacheDecorator{
		userRepo:  userRepo,
		user:      make(map[uuid.UUID]*userDTOWithTTL, cacheInitCapacity),
		userLogin: make(map[string]uuid.UUID, cacheInitCapacity),
	}

	cache.runJanitor(cleanupInterval, ttl)

	return cache, nil
}

func (c *CacheDecorator) runJanitor(cleanupInterval time.Duration, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				for key, val := range c.user {
					if val.lastUsedAt.Add(ttl).After(time.Now()) {
						delete(c.userLogin, val.user.Login)
						delete(c.user, key)
					}
				}
				c.mu.Unlock()
			case <-c.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *CacheDecorator) getUser(id uuid.UUID) (*userDTOWithTTL, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.user[id]
	return val, ok
}

func (c *CacheDecorator) getUserIDByLogin(login string) (uuid.UUID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.userLogin[login]
	return val, ok
}

func (c *CacheDecorator) setUser(id uuid.UUID, user *model.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.user[id] = &userDTOWithTTL{
		user:       user,
		lastUsedAt: time.Now(),
	}
	c.userLogin[user.Login] = id
}

func (c *CacheDecorator) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	if user, ok := c.getUser(userID); ok {
		user.lastUsedAt = time.Now()
		return user.user, nil
	}

	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		return user, err
	}

	c.setUser(userID, user)
	return user, nil
}

func (c *CacheDecorator) GetUserIDByLogin(ctx context.Context, login string) (*uuid.UUID, error) {
	if id, ok := c.getUserIDByLogin(login); ok {
		return &id, nil
	}

	id, err := c.userRepo.GetUserIDByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (c *CacheDecorator) UpdateUser(ctx context.Context, req model.UpdateUserRequest) (*uuid.UUID, error) {
	if user, ok := c.getUser(req.ID); ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		id, err := c.userRepo.UpdateUser(ctx, req)
		if err != nil {
			return nil, err
		}
		user.user, err = c.userRepo.GetUser(ctx, *id)
		if err != nil {
			return nil, err
		}
		user.lastUsedAt = time.Now()
		return id, nil
	}

	id, err := c.userRepo.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	user, err := c.userRepo.GetUser(ctx, *id)
	if err != nil {
		return nil, err
	}

	c.setUser(*id, user)
	return id, nil
}

func (c *CacheDecorator) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if _, ok := c.getUser(id); ok {
		c.mu.Lock()
		defer c.mu.Unlock()

		if err := c.userRepo.DeleteUser(ctx, id); err != nil {
			return err
		}
		delete(c.user, id)
	}

	return c.userRepo.DeleteUser(ctx, id)
}

func (c *CacheDecorator) AddUser(ctx context.Context, user model.User) error {
	return c.userRepo.AddUser(ctx, user)
}

func (c *CacheDecorator) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.done <- struct{}{}
	c.userRepo.Close()
}
