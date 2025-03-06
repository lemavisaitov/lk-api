package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lemavisaitov/lk-api/internal/model"
)

type Storage struct {
	dbc *pgxpool.Pool
}

func GetConnect(ctx context.Context, connStr string) (*Storage, error) {
	dbc, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &Storage{dbc: dbc}, nil
}

func (s *Storage) AddUser(ctx context.Context, user model.User) error {
	query := `
	INSERT INTO user (id, login, password, name, age)
	VALUES ($1, $2, $3, $4, $5)
	`

	args := []interface{}{
		user.ID,
		user.Login,
		user.Password,
		user.Name,
		user.Age,
	}
	_, err := s.dbc.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	query := `
		SELECT id, login, password, name, age FROM user
		WHERE id = $1
	`

	row := s.dbc.QueryRow(ctx, query, id)
	var user model.User
	err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Name, &user.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			user.ID = uuid.NullUUID{}
			return user, nil
		}

		return model.User{}, err
	}

	return user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, toUpdate model.UpdateUserRequest) (uuid.NullUUID, error) {
	query := `
		UPDATE user
	`
	query += " SET"

	args := make([]interface{}, 0, 3)
	if toUpdate.Name != "" {
		query += fmt.Sprintf(" name = $%d", len(args)+1)
		args = append(args, toUpdate.Name)
	}
	if toUpdate.Age != 0 {
		query += fmt.Sprintf(" age = $%d", len(args)+1)
		args = append(args, toUpdate.Age)
	}
	if toUpdate.Password != "" {
		query += fmt.Sprintf(" password = $%d", len(args)+1)
		args = append(args, toUpdate.Password)
	}

	var id uuid.NullUUID
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id", len(args)+1)
	err := s.dbc.QueryRow(ctx, query, toUpdate.ID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			id = uuid.NullUUID{}
			return id, nil
		}
		return id, err
	}

	return id, nil
}

func (s *Storage) GetUserIDByLogin(ctx context.Context, login string) (uuid.NullUUID, error) {
	query := `
		SELECT id FROM user WHERE login = $1
	`

	row := s.dbc.QueryRow(ctx, query, login)

	var id uuid.NullUUID
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.NullUUID{}, nil
		}
		return uuid.NullUUID{}, err
	}
	return id, nil
}

func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM user WHERE id = $1
	`

	_, err := s.dbc.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Close() {
	s.dbc.Close()
}
