package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/raiashpanda007/rivon/internals/utils"
	"golang.org/x/crypto/bcrypt"
)

type AccessToken string
type RefreshToken string

type AuthServices interface {
	CredentialSignUp(ctx context.Context, email, name, password string) (*User, AccessToken, RefreshToken, utils.ErrorType, error)
	CredentialSignIn(ctx context.Context, email, password string) (*User, AccessToken, RefreshToken, utils.ErrorType, error)
	CredentialSignOut(ctx context.Context, userId, refreshToken string) (bool, utils.ErrorType, error)
	CredentialRefreshToken(ctx context.Context, refreshToken, userId string) (string, utils.ErrorType, error)
	SendOTP(ctx context.Context, userID, name, email string) (utils.ErrorType, error)
	VerifyOTP(ctx context.Context, userID, otp string) (bool, utils.ErrorType, error)
}

type authUtils struct {
	UserRepo UserRepo
	Token    TokenServices
	OTP      OTPServices
}

func NewAuthServices(userRepo UserRepo, token TokenServices, otp OTPServices) AuthServices {
	return &authUtils{
		UserRepo: userRepo,
		Token:    token,
		OTP:      otp,
	}
}

func (r *authUtils) CredentialSignUp(ctx context.Context, email, name, password string) (*User, AccessToken, RefreshToken, utils.ErrorType, error) {
	cost := bcrypt.DefaultCost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		slog.Error("Error hashing password", "error", err)
		return nil, "", "", utils.ErrInternal, errors.New("Unable to hash your password :: " + err.Error())
	}
	createdUser, errType, err := r.UserRepo.CreateUserCredentials(ctx, email, name, string(hashedPassword))
	if err != nil {
		slog.Error("Error creating user credentials", "error", err)
		return nil, "", "", errType, err
	}
	refreshToken, errType, err := r.Token.GenerateRefreshToken(ctx, createdUser.Id.String())
	if err != nil {
		slog.Error("Error generating refresh token", "error", err)
		return nil, "", "", errType, err
	}
	accessToken, errType, err := r.Token.GenerateAccessToken(ctx, *createdUser, *refreshToken)
	if err != nil {
		slog.Error("Error generating access token", "error", err)
		return nil, "", "", errType, err
	}
	rt := RefreshToken(*refreshToken)
	at := AccessToken(*accessToken)
	return createdUser, at, rt, utils.NoError, nil

}

func (r *authUtils) CredentialSignIn(ctx context.Context, email, password string) (*User, AccessToken, RefreshToken, utils.ErrorType, error) {
	savedUser, userPassword, errType, err := r.UserRepo.GetUserByEmail(ctx, email, ProviderCredentials)

	if err != nil {
		slog.Error("Error getting user by email", "error", err)
		return nil, "", "", errType, err
	}

	verifyPassword := bcrypt.CompareHashAndPassword([]byte(*userPassword), []byte(password))
	if verifyPassword != nil {
		slog.Error("Password verification failed", "error", verifyPassword)
		return nil, "", "", utils.ErrBadRequest, errors.New("wrong password please login with valid password")
	}
	refreshToken, errType, err := r.Token.GenerateRefreshToken(ctx, savedUser.Id.String())
	if err != nil {
		slog.Error("Error generating refresh token", "error", err)
		return nil, "", "", errType, err
	}
	accessToken, errType, err := r.Token.GenerateAccessToken(ctx, *savedUser, *refreshToken)
	if err != nil {
		slog.Error("Error generating access token", "error", err)
		return nil, "", "", errType, err
	}

	rt := RefreshToken(*refreshToken)
	at := AccessToken(*accessToken)
	return savedUser, at, rt, utils.NoError, nil

}

func (r *authUtils) CredentialSignOut(ctx context.Context, userId, refreshToken string) (bool, utils.ErrorType, error) {
	ok, errType, err := r.Token.RevokeToken(ctx, refreshToken, userId)
	if err != nil {
		slog.Error("Error revoking token", "error", err)
		return false, errType, err
	}
	return ok, errType, err
}

func (r *authUtils) CredentialRefreshToken(ctx context.Context, refreshToken, userId string) (string, utils.ErrorType, error) {
	user, _, errType, err := r.UserRepo.GetUserByID(ctx, userId)

	if err != nil {
		slog.Error("Error getting user by ID", "error", err)
		return "", errType, err
	}
	token, errType, err := r.Token.GenerateAccessToken(ctx, *user, refreshToken)

	if err != nil {
		slog.Error("Error generating access token", "error", err)
		return "", errType, nil
	}

	return *token, utils.NoError, nil

}

func (r *authUtils) SendOTP(ctx context.Context, userID, name, email string) (utils.ErrorType, error) {
	otp, errType, err := r.OTP.GenerateOTP(ctx, userID)

	if err != nil {
		slog.Error("Error generating OTP", "error", err)
		return errType, err
	}

	errType, err = r.OTP.SendOTP(ctx, userID, name, *otp, email)

	if err != nil {
		slog.Error("Error sending OTP", "error", err)
		return errType, err
	}

	return utils.NoError, nil
}

func (r *authUtils) VerifyOTP(ctx context.Context, userID, otp string) (bool, utils.ErrorType, error) {

	isValid, errType, err := r.OTP.VerifyOTP(ctx, otp, userID)
	if err != nil {
		slog.Error("Error verifying OTP", "error", err)
		return false, errType, err
	}
	if !isValid {
		return false, utils.NoError, nil // we have catched this error in controllers .
	}
	errType, err = r.UserRepo.UpdateUserVerification(ctx, userID)
	if err != nil {
		slog.Error("Error updating user verification", "error", err)
		return true, errType, err
	}

	return true, utils.NoError, nil
}
