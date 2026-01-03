package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth/gothic"
	"github.com/raiashpanda007/rivon/internals/services"
	"github.com/raiashpanda007/rivon/internals/services/auth"
	"github.com/raiashpanda007/rivon/internals/types"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type AuthController interface {
	CredentialSignIn(res http.ResponseWriter, req *http.Request)
	CredentialSignUp(res http.ResponseWriter, req *http.Request)
	CredentialSignOut(res http.ResponseWriter, req *http.Request)
	CredentialRefresh(res http.ResponseWriter, req *http.Request)
	SendVerifyOTP(res http.ResponseWriter, req *http.Request)
	VerifyOTP(res http.ResponseWriter, req *http.Request)
	OAuthLogin(res http.ResponseWriter, req *http.Request)
	Me(res http.ResponseWriter, req *http.Request)
}

type authController struct {
	services      auth.AuthServices
	cookieSecure  bool
	clientBaseURL string
}

func InitAuthController(pgDb *pgxpool.Pool, otpRedis *redis.Client, jwtSecret string, mailServerURL string, cookieSecure bool, clientBaseUrl string) AuthController {
	authSvc := services.InitAuthServices(pgDb, otpRedis, jwtSecret, mailServerURL)
	return &authController{
		services:      *authSvc,
		cookieSecure:  cookieSecure,
		clientBaseURL: clientBaseUrl,
	}
}

func (r *authController) CredentialSignIn(res http.ResponseWriter, req *http.Request) {
	slog.Info("LOGINING USER ... ")
	var loginCredentials types.LoginType
	err := json.NewDecoder(req.Body).Decode(&loginCredentials)
	if errors.Is(err, io.EOF) {
		slog.Error("Error decoding JSON body (EOF)", "error", err)
		utils.WriteJson(res, http.StatusUnprocessableEntity, utils.GenerateError(utils.ErrUnprocessableData, err))
		return
	}
	if err != nil {
		slog.Error("Error decoding JSON body", "error", err)
		utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, err))
		return
	}
	err = validator.New().Struct(loginCredentials)
	if err != nil {
		slog.Error("Validation error", "error", err)
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteJson(res, http.StatusBadRequest, utils.ValidationError(validationErrors))
		return
	}
	loggedInUser, accessToken, refreshToken, errType, err := r.services.CredentialSignIn(req.Context(), loginCredentials.Email, loginCredentials.Password)
	if err != nil {
		slog.Error("CredentialSignIn service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "access_token",
		Value:    string(accessToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(res, &http.Cookie{
		Name:     "refresh_token",
		Value:    string(refreshToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	utils.WriteJson(res, http.StatusOK, utils.Response[auth.User]{
		Status:  http.StatusOK,
		Data:    *loggedInUser,
		Message: "You successfully logged in ... ",
		Heading: "Request Processed",
	})
}

func (r *authController) CredentialSignUp(res http.ResponseWriter, req *http.Request) {
	slog.Info("REGISTERING USER ... ")
	var signUpCredentials types.RegisterType
	err := json.NewDecoder(req.Body).Decode(&signUpCredentials)
	if errors.Is(err, io.EOF) {
		slog.Error("Error decoding JSON body (EOF)", "error", err)
		utils.WriteJson(res, http.StatusUnprocessableEntity, utils.GenerateError(utils.ErrUnprocessableData, err))
		return
	}
	if err != nil {
		slog.Error("Error decoding JSON body", "error", err)
		utils.WriteJson(res, http.StatusBadRequest, utils.GenerateError(utils.ErrBadRequest, err))
		return
	}
	err = validator.New().Struct(signUpCredentials)
	if err != nil {
		slog.Error("Validation error", "error", err)
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteJson(res, http.StatusBadRequest, utils.ValidationError(validationErrors))
		return
	}
	registedUser, accessToken, refreshToken, errType, err := r.services.CredentialSignUp(req.Context(), signUpCredentials.Email, signUpCredentials.Name, signUpCredentials.Password)
	if err != nil {
		slog.Error("CredentialSignUp service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "access_token",
		Value:    string(accessToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(res, &http.Cookie{
		Name:     "refresh_token",
		Value:    string(refreshToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	utils.WriteJson(res, http.StatusOK, utils.Response[auth.User]{
		Status:  http.StatusOK,
		Data:    *registedUser,
		Message: "You successfully registered in ... ",
		Heading: "Request Processed",
	})
}

func (r *authController) CredentialSignOut(res http.ResponseWriter, req *http.Request) {
	slog.Info("LOGGING OUT USER ... ")
	ClientSideToken, err := req.Cookie("refresh_token")
	if err != nil {
		slog.Error("Error retrieving refresh token cookie", "error", err)
		utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("please provide a valid refresh token")))
		return
	}

	userCred, ok := req.Context().Value("USER").(*auth.User)

	if !ok {
		slog.Error("Failed to retrieve user from context")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrForBidden, errors.New("please log in before login out you smart a**")))
		return
	}

	_, errType, err := r.services.CredentialSignOut(req.Context(), userCred.Id.String(), ClientSideToken.Value)

	if err != nil {
		slog.Error("CredentialSignOut service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.SetCookie(res, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	utils.WriteJson(res, http.StatusAccepted, utils.Response[string]{
		Status:  201,
		Message: "Successful logout see you soon",
		Heading: "Request processed .. ",
		Data:    "",
	})

}

func (r *authController) CredentialRefresh(res http.ResponseWriter, req *http.Request) {
	slog.Info("REFESHING ACCESS TOKEN ... ")
	var refreshCredentials types.RefreshTokenType
	err := json.NewDecoder(req.Body).Decode(&refreshCredentials)
	if err != nil {
		slog.Error("Error decoding JSON body", "error", err)
		utils.WriteJson(res, http.StatusUnprocessableEntity, utils.GenerateError(utils.ErrUnprocessableData, errors.New("Invalid JSON data ")))
		return
	}
	err = validator.New().Struct(refreshCredentials)
	if err != nil {
		slog.Error("Validation error", "error", err)
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteJson(res, http.StatusBadRequest, utils.ValidationError(validationErrors))
		return
	}
	clientSiderefreshToken, err := req.Cookie("refresh_token")
	if err != nil {
		slog.Error("Error retrieving refresh token cookie", "error", err)
		utils.WriteJson(res, http.StatusUnauthorized, utils.GenerateError(utils.ErrUnauthorized, errors.New("Please provide a valid refresh token")))
		return
	}
	accessToken, errType, err := r.services.CredentialRefreshToken(req.Context(), clientSiderefreshToken.Value, refreshCredentials.Id)

	if err != nil {
		slog.Error("CredentialRefreshToken service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60 * 60,
	})

	utils.WriteJson(res, http.StatusAccepted, utils.Response[string]{
		Status:  201,
		Message: "Successful refreshed your access token",
		Heading: "Request processed .. ",
		Data:    "",
	})

}

func (r *authController) SendVerifyOTP(res http.ResponseWriter, req *http.Request) {

	userCred, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		slog.Error("Failed to retrieve user from context")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrForBidden, errors.New("Please login to verify your account ... ")))
		return
	}
	errType, err := r.services.SendOTP(req.Context(), userCred.Id.String(), userCred.Name, userCred.Email)
	if err != nil {
		slog.Error("SendOTP service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}
	utils.WriteJson(res, http.StatusOK, utils.Response[string]{
		Status:  http.StatusOK,
		Data:    "OTP Sent ... ",
		Message: "OTP Has been sended please checkout the email ... ",
		Heading: "OTP SENT",
	})
}

func (r *authController) VerifyOTP(res http.ResponseWriter, req *http.Request) {
	var verifyOTPCredentials types.VerifyOTPCredentials
	err := json.NewDecoder(req.Body).Decode(&verifyOTPCredentials)
	if err != nil {
		slog.Error("Error decoding JSON body", "error", err)
		utils.WriteJson(res, http.StatusUnprocessableEntity, utils.GenerateError(utils.ErrUnprocessableData, err))
		return
	}
	err = validator.New().Struct(verifyOTPCredentials)
	if err != nil {
		slog.Error("Validation error", "error", err)
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteJson(res, http.StatusBadRequest, utils.ValidationError(validationErrors))
		return
	}
	userCred, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		slog.Error("Failed to retrieve user from context")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrForBidden, errors.New("Please login first in order to verify your OTP")))
		return
	}
	isValid, errType, err := r.services.VerifyOTP(req.Context(), userCred.Id.String(), verifyOTPCredentials.OTP)
	if err != nil {
		slog.Error("VerifyOTP service error", "error", err)
		utils.WriteJson(res, utils.ErrorMap[errType].StatusCode, utils.GenerateError(errType, err))
		return
	}

	if !isValid {
		slog.Error("Invalid OTP provided")
		utils.WriteJson(res, http.StatusBadRequest, utils.Response[string]{
			Status:  http.StatusBadRequest,
			Data:    "Invalid OTP",
			Message: "Wrong OTP ...",
			Heading: "Bad Request",
		})
		return
	}
	utils.WriteJson(res, http.StatusAccepted, utils.Response[string]{
		Status:  http.StatusAccepted,
		Data:    "You are verified ... ",
		Message: "Congratulations now you are a verified user ... ",
		Heading: "Request accepeted and processed",
	})
}

func (r *authController) OAuthLogin(res http.ResponseWriter, req *http.Request) {
	provider := auth.AuthProvider(chi.URLParam(req, "provider"))
	ctx := context.WithValue(req.Context(), "provider", provider)
	req = req.WithContext(ctx)
	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		slog.Error("UNABLE TO GET USER DETAILS :: ", err)
		http.Redirect(res, req, r.clientBaseURL+"/error?err="+err.Error(), http.StatusFound)
		return
	}

	email := user.Email
	if email == "" {
		slog.Error("Unable to read user email ")
		http.Redirect(res, req, r.clientBaseURL+"/error?err="+errors.New("From your oauth can't read email").Error(), http.StatusFound)
		return
	}
	name := user.Name
	profilePhoto := user.AvatarURL

	_, accessToken, refreshToken, _, err := r.services.OAuth(req.Context(), email, name, profilePhoto, provider)
	if err != nil {
		slog.Error("OAuth service error", "error", err)

	}

	http.SetCookie(res, &http.Cookie{
		Name:     "access_token",
		Value:    string(accessToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(res, &http.Cookie{
		Name:     "refresh_token",
		Value:    string(refreshToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
	_ = gothic.Logout(res, req)
	http.Redirect(res, req, r.clientBaseURL, http.StatusFound)
}

func (r *authController) Me(res http.ResponseWriter, req *http.Request) {
	userCred, ok := req.Context().Value("USER").(*auth.User)
	if !ok {
		slog.Error("Failed to retrieve user from context")
		utils.WriteJson(res, http.StatusForbidden, utils.GenerateError(utils.ErrForBidden, errors.New("Please login to get your details ")))
		return
	}

	utils.WriteJson(res, http.StatusAccepted, utils.Response[auth.User]{
		Status:  200,
		Heading: "Request Accepted and processed",
		Message: "Your logged in details",
		Data:    *userCred,
	})

}
