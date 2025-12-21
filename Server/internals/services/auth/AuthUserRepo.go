package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type AuthProvider string

const (
	ProviderCredentials AuthProvider = "credentials"
	ProviderGoogle      AuthProvider = "google"
	ProviderGithub      AuthProvider = "github"
)

type User struct {
	Id       uuid.UUID    `json:"id"`
	Type     string       `json:"type"`
	Name     string       `json:"name"`
	Email    string       `json:"email"`
	Verified bool         `json:"verified"`
	Provider AuthProvider `json:"provider"`
}

type userRepoServices struct {
	db *pgxpool.Pool
}

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string, provider AuthProvider) (*User, *string, utils.ErrorType, error)
	GetUserByID(ctx context.Context, id string) (*User, string, utils.ErrorType, error)
	CreateUserCredentials(ctx context.Context, email, name, passwordHash string) (*User, utils.ErrorType, error)
	DeleteUser(ctx context.Context, id string) (bool, utils.ErrorType, error)
	UpdatePassword(ctx context.Context, id string, newPassword string) (bool, utils.ErrorType, error)
	UpdateUserVerification(ctx context.Context, userID string) (utils.ErrorType, error)
}

func NewUserRepo(pgDb *pgxpool.Pool) UserRepo {
	return &userRepoServices{db: pgDb}
}

func (r *userRepoServices) GetUserByEmail(ctx context.Context, email string, provider AuthProvider) (*User, *string, utils.ErrorType, error) {
	var user User
	var password string
	err := r.db.QueryRow(
		ctx,
		`SELECT id, type, name, email, password_hash, verified, provider
		 FROM users
		 WHERE email = $1 AND provider = $2`,
		email,
		provider,
	).Scan(
		&user.Id,
		&user.Type,
		&user.Name,
		&user.Email,
		&password,
		&user.Verified,
		&user.Provider,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, utils.ErrNotFound, errors.New("user not found")
		}
		return nil, nil, utils.ErrInternal, err
	}

	return &user, &password, utils.NoError, nil
}

func (r *userRepoServices) GetUserByID(ctx context.Context, id string) (*User, string, utils.ErrorType, error) {
	var user User
	var password string
	err := r.db.QueryRow(
		ctx,
		`SELECT id, type, name, email, password_hash, verified, provider
		 FROM users
		 WHERE id = $1 `,
		id,
	).Scan(
		&user.Id,
		&user.Type,
		&user.Name,
		&user.Email,
		&password,
		&user.Verified,
		&user.Provider,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", utils.ErrNotFound, errors.New("user not found")
		}
		return nil, "", utils.ErrInternal, err
	}

	return &user, password, utils.NoError, nil
}

func (r *userRepoServices) CreateUserCredentials(
	ctx context.Context,
	email string,
	name string,
	passwordHash string,
) (*User, utils.ErrorType, error) {

	var user User

	query := `
	INSERT INTO users (
		id,
		name,
		email,
		provider,
		password_hash
	)
	VALUES ($1, $2, $3, 'credentials', $4)
	RETURNING
		id,
		type,
		name,
		email,
		verified,
		provider;
	`

	err := r.db.QueryRow(
		ctx,
		query,
		uuid.New(),
		name,
		email,
		passwordHash,
	).Scan(
		&user.Id,
		&user.Type,
		&user.Name,
		&user.Email,
		&user.Verified,
		&user.Provider,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, utils.ErrConflict, errors.New("User already exsits")
			}
		}
		return nil, utils.ErrInternal, err
	}

	return &user, utils.NoError, nil
}

func (r *userRepoServices) DeleteUser(
	ctx context.Context,
	id string,
) (bool, utils.ErrorType, error) {

	userId, err := uuid.Parse(id)
	if err != nil {
		return false, utils.ErrBadRequest, errors.New("invalid user id")
	}

	cmd, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userId)
	if err != nil {
		return false, utils.ErrInternal, err
	}

	if cmd.RowsAffected() == 0 {
		return false, utils.ErrNotFound, errors.New("user not found")
	}

	return true, utils.NoError, nil
}

func (r *userRepoServices) UpdatePassword(ctx context.Context, id string, newPassword string) (bool, utils.ErrorType, error) {
	userId, err := uuid.Parse(id)
	if err != nil {
		return false, utils.ErrBadRequest, errors.New("please provide a valid user id to update the password ")
	}
	query := `UPDATE users SET password_hash = $1 , updated_at = NOW() WHERE id = $2`

	cmd, err := r.db.Exec(ctx, query, newPassword, userId)
	if err != nil {
		return false, utils.ErrInternal, err
	}

	if cmd.RowsAffected() == 0 {
		return false, utils.ErrNotFound, errors.New("user not found")
	}

	return true, utils.NoError, nil
}

func (r *userRepoServices) UpdateUserVerification(ctx context.Context, userID string) (utils.ErrorType, error) {
	query := `
	UPDATE users 
	SET verified = TRUE, updated_at = NOW()
	WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return utils.ErrInternal, errors.New("unable to verify user :: " + err.Error())
	}
	if cmd.RowsAffected() == 0 {
		return utils.ErrNotFound, errors.New("unable to find user ")
	}
	return utils.NoError, nil
}
