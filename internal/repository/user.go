package repository

import (
	"context"
	"log"

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

type UserProvider interface {
	UpdateUser(context.Context, model.UpdateUserRequest) (uuid.UUID, error)
	AddUser(context.Context, model.User) error
	GetUser(context.Context, uuid.UUID) (model.User, error)
	GetUserIDByLogin(context.Context, string) (uuid.UUID, error)
	DeleteUser(context.Context, uuid.UUID) error
	Close()
}

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
		return err
	}

	return nil
}

func (s *UserRepo) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	builder := squirrel.Select(idColumn, loginColumn, passwordColumn, nameColumn, ageColumn).
		From(tableName).
		Where(squirrel.Eq{idColumn: id}).
		PlaceholderFormat(squirrel.Dollar)

	var user model.User

	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.pool.QueryRow(ctx, query, args...)
	log.Println(query)
	log.Println(args)
	err = row.Scan(&user.ID, &user.Login, &user.Password, &user.Name, &user.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			user.ID = uuid.Nil
			return user, nil
		}

		return user, errors.Wrap(err, "GetUser Scan")
	}

	return user, nil
}

func (s *UserRepo) UpdateUser(ctx context.Context, toUpdate model.UpdateUserRequest) (uuid.UUID, error) {
	builder := squirrel.Update("users")
	log.Println(toUpdate)
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
	log.Println(query)
	log.Println(args)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "UpdateUser ToSql")
	}

	err = s.pool.QueryRow(ctx, query, args...).Scan(&id)
	log.Println(err)
	log.Println(id)
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
		Where(squirrel.Eq{loginColumn: login}).
		PlaceholderFormat(squirrel.Dollar)

	var id uuid.UUID

	query, args, err := builder.ToSql()
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "GetUser ToSql")
	}

	row := s.pool.QueryRow(ctx, query, args...)

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
		Where(squirrel.Eq{idColumn: id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "DeleteUser ToSql")
	}

	if _, err := s.pool.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (s *UserRepo) Close() {
	s.pool.Close()
}
