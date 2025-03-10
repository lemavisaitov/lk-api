package cache

import (
	"context"
	"sync"
	"time"

	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/repository"

	"github.com/google/uuid"
)

type userDTOWithTTL struct {
	user      model.User
	lastUsage int64
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
	ttl time.Duration) *CacheDecorator {

	cache := &CacheDecorator{
		userRepo:  userRepo,
		user:      make(map[uuid.UUID]*userDTOWithTTL, 10000),
		userLogin: make(map[string]uuid.UUID, 10000),
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for {
			select {
			case <-ticker.C:
				cache.mu.Lock()
				for key, val := range cache.user {
					if time.Now().Unix()-val.lastUsage >= int64(ttl.Seconds()) {
						delete(cache.userLogin, val.user.Login)
						delete(cache.user, key)
					}
				}
				cache.mu.Unlock()
			case <-cache.done:
				ticker.Stop()
				return
			}
		}
	}()

	return cache
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
		user:      *user,
		lastUsage: time.Now().Unix(),
	}
	c.userLogin[user.Login] = id
}

func (c *CacheDecorator) GetUser(ctx context.Context, userID uuid.UUID) (model.User, error) {
	if user, ok := c.getUser(userID); ok {
		user.lastUsage = time.Now().Unix()
		return user.user, nil
	}

	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		return user, err
	}

	c.setUser(userID, &user)
	return user, nil
}

func (c *CacheDecorator) GetUserIDByLogin(ctx context.Context, login string) (uuid.UUID, error) {
	if id, ok := c.getUserIDByLogin(login); ok {
		return id, nil
	}

	id, err := c.userRepo.GetUserIDByLogin(ctx, login)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (c *CacheDecorator) UpdateUser(ctx context.Context, req model.UpdateUserRequest) (uuid.UUID, error) {
	if user, ok := c.getUser(req.ID); ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		id, err := c.userRepo.UpdateUser(ctx, req)
		if err != nil {
			return uuid.Nil, err
		}
		user.user, err = c.userRepo.GetUser(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		user.lastUsage = time.Now().Unix()
		return id, nil
	}

	id, err := c.userRepo.UpdateUser(ctx, req)
	if err != nil {
		return uuid.Nil, err
	}
	user, err := c.userRepo.GetUser(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}

	c.setUser(id, &user)
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
