//go:build integration
// +build integration

package usecase

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lemavisaitov/lk-api/internal/apperr"
	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/lemavisaitov/lk-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestUserCase_AddUser(t *testing.T) {
	testCases := []struct {
		valid    bool
		caseName string
		login    string
		password string
		name     string
		age      int
	}{
		{
			valid:    true,
			caseName: "valid test",
			login:    "qwerty1",
			password: "qwerty",
			name:     "qwerty",
			age:      18,
		}, {
			valid:    false,
			caseName: "invalid test: age less then 0",
			login:    "qwerty2",
			password: "qwerty",
			name:     "qwerty",
			age:      -1,
		},
		{
			valid:    false,
			caseName: "invalid test: login already exists",
			login:    "qwerty1",
			password: "qwerty",
			name:     "qwerty",
			age:      18,
		},
		{
			valid:    false,
			caseName: "invalid test: empty login",
			login:    "",
			password: "qwerty",
			name:     "qwerty",
			age:      18,
		},
	}
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserProvider(pool)
	userUC := NewUserProvider(userRepo)
	for _, tc := range testCases {
		id := uuid.New()
		user := model.User{
			ID:       id,
			Login:    tc.login,
			Password: tc.password,
			Name:     tc.name,
			Age:      tc.age,
		}
		t.Run(tc.caseName, func(t *testing.T) {
			storedID, err := userUC.AddUser(&gin.Context{}, user)
			if tc.valid {
				require.NoError(t, err)
				assert.Equal(t, id.String(), storedID.String())
			}
			require.Error(t, err)
		})
	}
}

func TestUserCase_DeleteUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserProvider(pool)
	userUC := NewUserProvider(userRepo)

	user := model.User{
		ID:       uuid.New(),
		Login:    "login",
		Password: "password",
		Name:     "name",
		Age:      18,
	}

	t.Run("add user", func(t *testing.T) {
		storedID, err := userUC.AddUser(&gin.Context{}, user)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), storedID.String())
	})
	t.Run("delete user", func(t *testing.T) {
		err := userUC.DeleteUser(&gin.Context{}, user.ID)
		require.NoError(t, err)
	})
	t.Run("get user", func(t *testing.T) {
		_, err := userUC.GetUser(&gin.Context{}, user.ID)
		require.ErrorIs(t, err, apperr.ErrNotFound)
	})
}

func TestUserCase_GetUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	userRepo := repository.NewUserProvider(pool)
	userUC := NewUserProvider(userRepo)

	user := model.User{
		ID:       uuid.New(),
		Login:    "login",
		Password: "password",
		Name:     "name",
		Age:      18,
	}

	t.Run("add user", func(t *testing.T) {
		storedID, err := userUC.AddUser(&gin.Context{}, user)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), storedID.String())
	})

	testCases := []struct {
		valid    bool
		caseName string
		id       string
	}{
		{
			valid:    true,
			caseName: "valid test",
			id:       user.ID.String(),
		},
		{
			valid:    false,
			caseName: "invalid test: id does not exist",
			id:       uuid.New().String(),
		},
		{
			valid:    false,
			caseName: "invalid test: invalid id",
			id:       "uuid.New().String()",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			user, err := userUC.GetUser(&gin.Context{}, user.ID)
			if tc.valid {
				require.NoError(t, err)
				assert.Equal(t, user.ID.String(), tc.id)
			}
			require.Error(t, err)
		})
	}
}

func TestUserCase_GetUserIDByLogin(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	userRepo := repository.NewUserProvider(pool)
	userUC := NewUserProvider(userRepo)
	user := model.User{
		ID:       uuid.New(),
		Login:    "login",
		Password: "password",
	}
	t.Run("add user", func(t *testing.T) {
		storedID, err := userUC.AddUser(&gin.Context{}, user)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), storedID.String())
	})

	testCases := []struct {
		valid    bool
		caseName string
		login    string
	}{
		{
			valid:    true,
			caseName: "valid test",
			login:    "login",
		},
		{
			valid:    false,
			caseName: "invalid test: login does not exist",
			login:    "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			id, err := userUC.GetUserIDByLogin(&gin.Context{}, user.Login)
			if tc.valid {
				require.NoError(t, err)
				assert.Equal(t, id.String(), user.ID.String())
			}
			require.Error(t, err)
		})
	}
}

func TestUserCase_UpdateUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	userRepo := repository.NewUserProvider(pool)
	userUC := NewUserProvider(userRepo)
	user := model.User{
		ID:       uuid.New(),
		Login:    "login1",
		Password: "password",
	}
	t.Run("add user", func(t *testing.T) {
		storedID, err := userUC.AddUser(&gin.Context{}, user)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), storedID.String())
	})

	testCases := []struct {
		valid    bool
		caseName string
		req      model.UpdateUserRequest
	}{
		{
			valid:    true,
			caseName: "valid test",
			req: model.UpdateUserRequest{
				ID:       user.ID,
				Password: "password2",
			},
		},
		{
			valid:    false,
			caseName: "invalid test: age less then 0",
			req: model.UpdateUserRequest{
				ID:  user.ID,
				Age: -1,
			},
		},
		{
			valid:    false,
			caseName: "invalid test: id doesnt exist",
			req: model.UpdateUserRequest{
				ID: user.ID,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			storedID, err := userUC.UpdateUser(&gin.Context{}, tc.req)
			if tc.valid {
				require.NoError(t, err)
				assert.Equal(t, tc.req.ID.String(), storedID.String())
			}
			require.Error(t, err)
		})
	}
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	connStr := "host=postgres user=postgres password=postgres dbname=testdb sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), "DELETE FROM users")
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
	}
	return pool, cleanup
}
