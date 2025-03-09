package repository

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lemavisaitov/lk-api/internal/model"
	"github.com/pkg/errors"
)

const (
	tableName      = "users"
	idColumn       = "id"
	loginColumn    = "login"
	passwordColumn = "password"
	nameColumn     = "name"
	ageColumn      = "age"
)

type UserProvider interface {
	UpdateUser(context.Context, model.UpdateUserRequest) (uuid.UUID, error)
	AddUser(context.Context, model.User) error
	GetUser(context.Context, uuid.UUID) (model.User, error)
	GetUserIDByLogin(context.Context, string) (uuid.UUID, error)
	DeleteUser(context.Context, uuid.UUID) error
	Close()
}

type UserRepo struct {
	conn *pgxpool.Pool
}

func (s *UserRepo) AddUser(ctx context.Context, user model.User) error {
	builder := squirrel.Insert(tableName).
		Columns(idColumn, loginColumn, passwordColumn, nameColumn, ageColumn).
		Values(user.ID, user.Login, user.Password, user.Name).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "AddUser ToSql")
	}

	_, err = s.conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserRepo) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	builder := squirrel.Select(idColumn, loginColumn, nameColumn, ageColumn).
		From(tableName).
		Where(squirrel.Eq{idColumn: id})

	var user model.User

	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.conn.QueryRow(ctx, query, args)

	err = row.Scan(&user.ID, &user.Login, &user.Name, &user.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			user.ID = uuid.NullUUID{}
			return user, nil
		}

		return user, errors.Wrap(err, "GetUser Scan")
	}

	return user, nil
}

func (s *UserRepo) UpdateUser(ctx context.Context, toUpdate model.UpdateUserRequest) (uuid.UUID, error) {
	builder := squirrel.Update("users")

	if toUpdate.Name != "" {
		builder.Set("name", toUpdate.Name)
	}
	if toUpdate.Age != 0 {
		builder.Set("age", toUpdate.Age)
	}
	if toUpdate.Password != "" {
		builder.Set("password", toUpdate.Password)
	}
	builder.Where("id", "=", toUpdate.ID).
		PlaceholderFormat(squirrel.Dollar)

	var id uuid.UUID

	query, args, err := builder.ToSql()
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "UpdateUser ToSql")
	}

	err = s.conn.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, errors.Wrap(err, "UpdateUser Scan")
	}

	return id, nil
}

func (s *UserRepo) GetUserIDByLogin(ctx context.Context, login string) (uuid.UUID, error) {
	builder := squirrel.Select(idColumn).
		From(tableName).
		Where(squirrel.Eq{loginColumn: login})

	var id uuid.UUID

	query, args, err := builder.ToSql()
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.conn.QueryRow(ctx, query, args)

	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, errors.Wrap(err, "GetUser Scan")
	}

	return id, nil
}

func (s *UserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	builder := squirrel.Delete(tableName).
		Where(squirrel.Eq{idColumn: id})

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "DeleteUser ToSql")
	}

	if _, err := s.conn.Exec(ctx, query, args); err != nil {
		return err
	}

	return nil
}

func (s *UserRepo) Close() {
	s.conn.Close()
}
