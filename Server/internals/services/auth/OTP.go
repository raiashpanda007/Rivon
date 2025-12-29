package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/raiashpanda007/rivon/internals/utils"
)

type OTPServices interface {
	GenerateOTP(ctx context.Context, userID string) (*string, utils.ErrorType, error)
	VerifyOTP(ctx context.Context, otp string, userID string) (bool, utils.ErrorType, error)
	SendOTP(ctx context.Context, userID string, name string, otp string, email string) (utils.ErrorType, error)
}

type otpUtils struct {
	rdb           *redis.Client
	mailServerUrl string
}

func NewOTPServices(rdb *redis.Client, mailServerURL string) OTPServices {
	return &otpUtils{rdb: rdb, mailServerUrl: mailServerURL}
}

func (r *otpUtils) GetValidOTPfromRedis(ctx context.Context, userId string, regen bool) (*string, error) {
	keyString := fmt.Sprintf("auth:otp:%s", userId)
	otpRes, err := r.rdb.Get(ctx, keyString).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		slog.Error("Redis get error", "error", err)
		return nil, errors.New("Unable to get exisitng OTPs :: " + err.Error())
	}
	if regen {
		err = r.rdb.Set(ctx, keyString, otpRes, 5*time.Minute).Err()
		if err != nil {
			slog.Error("Redis set error (regen)", "error", err)
			return nil, errors.New("Unable to save regenerated otp :: " + err.Error())
		}
	}
	return &otpRes, nil
}

func (r *otpUtils) GenerateOTP(ctx context.Context, userID string) (*string, utils.ErrorType, error) {
	existingOTP, err := r.GetValidOTPfromRedis(ctx, userID, true)
	if err != nil {
		slog.Error("Error getting valid OTP from Redis", "error", err)
		return nil, utils.ErrInternal, err
	}
	if existingOTP != nil {
		return existingOTP, utils.NoError, nil
	}

	var otp string

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		slog.Error("Error generating random number", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to generate OTP token :: " + err.Error())
	}
	otp = fmt.Sprintf("%06d", n)
	keyString := fmt.Sprintf("auth:otp:%s", userID)
	err = r.rdb.Set(ctx, keyString, otp, 5*time.Minute).Err()
	if err != nil {
		slog.Error("Redis set error (new OTP)", "error", err)
		return nil, utils.ErrInternal, errors.New("Unable to save the generated otp :: " + err.Error())
	}
	return &otp, utils.NoError, nil
}

func (r *otpUtils) VerifyOTP(ctx context.Context, otp string, userID string) (bool, utils.ErrorType, error) {
	existingOTP, err := r.GetValidOTPfromRedis(ctx, userID, false)
	if err != nil {
		slog.Error("Error getting valid OTP from Redis", "error", err)
		return false, utils.ErrInternal, err
	}
	if existingOTP == nil {
		slog.Error("No existing OTP found for user", "userID", userID)
		return false, utils.ErrNotFound, errors.New("Unable to generated otp for that particular , generate an OTP first")
	}

	if *existingOTP == otp {
		keyString := fmt.Sprintf("auth:otp:%s", userID)
		err = r.rdb.Del(ctx, keyString).Err()
		if err != nil {
			slog.Error("Redis delete error", "error", err)
			return false, utils.ErrInternal, errors.New("Unable to delete the matched key :: " + err.Error())
		}
		return true, utils.NoError, nil
	}
	slog.Error("Invalid OTP provided", "userID", userID)
	return false, utils.ErrUnprocessableData, errors.New("Invalid OTP ")

}
func (r *otpUtils) SendOTP(ctx context.Context, userID string, name string, otp string, email string) (utils.ErrorType, error) {
	var template = fmt.Sprintf(`
		<h2 style="color: #ffffff; margin-top: 0; font-weight: 700;">Verify your email</h2>
		<p>Hi %s,</p>
		<p>Welcome to Rivon 👋 Thanks for signing up. To complete your registration, please verify your email using the OTP below:</p>
		
		<div class="otp-block">
			<div class="otp-text">%s</div>
			<div class="copy-instruction">This code expires in 5 minutes</div>
		</div>

		<p>If you didn't request this, you can safely ignore this email.</p>
		<br>
		<p>Rivon is being built step by step, and I genuinely appreciate you trying it out early. If you face any issues or have feedback, just reply to this email, I read everything.</p>
		<br>
		<p>Ashwin Rai<br>Creator, Rivon</p>
	`, name, otp)
	payload := map[string]string{
		"email":   email,
		"subject": "Verify your email for Rivon",
		"body":    template,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("JSON marshal error", "error", err)
		return utils.ErrInternal, errors.New("Unable jsonify the data :: " + err.Error())
	}

	res, err := http.Post(
		r.mailServerUrl,
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		slog.Error("HTTP post error", "error", err)
		return utils.ErrInternal, errors.New("Unable to send the otp to mailing server :: " + err.Error())
	}
	if res.StatusCode >= 400 {
		slog.Error("Mail server returned error status", "status", res.StatusCode)
		return utils.ErrInternal, errors.New("Unable to send the otp to mailing server :: ")
	}
	defer res.Body.Close()
	return utils.NoError, nil

}
