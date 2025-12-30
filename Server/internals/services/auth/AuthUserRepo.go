package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

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
	Photo    string       `json:"profile"`
	Provider AuthProvider `json:"provider"`
}

type userRepoServices struct {
	db *pgxpool.Pool
}

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string, provider AuthProvider) (*User, *string, utils.ErrorType, error)
	GetUserByID(ctx context.Context, id string) (*User, string, utils.ErrorType, error)
	CreateUserCredentials(ctx context.Context, email, name, passwordHash string) (*User, utils.ErrorType, error)
	CreateUserOAuth(ctx context.Context, email, name, profile string, provider AuthProvider) (*User, utils.ErrorType, error)
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
			slog.Error("User not found by email", "email", email, "error", err)
			return nil, nil, utils.ErrNotFound, errors.New("user not found")
		}
		slog.Error("Database error getting user by email", "error", err)
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
			slog.Error("User not found by ID", "id", id, "error", err)
			return nil, "", utils.ErrNotFound, errors.New("user not found")
		}
		slog.Error("Database error getting user by ID", "error", err)
		return nil, "", utils.ErrInternal, err
	}

	return &user, password, utils.NoError, nil
}

func (r *userRepoServices) CreateUserCredentials(ctx context.Context, email string, name string, passwordHash string) (*User, utils.ErrorType, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("Failed to start transaction", "error", err)
		return nil, utils.ErrInternal, err
	}
	defer tx.Rollback(ctx)

	userID := uuid.New()
	var user User
	userQuery := `
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

	err = tx.QueryRow(
		ctx,
		userQuery,
		userID,
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
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("User already exists", "email", email)
			return nil, utils.ErrConflict, errors.New("user already exists")
		}
		slog.Error("Failed to create user", "error", err)
		return nil, utils.ErrInternal, err
	}

	walletQuery := `
	INSERT INTO wallets (
		id,
		user_id
	)
	VALUES ($1, $2);
	`

	_, err = tx.Exec(
		ctx,
		walletQuery,
		uuid.New(),
		userID,
	)
	if err != nil {
		slog.Error("Failed to create wallet", "user_id", userID, "error", err)
		return nil, utils.ErrInternal, err
	}

	// 3️⃣ Commit transaction
	if err := tx.Commit(ctx); err != nil {
		slog.Error("Transaction commit failed", "error", err)
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
		slog.Error("Invalid user ID format", "id", id, "error", err)
		return false, utils.ErrBadRequest, errors.New("invalid user id")
	}

	cmd, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userId)
	if err != nil {
		slog.Error("Database error deleting user", "error", err)
		return false, utils.ErrInternal, err
	}

	if cmd.RowsAffected() == 0 {
		slog.Error("User not found for deletion", "id", id)
		return false, utils.ErrNotFound, errors.New("user not found")
	}

	return true, utils.NoError, nil
}

func (r *userRepoServices) UpdatePassword(ctx context.Context, id string, newPassword string) (bool, utils.ErrorType, error) {
	userId, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Invalid user ID format", "id", id, "error", err)
		return false, utils.ErrBadRequest, errors.New("please provide a valid user id to update the password ")
	}
	query := `UPDATE users SET password_hash = $1 , updated_at = NOW() WHERE id = $2`

	cmd, err := r.db.Exec(ctx, query, newPassword, userId)
	if err != nil {
		slog.Error("Database error updating password", "error", err)
		return false, utils.ErrInternal, err
	}

	if cmd.RowsAffected() == 0 {
		slog.Error("User not found for password update", "id", id)
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
		slog.Error("Database error updating user verification", "error", err)
		return utils.ErrInternal, errors.New("unable to verify user :: " + err.Error())
	}
	if cmd.RowsAffected() == 0 {
		slog.Error("User not found for verification update", "id", userID)
		return utils.ErrNotFound, errors.New("unable to find user ")
	}
	return utils.NoError, nil
}

func (r *userRepoServices) CreateUserOAuth(ctx context.Context, email, name, profilePhoto string, provider AuthProvider) (*User, utils.ErrorType, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("Failed to start transaction", "error", err)
		return nil, utils.ErrInternal, err
	}
	defer tx.Rollback(ctx)

	var user User
	var isNew bool
	userID := uuid.New()

	query := `
	INSERT INTO users (
		id,
		name,
		email,
		provider,
		verified,
		display_photo
	)
	VALUES ($1, $2, $3, $4, true, $5)
	ON CONFLICT (email, provider)
	DO UPDATE SET
		name = EXCLUDED.name,
		display_photo = EXCLUDED.display_photo
	RETURNING
		id,
		type,
		name,
		email,
		verified,
		provider,
		display_photo,
		(xmax = 0) AS is_new;
	`

	err = tx.QueryRow(
		ctx,
		query,
		userID,
		name,
		email,
		provider,
		profilePhoto,
	).Scan(
		&user.Id,
		&user.Type,
		&user.Name,
		&user.Email,
		&user.Verified,
		&user.Provider,
		&user.Photo,
		&isNew,
	)

	if err != nil {
		slog.Error("OAuth upsert failed", "error", err)
		return nil, utils.ErrInternal, err
	}

	if isNew {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO wallets (id, user_id)
			VALUES ($1, $2);
			`,
			uuid.New(),
			user.Id,
		)
		if err != nil {
			slog.Error("Failed to create wallet for OAuth user", "user_id", user.Id, "error", err)
			return nil, utils.ErrInternal, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Transaction commit failed", "error", err)
		return nil, utils.ErrInternal, err
	}

	return &user, utils.NoError, nil
}
