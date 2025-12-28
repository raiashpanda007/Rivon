package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raiashpanda007/rivon/internals/utils"
	"golang.org/x/crypto/bcrypt"
)

type TokenServices interface {
	CheckValidRefreshToken(ctx context.Context, refreshToken string, userID string) (bool, *string, utils.ErrorType, error)
	GenerateAccessToken(ctx context.Context, user User, refreshToken string) (*string, utils.ErrorType, error)
	GenerateRefreshToken(ctx context.Context, userID string) (*string, utils.ErrorType, error)
	VerifyAccessToken(ctx context.Context, accessToken string) (*User, utils.ErrorType, error)
	RevokeToken(ctx context.Context, refreshToken string, userID string) (bool, utils.ErrorType, error)
}

type tokenUtils struct {
	AuthSecret string
	Db         *pgxpool.Pool
}

func NewTokenServices(authSecret string, db *pgxpool.Pool) TokenServices {
	return &tokenUtils{AuthSecret: authSecret, Db: db}
}

func GenerateBase64Token() (*string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		slog.Error("Random read error", "error", err)
		return nil, err
	}
	str := base64.RawURLEncoding.EncodeToString(b)
	return &str, nil
}

func (r *tokenUtils) CheckValidRefreshToken(ctx context.Context, refreshToken string, userID string) (bool, *string, utils.ErrorType, error) {
	query := `
		SELECT id, token_hash
		FROM tokens
		WHERE user_id = $1
		  AND revoked = FALSE
		  AND expires_at > NOW()
	`

	rows, err := r.Db.Query(ctx, query, userID)
	if err != nil {
		slog.Error("Database query error (refresh tokens)", "error", err)
		return false, nil, utils.ErrInternal,
			errors.New("failed to query refresh tokens: " + err.Error())
	}
	defer rows.Close()

	found := false

	for rows.Next() {
		found = true

		var tokenID string
		var tokenHash string

		if err := rows.Scan(&tokenID, &tokenHash); err != nil {
			slog.Error("Row scan error", "error", err)
			return false, nil, utils.ErrInternal,
				errors.New("failed to scan token row: " + err.Error())
		}

		if bcrypt.CompareHashAndPassword(
			[]byte(tokenHash),
			[]byte(refreshToken),
		) == nil {
			return true, &tokenID, utils.NoError, nil
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("Rows iteration error", "error", err)
		return false, nil, utils.ErrInternal,
			errors.New("row iteration error: " + err.Error())
	}

	if !found {
		slog.Error("No active refresh token found")
		return false, nil, utils.ErrUnauthorized,
			errors.New("no active refresh token found, please login again")
	}

	return false, nil, utils.ErrUnauthorized,
		errors.New("invalid or expired refresh token")
}

func (r *tokenUtils) GenerateAccessToken(ctx context.Context, user User, refreshToken string) (*string, utils.ErrorType, error) {
	isValid, _, errType, err := r.CheckValidRefreshToken(ctx, refreshToken, user.Id.String())
	if err != nil || !isValid {
		slog.Error("Invalid refresh token check", "error", err, "isValid", isValid)
		return nil, errType, err
	}
	signingToken := []byte(r.AuthSecret)
	claims := jwt.MapClaims{
		"id":       user.Id.String(),
		"name":     user.Name,
		"email":    user.Email,
		"verified": user.Verified,
		"provider": user.Provider,
		"profile":  user.Photo,
		"exp":      time.Now().Add(10 * time.Minute).Unix(),
		"issuedAt": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(signingToken)
	if err != nil {
		slog.Error("Error signing access token", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to generate accessToken :: " + err.Error())
	}
	return &accessToken, utils.NoError, nil
}

func (r *tokenUtils) GenerateRefreshToken(ctx context.Context, userID string) (*string, utils.ErrorType, error) {

	parsedUserId, err := uuid.Parse(userID)

	if err != nil {
		slog.Error("Invalid user ID format", "userID", userID, "error", err)
		return nil, utils.ErrInternal, errors.New("invalid user id")
	}
	query := `INSERT INTO tokens (id , user_id, token_hash, expires_at) VALUES($1, $2, $3, $4)`

	refreshToken, err := GenerateBase64Token()
	if err != nil {
		slog.Error("Error generating base64 token", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to generate refresh token" + err.Error())
	}
	token_hash, err := bcrypt.GenerateFromPassword([]byte(*refreshToken), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Error hashing refresh token", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to generate hash for refresh token :: " + err.Error())
	}
	expiresAt := time.Now().Add(15 * 24 * time.Hour)
	_, err = r.Db.Exec(ctx, query, uuid.New(), parsedUserId, token_hash, expiresAt)
	if err != nil {
		slog.Error("Database error saving refresh token", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to save the token hash in db :: " + err.Error())
	}
	return refreshToken, utils.NoError, nil
}

func (r *tokenUtils) VerifyAccessToken(ctx context.Context, accessToken string) (*User, utils.ErrorType, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(r.AuthSecret), nil
	})

	if err != nil || !token.Valid {
		slog.Error("Invalid access token", "error", err)
		return nil, utils.ErrUnauthorized, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("Invalid token claims type")
		return nil, utils.ErrUnauthorized, errors.New("invalid token claims")
	}

	idStr, ok := claims["id"].(string)
	if !ok || idStr == "" {
		slog.Error("Invalid or missing id claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid id claim")
	}

	name, ok := claims["name"].(string)
	if !ok {
		slog.Error("Invalid or missing name claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid name claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		slog.Error("Invalid or missing email claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid email claim")
	}

	verified, ok := claims["verified"].(bool)
	if !ok {
		slog.Error("Invalid or missing verified claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid verified claim")
	}

	provider, ok := claims["provider"].(string)
	if !ok {
		slog.Error("Invalid or missing provider claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid provider claim")
	}
	photo, ok := claims["profile"].(string)
	if !ok {
		slog.Error("Invalid or missing profile claim")
		return nil, utils.ErrUnauthorized, errors.New("invalid profile claim")
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Invalid user ID in token", "idStr", idStr, "error", err)
		return nil, utils.ErrBadRequest, errors.New("Invalid User id token please login again ... " + err.Error())
	}
	return &User{
		Id:       uid,
		Name:     name,
		Email:    email,
		Verified: verified,
		Provider: AuthProvider(provider),
		Photo:    photo,
	}, utils.NoError, nil

}

func (r *tokenUtils) RevokeToken(ctx context.Context, refreshToken string, userID string) (bool, utils.ErrorType, error) {
	query := `
		UPDATE tokens
		SET revoked = TRUE , updated_at = NOW()
		WHERE id = $1 
		`
	isValid, tokenId, errType, err := r.CheckValidRefreshToken(ctx, refreshToken, userID)
	if err != nil || !isValid {
		slog.Error("Invalid refresh token during revocation", "error", err, "isValid", isValid)
		return false, errType, errors.New("Unable to validate the refreshtoken please provide a valid refresh token")
	}
	cmd, err := r.Db.Exec(ctx, query, tokenId)
	if err != nil {
		slog.Error("Database error revoking token", "error", err)
		return false, utils.ErrInternal, errors.New("Unable to save new token status" + err.Error())
	}
	if cmd.RowsAffected() == 0 {
		slog.Error("Token not found for revocation", "tokenId", tokenId)
		return false, utils.ErrBadRequest, errors.New("Please login first before revoking refresh token ")
	}
	return true, utils.NoError, nil

}
