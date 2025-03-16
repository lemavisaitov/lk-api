package cache

import (
	"context"
	"testing"
	"time"

	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/testutils/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
	// Создаем контроллер мока
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mocks.NewMockUserProvider(ctrl)

	// Создаем кэш
	cache, err := NewDecorator(mockRepo, time.Minute, time.Minute)
	require.NoError(t, err)

	// Тестовые данные
	userID := uuid.New()
	user := &model.User{
		ID:    userID,
		Name:  "John Doe",
		Login: "johndoe",
	}

	// Кейс 1: Пользователь есть в кэше
	cache.setUser(userID, user)
	cachedUser, err := cache.GetUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, user, cachedUser)

	// Кейс 2: Пользователя нет в кэше, но он есть в репозитории
	mockRepo.EXPECT().
		GetUser(context.Background(), userID).
		Return(user, nil)

	cache.deleteUser(userID) // Удаляем из кэша
	retrievedUser, err := cache.GetUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
}

func TestGetUserIDByLogin(t *testing.T) {
	// Создаем контроллер мока
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mocks.NewMockUserProvider(ctrl)

	// Создаем кэш
	cache, err := NewDecorator(mockRepo, time.Minute, time.Minute)
	require.NoError(t, err)

	// Тестовые данные
	login := "johndoe"
	userID := uuid.New()

	// Кейс 1: ID пользователя есть в кэше
	cache.setUser(userID, &model.User{
		Login: login,
		ID:    userID,
	})
	id, err := cache.GetUserIDByLogin(context.Background(), login)

	require.NoError(t, err)
	assert.Equal(t, userID.String(), id.String())

	// Кейс 2: ID пользователя нет в кэше, но он есть в репозитории
	mockRepo.EXPECT().
		GetUserIDByLogin(context.Background(), login).
		Return(&userID, nil)

	cache.deleteUser(userID) // Удаляем из кэша
	retrievedID, err := cache.GetUserIDByLogin(context.Background(), login)
	require.NoError(t, err)
	assert.Equal(t, userID, *retrievedID)
}

func TestUpdateUser(t *testing.T) {
	// Создаем контроллер мока
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mocks.NewMockUserProvider(ctrl)

	// Создаем кэш
	cache, err := NewDecorator(mockRepo, time.Minute, time.Minute)
	require.NoError(t, err)

	// Тестовые данные
	userID := uuid.New()
	req := model.UpdateUserRequest{
		ID:       userID,
		Name:     "Updated Name",
		Password: "newpassword",
	}
	updatedUser := &model.User{
		ID:       userID,
		Name:     req.Name,
		Password: req.Password,
		Login:    "johndoe",
	}

	// Настройка мока
	mockRepo.EXPECT().
		UpdateUser(context.Background(), req).
		Return(&userID, nil)
	mockRepo.EXPECT().
		GetUser(context.Background(), userID).
		Return(updatedUser, nil)

	// Вызываем метод UpdateUser
	id, err := cache.UpdateUser(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, userID, *id)

	// Проверяем, что пользователь обновлен в кэше
	cachedUser, _ := cache.getUser(userID)
	assert.Equal(t, updatedUser, cachedUser.user)
}

func TestAddUser(t *testing.T) {
	// Создаем контроллер мока
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mocks.NewMockUserProvider(ctrl)

	// Создаем кэш
	cache, err := NewDecorator(mockRepo, time.Minute, time.Minute)
	require.NoError(t, err)

	// Тестовые данные
	user := model.User{
		ID:    uuid.New(),
		Name:  "John Doe",
		Login: "johndoe",
	}

	// Настройка мока
	mockRepo.EXPECT().
		AddUser(context.Background(), user).
		Return(nil)

	// Вызываем метод AddUser
	err = cache.AddUser(context.Background(), user)
	require.NoError(t, err)
}

func TestClose(t *testing.T) {
	// Создаем контроллер мока
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mocks.NewMockUserProvider(ctrl)

	// Создаем кэш
	cache, err := NewDecorator(mockRepo, time.Minute, time.Minute)
	require.NoError(t, err)

	// Закрываем кэш
	go func() {
		cache.Close()
	}()

	time.Sleep(time.Millisecond * 1)
	// Проверяем, что канал закрыт
	select {
	case <-cache.done:
		// Ожидаем, что канал закрыт
	default:
		t.Fatal("expected done channel to be closed")
	}
}
