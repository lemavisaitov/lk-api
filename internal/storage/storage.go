package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/model"
	"log"
)

type Storage struct {
	dbc *pgxpool.Pool
}

func GetConnect(ctx context.Context, connStr string) (handler.Storage, error) {
	dbc, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return &Storage{dbc: dbc}, nil
}

func (s *Storage) AddUser(ctx context.Context, user model.User) error {
	query := `
	INSERT INTO users (id, login, password, name, age)
	VALUES ($1, $2, $3, $4, $5);
	`

	args := []interface{}{
		user.ID.UUID,
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
		SELECT id, login, password, name, age FROM users
		WHERE id = $1;
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
		UPDATE users
	`
	query += " SET"

	args := make([]any, 0, 4)
	firstField := true
	if toUpdate.Name != "" {
		firstField = false
		query += fmt.Sprintf(" name = $%d", len(args)+1)
		args = append(args, toUpdate.Name)
	}
	if toUpdate.Age != 0 {
		if !firstField {
			query += ","
		}
		firstField = false
		query += fmt.Sprintf(" age = $%d", len(args)+1)
		args = append(args, toUpdate.Age)
	}
	if toUpdate.Password != "" {
		if !firstField {
			query += ","
		}
		query += fmt.Sprintf(" password = $%d", len(args)+1)
		args = append(args, toUpdate.Password)
	}

	var id uuid.NullUUID
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id;", len(args)+1)
	args = append(args, toUpdate.ID)
	log.Println(query)
	log.Println(args)
	err := s.dbc.QueryRow(ctx, query, args...).Scan(&id.UUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			id = uuid.NullUUID{}
			log.Printf("valid: %v, uuid: %s", id.Valid, id.UUID.String())
			return id, nil
		}
		log.Printf("valid: %v, uuid: %s", id.Valid, id.UUID.String())
		return id, err
	}
	id.Valid = true
	log.Printf("valid: %v, uuid: %s", id.Valid, id.UUID.String())
	return id, nil
}

func (s *Storage) GetUserIDByLogin(ctx context.Context, login string) (uuid.NullUUID, error) {
	query := `
		SELECT id FROM users WHERE login = $1;
	`

	row := s.dbc.QueryRow(ctx, query, login)

	var id uuid.NullUUID
	err := row.Scan(&id.UUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.NullUUID{}, nil
		}
		return uuid.NullUUID{}, err
	}
	id.Valid = true
	log.Printf("Return id: %s, valid: %v", id.UUID, id.Valid)
	return id, nil
}

func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users WHERE id = $1
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
