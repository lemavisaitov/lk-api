package cache

import (
	"context"
	"sync"
	"time"
	"unsafe"

	"github.com/lemavisaitov/lk-api/internal/logger"
	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/repository"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
		done:      make(chan struct{}),
	}

	cache.runCleaner(cleanupInterval, ttl)

	return cache, nil
}

func (c *CacheDecorator) runCleaner(cleanupInterval time.Duration, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for {
			select {
			case <-ticker.C:
				c.cleanExpired(ttl)
			case <-c.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *CacheDecorator) cleanExpired(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, val := range c.user {
		if val.lastUsedAt.Add(ttl).Before(time.Now()) {
			logger.Debug("cleanup expired user",
				zap.String("userID", key.String()),
				zap.String("user login", val.user.Login),
			)
			delete(c.userLogin, val.user.Login)
			delete(c.user, key)
		}
	}
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

func (c *CacheDecorator) deleteUser(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.userLogin, c.user[id].user.Login)
	delete(c.user, id)
}

func (c *CacheDecorator) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	if user, ok := c.getUser(userID); ok {
		user.lastUsedAt = time.Now()
		return user.user, nil
	}

	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		return user, errors.Wrap(err, "from GetUser in CacheDecorator")
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
		return nil, errors.Wrap(err, "from GetUserIDByLogin in CacheDecorator")
	}

	return id, nil
}

func (c *CacheDecorator) UpdateUser(ctx context.Context, req model.UpdateUserRequest) (*uuid.UUID, error) {
	id, err := c.userRepo.UpdateUser(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "from UpdateUser in CacheDecorator")
	}
	if id == nil {
		return nil, errors.New("user ID is empty")
	}

	user, err := c.userRepo.GetUser(ctx, *id)
	if err != nil {
		return nil, errors.Wrap(err, "from GetUser in CacheDecorator")
	}

	c.setUser(*id, user)
	return id, nil
}

func (c *CacheDecorator) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if _, ok := c.getUser(id); ok {
		c.deleteUser(id)
	}
	err := c.userRepo.DeleteUser(ctx, id)
	return errors.Wrap(err, "from DeleteUser in CacheDecorator")
}

func (c *CacheDecorator) AddUser(ctx context.Context, user model.User) error {
	err := c.userRepo.AddUser(ctx, user)
	return errors.Wrap(err, "from AddUser in CacheDecorator")
}

func (c *CacheDecorator) Close() {
	close(c.done)
}

func (c *CacheDecorator) copy() *CacheDecorator {
	c.mu.RLock()
	defer c.mu.RUnlock()

	replica := &CacheDecorator{
		user:      make(map[uuid.UUID]*userDTOWithTTL, len(c.user)),
		userLogin: make(map[string]uuid.UUID, len(c.userLogin)),
		done:      make(chan struct{}),
	}

	// Копируем карту user
	for k, v := range c.user {
		userCopy := &userDTOWithTTL{
			user:       &model.User{},
			lastUsedAt: v.lastUsedAt,
		}
		*userCopy.user = *v.user // Глубокая копия структуры User
		replica.user[k] = userCopy
	}

	// Копируем карту userLogin
	for k, v := range c.userLogin {
		replica.userLogin[k] = v
	}

	return replica
}

func (c *CacheDecorator) MemoryUsage() uint64 {
	replica := c.copy()

	var size uint64

	// Размер самой структуры
	size += uint64(unsafe.Sizeof(*replica))

	// Размер карты user (хеш-таблица + ключи + значения)
	size += uint64(unsafe.Sizeof(replica.user))
	for k, v := range replica.user {
		size += uint64(unsafe.Sizeof(k)) + uint64(unsafe.Sizeof(*v))
		if v.user != nil {
			size += uint64(unsafe.Sizeof(*v.user)) // Размер userDTOWithTTL
			size += uint64(len(v.user.Login))
			size += uint64(len(v.user.Password))
			size += uint64(len(v.user.Name))
		}
	}

	// Размер карты userLogin (ключи строки + UUID)
	size += uint64(unsafe.Sizeof(replica.userLogin))
	for k, v := range replica.userLogin {
		size += uint64(len(k)) + uint64(unsafe.Sizeof(v)) // Строка (длина) + UUID
	}

	return size
}
