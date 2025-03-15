package repository

import (
	"context"

	"github.com/lemavisaitov/lk-api/internal/apperr"
	"github.com/lemavisaitov/lk-api/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserProvider(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		pool: pool,
	}
}

func (s *UserRepo) AddUser(ctx context.Context, user model.User) error {
	builder := squirrel.Insert(tableName).
		Columns(idColumn, loginColumn, passwordColumn, nameColumn, ageColumn).
		Values(user.ID, user.Login, user.Password, user.Name, user.Age).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "AddUser ToSql")
	}

	_, err = s.pool.Exec(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "AddUser Exec")
	}

	return nil
}

func (s *UserRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	builder := squirrel.Select(idColumn, loginColumn, passwordColumn, nameColumn, ageColumn).
		From(tableName).
		Where(squirrel.Eq{idColumn: id}).
		PlaceholderFormat(squirrel.Dollar)

	var user model.User

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.pool.QueryRow(ctx, query, args...)
	err = row.Scan(&user.ID, &user.Login, &user.Password, &user.Name, &user.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			user.ID = uuid.Nil
			return nil, errors.Wrap(apperr.ErrNotFound, "id not found, GetUser repository")
		}

		return nil, errors.Wrap(err, "GetUser Scan")
	}

	return &user, nil
}

func (s *UserRepo) UpdateUser(ctx context.Context, toUpdate model.UpdateUserRequest) (*uuid.UUID, error) {
	builder := squirrel.Update("users")
	if toUpdate.Name != "" {
		builder = builder.Set(nameColumn, toUpdate.Name)
	}
	if toUpdate.Age != 0 {
		builder = builder.Set(ageColumn, toUpdate.Age)
	}
	if toUpdate.Password != "" {
		builder = builder.Set(passwordColumn, toUpdate.Password)
	}
	builder = builder.Where(squirrel.Eq{idColumn: toUpdate.ID}).
		Suffix("RETURNING " + idColumn).
		PlaceholderFormat(squirrel.Dollar)

	var id uuid.UUID

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "UpdateUser ToSql")
	}

	err = s.pool.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(apperr.ErrNotFound, "User ID not found")
		}
		return nil, errors.Wrap(err, "UpdateUser Scan")
	}

	return &id, nil
}

func (s *UserRepo) GetUserIDByLogin(ctx context.Context, login string) (*uuid.UUID, error) {
	builder := squirrel.Select(idColumn).
		From(tableName).
		Where(squirrel.Eq{loginColumn: login}).
		PlaceholderFormat(squirrel.Dollar)

	var id uuid.UUID

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.pool.QueryRow(ctx, query, args...)

	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(apperr.ErrNotFound, "login not found")
		}
		return nil, errors.Wrap(err, "GetUser Scan")
	}

	return &id, nil
}

func (s *UserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	builder := squirrel.Delete(tableName).
		Where(squirrel.Eq{idColumn: id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "DeleteUser ToSql")
	}

	if _, err := s.pool.Exec(ctx, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.Wrap(apperr.ErrNotFound, "User ID not found")
		}
		return errors.Wrap(err, "DeleteUser Exec")
	}

	return nil
}
